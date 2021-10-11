package service

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/dimension"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/log.go/v2/log"
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

	log.Info(ctx, "starting observation importer")

	if cfg.GraphDriverChoice == "neo4j" && !cfg.EnableGetGraphDimensionID {
		errStr := "invalid flag combination, getGraphDimensionID must not be false for Neo4j"
		err := errors.New(errStr)
		log.Fatal(ctx, errStr, err)
		return err
	}

	// Get syncConsumerGroup Kafka Consumer
	syncConsumerGroup, err := serviceList.GetConsumer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "could not obtain consumer group", err)
		return err
	}

	// Get observations inserted Kafka Producer
	observationsImportedProducer, err := serviceList.GetProducer(
		ctx,
		cfg.KafkaConfig.Brokers,
		cfg.KafkaConfig.ResultProducerTopic,
		initialise.ObservationsImported,
		cfg,
	)
	if err != nil {
		log.Fatal(ctx, "could not obtain observations inserted producer", err)
		return err
	}

	// Get observations inserted error Kafka Producer
	observationsImportedErrProducer, err := serviceList.GetProducer(
		ctx,
		cfg.KafkaConfig.Brokers,
		cfg.KafkaConfig.ErrorProducerTopic,
		initialise.ObservationsImportedErr,
		cfg,
	)
	if err != nil {
		log.Fatal(ctx, "could not obtain observations inserted error producer", err)
		return err
	}

	// Get Error reporter
	errorReporter, err := serviceList.GetImportErrorReporter(observationsImportedErrProducer, log.Namespace)
	if err != nil {
		log.Fatal(ctx, "error while attempting to create error reporter client", err)
		return err
	}

	// Get graphdb connection for observation store
	graphDB, err := serviceList.GetGraphDB(ctx)
	if err != nil {
		log.Fatal(ctx, "failed to instantiate graph observation store", err)
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
		log.Fatal(ctx, "could not instantiate healthcheck", err)
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
		log.Info(ctx, "starting http server", log.Data{"bind_addr": cfg.BindAddr})
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
	observationStore := observation.NewStore(dimensionIDCache, graphDB, errorReporter, cfg.EnableGetGraphDimensionID)

	// write import results to kafka topic.
	resultWriter := observation.NewResultWriter(observationsImportedProducer)

	// handle a batch of events.
	batchHandler := event.NewBatchHandler(observationMapper, observationStore, resultWriter, errorReporter)

	eventConsumer := event.NewConsumer()

	// Start listening for event messages.
	eventConsumer.Consume(syncConsumerGroup, cfg.KafkaConfig.BatchSize, batchHandler, cfg.KafkaConfig.BatchWaitTime, errorChannel)

	syncConsumerGroup.Channels().LogErrors(ctx, "error received from kafka consumer, topic: "+cfg.KafkaConfig.ObservationConsumerTopic)
	observationsImportedProducer.Channels().LogErrors(ctx, "error received from kafka producer, topic: "+cfg.KafkaConfig.ResultProducerTopic)
	observationsImportedErrProducer.Channels().LogErrors(ctx, "error received from kafka producer, topic: "+cfg.KafkaConfig.ErrorProducerTopic)

	go func() {
		apiError := <-errorChannel
		log.Error(ctx, "error received from http server", apiError)
	}()

	// block until a fatal error occurs
	<-signals
	log.Info(ctx, "os signal received")

	log.Info(ctx, fmt.Sprintf("shutdown with timeout: %s", cfg.GracefulShutdownTimeout))
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
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// If kafka consumer exists, stop listening to it. (Will close later)
		if serviceList.Consumer {
			log.Info(shutdownContext, "stopping kafka consumer listener")
			syncConsumerGroup.StopListeningToConsumer(shutdownContext)
			log.Info(shutdownContext, "stopped kafka consumer listener")
		}

		// If observation imported kafka producer exists, close it
		if serviceList.ObservationsImportedProducer {
			log.Info(shutdownContext, "closing observation imported kafka producer")
			observationsImportedProducer.Close(shutdownContext)
			log.Info(shutdownContext, "closed observation imported kafka producer")
		}

		// If observation imported error kafka producer exists, close it
		if serviceList.ObservationsImportedErrProducer {
			log.Info(shutdownContext, "closing observation imported error kafka producer")
			observationsImportedErrProducer.Close(shutdownContext)
			log.Info(shutdownContext, "closed observation imported error kafka producer")
		}

		// Attempting to close event consumer
		log.Info(shutdownContext, "closing event wrapper")
		eventConsumer.Close(shutdownContext)
		log.Info(shutdownContext, "closed event wrapper")

		// If kafka consumer exists, close it.
		if serviceList.Consumer {
			log.Info(shutdownContext, "closing kafka consumer")
			syncConsumerGroup.Close(shutdownContext)
			log.Info(shutdownContext, "closed kafka consumer")
		}

		if serviceList.Graph {
			log.Info(ctx, "closing graph db")
			err = graphDB.Close(ctx)
			if err != nil {
				log.Error(ctx, "failed to close graph db", err)
				hasShutdownError = true
			}
			err = graphErrorConsumer.Close(ctx)
			if err != nil {
				log.Error(ctx, "failed to close graph db error consumer", err)
				hasShutdownError = true
			}
			log.Info(ctx, "closed graph db")
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-shutdownContext.Done()

	if !gracefulShutdown {
		err = errors.New("failed to shutdown gracefully")
		log.Error(shutdownContext, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(shutdownContext, "graceful shutdown was successful")

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
		log.Error(ctx, "error adding check for kafka consumer", err)
	}

	if err = hc.AddCheck("Kafka Import Producer", observationsImportedProducer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for kafka import producer", err)
	}

	if err = hc.AddCheck("Kafka Error Producer", observationsImportedErrProducer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for kafka error producer", err)
	}

	if err = hc.AddCheck("Dataset API", datasetClient.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for dataset client", err)
	}

	if err = hc.AddCheck("Graph DB", graphDB.Driver.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for graph db", err)
	}

	if hasErrors {
		return errors.New("error registering checkers for health check")
	}
	return nil
}
