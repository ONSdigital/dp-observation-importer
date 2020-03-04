package main

import (
	"context"
	"time"

	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/schema"
)

func main() {
	ctx := context.Background()
	var brokers []string
	brokers = append(brokers, "localhost:9092")

	pChannels := kafka.CreateProducerChannels()

	producer, _ := kafka.NewProducer(ctx, brokers, "observation-extracted", int(2000000), pChannels)

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
