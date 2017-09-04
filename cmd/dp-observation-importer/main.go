package main

import (
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/dimension"
	"github.com/ONSdigital/dp-observation-importer/errors"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
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
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	// Avoid logging the neo4j URL as it may contain a password
	// Avoid logging the neo4j URL as it may contain a password
	log.Debug("loaded config", log.Data{
		"topics":                     []string{config.ObservationConsumerTopic, config.ErrorProducerTopic, config.ResultProducerTopic},
		"brokers":                    config.KafkaAddr,
		"bind_addr":                  config.BindAddr,
		"import_api_url":             config.ImportAPIURL,
		"observation_consumer_group": config.ObservationConsumerGroup,
		"cache_ttl":                  config.CacheTTL,
		"batch_size":                 config.BatchSize,
		"batch_time":                 config.BatchWaitTime})

	kafkaBrokers := config.KafkaAddr

	kafkaConsumer, err := kafka.NewConsumerGroup(
		kafkaBrokers,
		config.ObservationConsumerTopic,
		config.ObservationConsumerGroup,
		kafka.OffsetNewest)
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	kafkaErrorProducer, err := kafka.NewProducer(kafkaBrokers, config.ErrorProducerTopic, 0)
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	kafkaResultProducer, err := kafka.NewProducer(kafkaBrokers, config.ResultProducerTopic, 0)
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	dbConnection, err := bolt.NewDriver().OpenNeo(config.DatabaseAddress)

	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// a channel used to signal a graceful exit.
	// it can be listened to by multiple consumers as the closing of the channel is the signal.
	error := make(chan error)

	// when errors occur - we send a message on an error topic.
	errorHandler := errors.NewKafkaHandler(kafkaErrorProducer)

	// objects to get dimension data - via the import API + cached locally in memory.
	httpClient := http.Client{Timeout: time.Second * 15}
	dimensionStore := dimension.NewStore(config.ImportAPIURL, &httpClient)
	dimensionOrderCache := dimension.NewOrderCache(dimensionStore, config.CacheTTL)
	dimensionIDCache := dimension.NewIDCache(dimensionStore, config.CacheTTL)

	// maps from CSV row to observation data.
	observationMapper := observation.NewMapper(dimensionOrderCache)

	// stores observations in the DB.
	observationStore := observation.NewStore(dimensionIDCache, dbConnection, errorHandler)

	// write import results to kafka topic.
	resultWriter := observation.NewResultWriter(kafkaResultProducer)

	// handle a batch of events.
	batchHandler := event.NewBatchHandler(observationMapper, observationStore, resultWriter, errorHandler)

	eventConsumer := event.NewConsumer()

	// Start listening for event messages.
	eventConsumer.Consume(kafkaConsumer, config.BatchSize, batchHandler, config.BatchWaitTime, error)

	//shutdownGracefully := func(err error) {
	//
	//	if err != nil {
	//		log.Error(err, nil)
	//	}
	//
	//	// gracefully dispose resources
	//	eventConsumer.Close()
	//	kafkaConsumer.Close()
	//	kafkaErrorProducer.Close()
	//	kafkaResultProducer.Close()
	//
	//	log.Debug("graceful shutdown was successful", nil)
	//	os.Exit(0)
	//}
	//
	//for {
	//	select {
	//
	//	case err := <-kafkaConsumer.Errors():
	//		shutdownGracefully(err)
	//	case err := <-kafkaResultProducer.Errors():
	//		shutdownGracefully(err)
	//	case err := <-kafkaResultProducer.Errors():
	//		shutdownGracefully(err)
	//	case err := <-error:
	//		shutdownGracefully(err)
	//	case <-signals:
	//		shutdownGracefully(fmt.Errorf("os signal receieved"))
	//	}
	//}
}
