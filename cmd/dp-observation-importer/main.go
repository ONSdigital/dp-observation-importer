package main

import (
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/go-ns/log"
	"os"
	"os/signal"
	"syscall"
	"github.com/ONSdigital/dp-observation-importer/event"
	"time"
	"github.com/ONSdigital/dp-observation-importer/errors"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-observation-importer/dimension"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

func main() {
	log.Namespace = "dp-observation-importer"

	config, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
	log.Debug("loaded config", log.Data{"config": config})

	kafkaBrokers := []string{config.KafkaAddr}
	kafkaConsumer, err := kafka.NewConsumerGroup(
		kafkaBrokers,
		config.ObservationConsumerTopic,
		config.ObservationConsumerGroup,
		kafka.OffsetNewest)

	if err != nil {
		log.Error(err, log.Data{"message": "failed to create kafka consumer"})
		os.Exit(1)
	}

	kafkaErrorProducer := kafka.NewProducer(kafkaBrokers, config.ErrorProducerTopic, 0)

	driver := bolt.NewDriver()
	dbConnection, err := driver.OpenNeo("bolt://localhost:7687")

	if err != nil {
		log.Error(err, log.Data{"message": "failed to create connection to Neo4j"})
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	exit := make(chan struct{})

	go func() {

		<-signals

		close(exit)

		// gracefully dispose resources
		kafkaConsumer.Closer() <- true

		log.Debug("graceful shutdown was successful", nil)
		os.Exit(0)
	}()

	// How long do we wait for a batch to be full before just processing a partially full batch.
	batchWaitTime := time.Millisecond * time.Duration(config.BatchWaitTimeMS)

	// objects to get dimension data - via the import API + cached locally in memory.
	dimensionStore := dimension.NewStore(config.ImportAPIURL)
	dimensionOrderCache := dimension.NewOrderCache(dimensionStore)
	dimensionIDCache := dimension.NewIDCache(dimensionStore)

	// maps from CSV row to observation data.
	observationMapper := observation.NewMapper(dimensionOrderCache)

	// stores observations in the DB.
	observationStore := observation.NewStore(dimensionIDCache, dbConnection)

	// handle a batch of events.
	batchHandler := event.NewBatchHandler(observationMapper, observationStore)

	// when errors occur - we send a message on an error topic.
	errorHandler := errors.NewKafkaHandler(kafkaErrorProducer)

	// Start listening for event messages.
	event.Consume(kafkaConsumer, config.BatchSize, errorHandler, batchHandler, batchWaitTime, exit)
}
