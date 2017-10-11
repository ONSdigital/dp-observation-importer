package main

import (
	"context"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/dimension"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/ONSdigital/go-ns/handlers/healthcheck"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Namespace = "dp-observation-importer"

	config, err := config.Get()
	checkForError(err)

	// Avoid logging the neo4j URL as it may contain a password
	log.Debug("loaded config", log.Data{
		"topics":                     []string{config.ObservationConsumerTopic, config.ErrorProducerTopic, config.ResultProducerTopic},
		"brokers":                    config.KafkaAddr,
		"bind_addr":                  config.BindAddr,
		"dataset_api_url":            config.DatasetAPIURL,
		"observation_consumer_group": config.ObservationConsumerGroup,
		"cache_ttl":                  config.CacheTTL,
		"batch_size":                 config.BatchSize,
		"batch_time":                 config.BatchWaitTime,
		"graceful_shutdown_timeout":  config.GracefulShutdownTimeout,
	})

	// a channel used to signal a graceful exit is required.
	errorChannel := make(chan error)

	router := mux.NewRouter()
	router.Path("/healthcheck").HandlerFunc(healthcheck.Handler)
	httpServer := server.New(config.BindAddr, router)

	// Disable auto handling of os signals by the HTTP server. This is handled
	// in the service so we can gracefully shutdown resources other than just
	// the HTTP server.
	httpServer.HandleOSSignals = false

	go func() {
		log.Debug("starting http server", log.Data{"bind_addr": config.BindAddr})
		if err := httpServer.ListenAndServe(); err != nil {
			errorChannel <- err
		}
	}()

	kafkaConsumer, err := kafka.NewConsumerGroup(
		config.KafkaAddr,
		config.ObservationConsumerTopic,
		config.ObservationConsumerGroup,
		kafka.OffsetNewest)
	checkForError(err)

	kafkaErrorProducer, err := kafka.NewProducer(config.KafkaAddr, config.ErrorProducerTopic, 0)
	checkForError(err)

	kafkaResultProducer, err := kafka.NewProducer(config.KafkaAddr, config.ResultProducerTopic, 0)
	checkForError(err)

	dbConnection, err := bolt.NewDriver().OpenNeo(config.DatabaseAddress)
	checkForError(err)

	// when errors occur - we send a message on an error topic.
	errorReporter, err := reporter.NewImportErrorReporter(kafkaErrorProducer, log.Namespace)
	checkForError(err)

	// objects to get dimension data - via the dataset API + cached locally in memory.
	httpClient := http.Client{Timeout: time.Second * 15}
	dimensionStore := dimension.NewStore(config.DatasetAPIURL, &httpClient)
	dimensionOrderCache := dimension.NewOrderCache(dimensionStore, config.CacheTTL)
	dimensionIDCache := dimension.NewIDCache(dimensionStore, config.CacheTTL)

	// maps from CSV row to observation data.
	observationMapper := observation.NewMapper(dimensionOrderCache)

	// stores observations in the DB.
	observationStore := observation.NewStore(dimensionIDCache, dbConnection, errorReporter)

	// write import results to kafka topic.
	resultWriter := observation.NewResultWriter(kafkaResultProducer)

	// handle a batch of events.
	batchHandler := event.NewBatchHandler(observationMapper, observationStore, resultWriter, errorReporter)

	eventConsumer := event.NewConsumer()

	// Start listening for event messages.
	eventConsumer.Consume(kafkaConsumer, config.BatchSize, batchHandler, config.BatchWaitTime, errorChannel)

	shutdownGracefully := func() {

		ctx, cancel := context.WithTimeout(context.Background(), config.GracefulShutdownTimeout)

		// gracefully dispose resources
		eventConsumer.Close(ctx)
		kafkaConsumer.Close(ctx)
		kafkaErrorProducer.Close(ctx)
		kafkaResultProducer.Close(ctx)
		httpServer.Shutdown(ctx)

		// cancel the timer in the shutdown context.
		cancel()

		log.Debug("graceful shutdown was successful", nil)
		os.Exit(0)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case err := <-kafkaConsumer.Errors():
			log.ErrorC("kafka consumer", err, nil)
			shutdownGracefully()
		case err := <-kafkaResultProducer.Errors():
			log.ErrorC("kafka result producer", err, nil)
			shutdownGracefully()
		case err := <-kafkaErrorProducer.Errors():
			log.ErrorC("kafka error producer", err, nil)
			shutdownGracefully()
		case err := <-errorChannel:
			log.ErrorC("error channel", err, nil)
			shutdownGracefully()
		case <-signals:
			log.Debug("os signal received", nil)
			shutdownGracefully()
		}
	}
}
func checkForError(err error) {
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
}
