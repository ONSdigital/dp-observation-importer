package main

import (
	"context"
	"os"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/log.go/log"
)

func main() {
	ctx := context.Background()
	var brokers []string
	brokers = append(brokers, "localhost:9092")

	var maxBytes = int(2000000)
	var kafkaVersion = "1.0.2"

	pChannels := kafka.CreateProducerChannels()
	pConfig := &kafka.ProducerConfig{
		MaxMessageBytes: &maxBytes,
		KafkaVersion:    &kafkaVersion,
	}

	producer, err := kafka.NewProducer(ctx, brokers, "observation-extracted", pChannels, pConfig)
	if err != nil {
		log.Event(ctx, "failed to create kafka prodecer", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	event1 := event.ObservationExtracted{InstanceID: "7", Row: "5,,sex,male,age,30"}
	sendEvent(producer, event1)
	event2 := event.ObservationExtracted{InstanceID: "7", Row: "5,,sex,female,age,20"}
	sendEvent(producer, event2)
	time.Sleep(time.Duration(5000 * time.Millisecond))

	producer.Close(nil)
}

func sendEvent(producer *kafka.Producer, extracted event.ObservationExtracted) {
	bytes, error := schema.ObservationExtractedEvent.Marshal(extracted)
	if error != nil {
		panic(error)
	}
	producer.Channels().Output <- bytes
}
