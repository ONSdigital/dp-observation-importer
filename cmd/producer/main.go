package main

import (
	"time"

	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/go-ns/kafka"
)

func main() {
	var brokers []string
	brokers = append(brokers, "localhost:9092")
	// instanceID := "a4695fee-f0a2-49c4-b136-e3ca8dd40476"
	producer := kafka.NewProducer(brokers, "observation-extracted", int(2000000))
	event1 := event.ObservationExtracted{InstanceID: "7", Row: "5,,sex,male,age,30"}
	sendEvent(producer, event1)
	// event2 := event.ObservationExtracted{InstanceID: "a4695fee-f0a2-49c4-b136-e3ca8dd40476", Row: "6,2,"}
	// sendEvent(producer, event2)
	// event3 := event.ObservationExtracted{InstanceID: instanceID, Row: "5,,sex,female,age,20"}
	// sendEvent(producer, event3)
	time.Sleep(time.Duration(5000 * time.Millisecond))
	producer.Closer() <- true
}

func sendEvent(producer kafka.Producer, extracted event.ObservationExtracted) {
	bytes, error := schema.ObservationExtractedEvent.Marshal(extracted)
	if error != nil {
		panic(error)
	}
	producer.Output() <- bytes
}
