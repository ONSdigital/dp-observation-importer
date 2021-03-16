package service

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/dimension"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func Run(ctx context.Context, cfg *config.Config, serviceList initialise.ExternalServiceList, signals chan os.Signal, BuildTime string, GitCommit string, Version string) error {

	log.Event(ctx, "starting observation importer", log.INFO)

	// Get syncConsumerGroup Kafka Consumer
	syncConsumerGroup, err := serviceList.GetConsumer(ctx, cfg)
	if err != nil {
		log.Event(ctx, "could not obtain consumer group", log.FATAL, log.Error(err))
		return err
	}

	// Get observations inserted Kafka Producer
	observationsImportedProducer, err := serviceList.GetProducer(
		ctx,
		cfg.Brokers,
		cfg.ResultProducerTopic,
		initialise.ObservationsImported,
		cfg,
	)
	if err != nil {
		log.Event(ctx, "could not obtain observations inserted producer", log.FATAL, log.Error(err))
		return err
	}

	// Get observations inserted error Kafka Producer
	observationsImportedErrProducer, err := serviceList.GetProducer(
		ctx,
		cfg.Brokers,
		cfg.ErrorProducerTopic,
		initialise.ObservationsImportedErr,
		cfg,
	)
	if err != nil {
		log.Event(ctx, "could not obtain observations inserted error producer", log.FATAL, log.Error(err))
		return err
	}

	// Get Error reporter
	errorReporter, err := serviceList.GetImportErrorReporter(observationsImportedErrProducer, log.Namespace)
	if err != nil {
		log.Event(ctx, "error while attempting to create error reporter client", log.FATAL, log.Error(err))
		return err
	}

	// Get graphdb connection for observation store
	graphDB, err := serviceList.GetGraphDB(ctx)
	if err != nil {
		log.Event(ctx, "failed to instantiate graph observation store", log.FATAL, log.Error(err))
		return err
	}

	var graphErrorConsumer *graph.ErrorConsumer
	if serviceList.Graph {
		graphErrorConsumer = graph.NewLoggingErrorConsumer(ctx, graphDB.Errors)
	}

	datasetClient := dataset.NewAPIClient(cfg.DatasetAPIURL)

	// Get HealthCheck
	hc, err := serviceList.GetHealthCheck(cfg, BuildTime, GitCommit, Version)
	if err != nil {
		log.Event(ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		return err
	}

	// Add dataset API and graph checks
	if err := registerCheckers(ctx, &hc, syncConsumerGroup, observationsImportedProducer, observationsImportedErrProducer, *datasetClient, graphDB); err != nil {
		return err
	}

	router := mux.NewRouter()
	router.Path("/health").HandlerFunc(hc.Handler)
	httpServer := dphttp.NewServer(cfg.BindAddr, router)

	// Disable auto handling of os signals by the HTTP server. This is handled
	// in the service so we can gracefully shutdown resources other than just
	// the HTTP server.
	httpServer.HandleOSSignals = false

	// a channel to signal a server error
	errorChannel := make(chan error)

	go func() {
		log.Event(ctx, "starting http server", log.INFO, log.Data{"bind_addr": cfg.BindAddr})
		if err = httpServer.ListenAndServe(); err != nil {
			errorChannel <- err
		}
	}()

	hc.Start(ctx)

	// objects to get dimension data - via the dataset API + cached locally in memory.
	dimensionStore := dimension.NewStore(cfg.ServiceAuthToken, cfg.DatasetAPIURL, datasetClient)
	dimensionOrderCache := dimension.NewOrderCache(dimensionStore, cfg.CacheTTL)
	dimensionIDCache := dimension.NewIDCache(dimensionStore, cfg.CacheTTL)

	// maps from CSV row to observation data.
	observationMapper := observation.NewMapper(dimensionOrderCache)

	// stores observations in the DB.
	observationStore := observation.NewStore(dimensionIDCache, graphDB, errorReporter, true)

	// write import results to kafka topic.
	resultWriter := observation.NewResultWriter(observationsImportedProducer)

	// handle a batch of events.
	batchHandler := event.NewBatchHandler(observationMapper, observationStore, resultWriter, errorReporter)

	eventConsumer := event.NewConsumer()

	// Start listening for event messages.
	eventConsumer.Consume(syncConsumerGroup, cfg.BatchSize, batchHandler, cfg.BatchWaitTime, errorChannel)

	syncConsumerGroup.Channels().LogErrors(ctx, "error received from kafka consumer, topic: "+cfg.ObservationConsumerTopic)
	observationsImportedProducer.Channels().LogErrors(ctx, "error received from kafka producer, topic: "+cfg.ResultProducerTopic)
	observationsImportedErrProducer.Channels().LogErrors(ctx, "error received from kafka producer, topic: "+cfg.ErrorProducerTopic)

	go func() {
		select {
		case apiError := <-errorChannel:
			log.Event(ctx, "error received from http server", log.ERROR, log.Error(apiError))
		}
	}()

	// block until a fatal error occurs
	select {
	case <-signals:
		log.Event(ctx, "os signal received", log.INFO)
	}

	log.Event(ctx, fmt.Sprintf("shutdown with timeout: %s", cfg.GracefulShutdownTimeout), log.INFO)
	shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	// track shutdown gracefully closes app
	var gracefulShutdown bool

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		if serviceList.HealthCheck {
			hc.Stop()
		}

		var hasShutdownError bool
		err := httpServer.Shutdown(shutdownContext)
		if err != nil {
			log.Event(ctx, "failed to shutdown http server", log.ERROR, log.Error(err))
			hasShutdownError = true
		}

		// If kafka consumer exists, stop listening to it. (Will close later)
		if serviceList.Consumer {
			log.Event(shutdownContext, "stopping kafka consumer listener", log.INFO)
			syncConsumerGroup.StopListeningToConsumer(shutdownContext)
			log.Event(shutdownContext, "stopped kafka consumer listener", log.INFO)
		}

		// If observation imported kafka producer exists, close it
		if serviceList.ObservationsImportedProducer {
			log.Event(shutdownContext, "closing observation imported kafka producer", log.INFO)
			observationsImportedProducer.Close(shutdownContext)
			log.Event(shutdownContext, "closed observation imported kafka producer", log.INFO)
		}

		// If observation imported error kafka producer exists, close it
		if serviceList.ObservationsImportedErrProducer {
			log.Event(shutdownContext, "closing observation imported error kafka producer", log.INFO)
			observationsImportedErrProducer.Close(shutdownContext)
			log.Event(shutdownContext, "closed observation imported error kafka producer", log.INFO)
		}

		// Attempting to close event consumer
		log.Event(shutdownContext, "closing event wrapper", log.INFO)
		eventConsumer.Close(shutdownContext)
		log.Event(shutdownContext, "closed event wrapper", log.INFO)

		// If kafka consumer exists, close it.
		if serviceList.Consumer {
			log.Event(shutdownContext, "closing kafka consumer", log.INFO)
			syncConsumerGroup.Close(shutdownContext)
			log.Event(shutdownContext, "closed kafka consumer", log.INFO)
		}

		if serviceList.Graph {
			log.Event(ctx, "closing graph db", log.INFO)
			err = graphDB.Close(ctx)
			if err != nil {
				log.Event(ctx, "failed to close graph db", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
			err = graphErrorConsumer.Close(ctx)
			if err != nil {
				log.Event(ctx, "failed to close graph db error consumer", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
			log.Event(ctx, "closed graph db", log.INFO)
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-shutdownContext.Done()

	if !gracefulShutdown {
		err = errors.New("failed to shutdown gracefully")
		log.Event(shutdownContext, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(shutdownContext, "graceful shutdown was successful", log.INFO)

	return nil
}

// registerCheckers adds the checkers for the provided clients to the healthcheck object
func registerCheckers(ctx context.Context, hc *healthcheck.HealthCheck,
	kafkaConsumer kafka.IConsumerGroup,
	observationsImportedProducer kafka.IProducer,
	observationsImportedErrProducer kafka.IProducer,
	datasetClient dataset.Client,
	graphDB *graph.DB) (err error) {

	hasErrors := false

	if err = hc.AddCheck("Kafka Consumer", kafkaConsumer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for kafka consumer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Kafka Import Producer", observationsImportedProducer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for kafka import producer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Kafka Error Producer", observationsImportedErrProducer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for kafka error producer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Dataset API", datasetClient.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for dataset client", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Graph DB", graphDB.Driver.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for graph db", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("error registering checkers for health check")
	}
	return nil
}
