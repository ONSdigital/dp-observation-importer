package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-graph/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/dimension"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/go-ns/server"
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

func main() {
	log.Namespace = "dp-observation-importer"
	ctx := context.Background()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	log.Event(ctx, "Starting observation importer", log.INFO)

	cfg, err := config.Get()
	exitIfFatal(ctx, "failed to retrieve configuration", err)

	// Sensitive fields are omitted from config.String()
	log.Event(ctx, "loaded config", log.INFO, log.Data{"config": cfg})

	// Attempt to parse envMax from config. Exit on failure.
	envMax, err := strconv.ParseInt(cfg.KafkaMaxBytes, 10, 32)
	exitIfFatal(ctx, "encountered error parsing kafka max bytes", err)

	// External services and their initialization state
	var serviceList initialise.ExternalServiceList

	// Get syncConsumerGroup Kafka Consumer
	syncConsumerGroup, err := serviceList.GetConsumer(ctx, cfg)
	exitIfFatal(ctx, "could not obtain consumer group", err)

	// Get observations inserted Kafka Producer
	observationsImportedProducer, err := serviceList.GetProducer(
		ctx,
		cfg.Brokers,
		cfg.ResultProducerTopic,
		initialise.ObservationsImported,
		int(envMax),
	)
	exitIfFatal(ctx, "could not obtain observations inserted producer", err)

	// Get observations inserted error Kafka Producer
	observationsImportedErrProducer, err := serviceList.GetProducer(
		ctx,
		cfg.Brokers,
		cfg.ErrorProducerTopic,
		initialise.ObservationsImportedErr,
		int(envMax),
	)
	exitIfFatal(ctx, "could not obtain observations inserted error producer", err)

	// Get graphdb connection for observation store
	graphDB, err := serviceList.GetGraphDB(ctx)
	logIfError(ctx, "failed to instantiate neo4j observation store", err)

	datasetClient := dataset.NewAPIClient(cfg.DatasetAPIURL)

	// Get HealthCheck
	hc, err := serviceList.GetHealthCheck(cfg, BuildTime, GitCommit, Version)
	exitIfFatal(ctx, "could not instantiate healthcheck", err)

	// Add dataset API and graph checks
	if err := registerCheckers(ctx, &hc, syncConsumerGroup, observationsImportedProducer, observationsImportedErrProducer, *datasetClient, graphDB); err != nil {
		os.Exit(1)
	}

	router := mux.NewRouter()
	router.Path("/health").HandlerFunc(hc.Handler)
	httpServer := server.New(cfg.BindAddr, router)

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

	// Get Error reporter
	errorReporter, err := serviceList.GetImportErrorReporter(observationsImportedErrProducer, log.Namespace)
	logIfError(ctx, "error while attempting to create error reporter client", err)

	// objects to get dimension data - via the dataset API + cached locally in memory.
	dimensionStore := dimension.NewStore(cfg.ServiceAuthToken, cfg.DatasetAPIURL, datasetClient)
	dimensionOrderCache := dimension.NewOrderCache(dimensionStore, cfg.CacheTTL)
	dimensionIDCache := dimension.NewIDCache(dimensionStore, cfg.CacheTTL)

	// maps from CSV row to observation data.
	observationMapper := observation.NewMapper(dimensionOrderCache)

	// stores observations in the DB.
	observationStore := observation.NewStore(dimensionIDCache, graphDB, errorReporter)

	// write import results to kafka topic.
	resultWriter := observation.NewResultWriter(observationsImportedProducer)

	// handle a batch of events.
	batchHandler := event.NewBatchHandler(observationMapper, observationStore, resultWriter, errorReporter)

	eventConsumer := event.NewConsumer()

	// Start listening for event messages.
	eventConsumer.Consume(ctx, syncConsumerGroup, cfg.BatchSize, batchHandler, cfg.BatchWaitTime, errorChannel)

	// block until a fatal error occurs
	select {
	case <-signals:
		log.Event(ctx, "os signal received", log.INFO)
	case err = <-errorChannel:
		log.Event(ctx, "error received from http server", log.ERROR, log.Error(err))
	}

	log.Event(ctx, fmt.Sprintf("Shutdown with timeout: %s", cfg.GracefulShutdownTimeout), log.INFO)
	shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		hc.Stop()

		err = httpServer.Shutdown(shutdownContext)
		logIfError(ctx, "failed to shutdown http server", err)

		// If kafka consumer exists, stop listening to it. (Will close later)
		if serviceList.Consumer {
			log.Event(shutdownContext, "stopping kafka consumer listener", log.INFO)
			syncConsumerGroup.StopListeningToConsumer(shutdownContext)
			log.Event(shutdownContext, "stopped kafka consumer listener", log.INFO)
		}

		if serviceList.Graph {
			err = graphDB.Close(ctx)
			logIfError(ctx, "failed to close graph db", err)
		}

		// If observation imported kafka producer exists, close it
		if serviceList.ObservationsImportedProducer {
			log.Event(shutdownContext, "closing observation imported kafka producer", log.INFO, log.Data{"producer": "DimensionExtracted"})
			observationsImportedProducer.Close(shutdownContext)
			log.Event(shutdownContext, "closed observation imported kafka producer", log.INFO, log.Data{"producer": "DimensionExtracted"})
		}

		// If observation imported error kafka producer exists, close it
		if serviceList.ObservationsImportedErrProducer {
			log.Event(shutdownContext, "closing observation imported error kafka producer", log.INFO, log.Data{"producer": "DimensionExtracted"})
			observationsImportedErrProducer.Close(shutdownContext)
			log.Event(shutdownContext, "closed observation imported error kafka producer", log.INFO, log.Data{"producer": "DimensionExtracted"})
		}

		// If kafka consumer exists, close it.
		if serviceList.Consumer {
			log.Event(shutdownContext, "closing kafka consumer", log.INFO, log.Data{"consumer": "SyncConsumerGroup"})
			syncConsumerGroup.Close(shutdownContext)
			log.Event(shutdownContext, "closed kafka consumer", log.INFO, log.Data{"consumer": "SyncConsumerGroup"})
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	os.Exit(0)
}

// registerCheckers adds the checkers for the provided clients to the healthcheck object
func registerCheckers(ctx context.Context, hc *healthcheck.HealthCheck,
	kafkaConsumer *kafka.ConsumerGroup,
	observationsImportedProducer *kafka.Producer,
	observationsImportedErrProducer *kafka.Producer,
	datasetClient dataset.Client,
	graphDB *graph.DB) (err error) {

	if err = hc.AddCheck("Kafka Consumer", kafkaConsumer.Checker); err != nil {
		log.Event(ctx, "error adding check for kafka consumer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Kafka Producer", observationsImportedProducer.Checker); err != nil {
		log.Event(ctx, "error adding check for kafka producer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Kafka Error Producer", observationsImportedErrProducer.Checker); err != nil {
		log.Event(ctx, "error adding check for kafka error producer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("dataset API", datasetClient.Checker); err != nil {
		log.Event(ctx, "error adding check dataset client", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("graph db", graphDB.Driver.Checker); err != nil {
		log.Event(ctx, "error adding check dataset client", log.ERROR, log.Error(err))
	}

	return
}

func exitIfFatal(ctx context.Context, message string, err error) {
	if err != nil {
		log.Event(ctx, message, log.FATAL, log.Error(err))
		os.Exit(1)
	}
}

func logIfError(ctx context.Context, message string, err error) {
	if err != nil {
		log.Event(ctx, message, log.ERROR, log.Error(err))
	}
}
