package main

import (
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"time"
)

func main() {
	var brokers []string
	brokers = append(brokers, "localhost:9092")

	producer := kafka.NewProducer(brokers, "observation-extracted", int(2000000))

	event1 := event.ObservationExtracted{InstanceID: "7", Row: "5,,sex,male,age,30"}
	sendEvent(producer, event1)
	//time.Sleep(time.Duration(5000 * time.Millisecond))
	event2 := event.ObservationExtracted{InstanceID: "7", Row: "5,,sex,female,age,20"}
	sendEvent(producer, event2)
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
