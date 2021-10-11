package main

import (
	"context"
	"os"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

func main() {
	ctx := context.Background()
	var brokers []string
	brokers = append(brokers, "localhost:9092")

	var cfg *config.Config

	cfg, err := config.Get(ctx)
	if err != nil {
		log.Fatal(ctx, "failed to retrieve configuration", err)
		os.Exit(1)
	}

	pChannels := kafka.CreateProducerChannels()
	pConfig := &kafka.ProducerConfig{
		MaxMessageBytes: &cfg.KafkaConfig.MaxBytes,
		KafkaVersion:    &cfg.KafkaConfig.Version,
	}

	producer, err := kafka.NewProducer(ctx, brokers, "observation-extracted", pChannels, pConfig)
	if err != nil {
		log.Fatal(ctx, "failed to create kafka producer", err)
		os.Exit(1)
	}

	event1 := event.ObservationExtracted{InstanceID: "7", Row: "5,,sex,male,age,30"}
	sendEvent(producer, event1)
	event2 := event.ObservationExtracted{InstanceID: "7", Row: "5,,sex,female,age,20"}
	sendEvent(producer, event2)
	time.Sleep(time.Duration(5000 * time.Millisecond))

	producer.Close(context.TODO())
}

func sendEvent(producer *kafka.Producer, extracted event.ObservationExtracted) {
	bytes, error := schema.ObservationExtractedEvent.Marshal(extracted)
	if error != nil {
		panic(error)
	}
	producer.Channels().Output <- bytes
}
