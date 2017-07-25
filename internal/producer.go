package main

import (
	"time"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/schema"
)


func main() {
	var brokers []string
	brokers = append(brokers, "localhost:9092")

	producer := kafka.NewProducer(brokers, "observation-extracted", int(2000000))
	//inputFileAvailableProducer := kafka.NewProducer(brokers, "input-test-file", int(2000000))

	//event1 := event.ObservationExtracted{InstanceID:"7", Row:"5,,sex,male,age,30"}
	//sendEvent(producer, event1)
	event2 := event.ObservationExtracted{InstanceID:"7", Row:"5,,sex,female,age,20"}
	sendEvent(producer, event2)
	time.Sleep(time.Duration(1000 * time.Millisecond))
	producer.Closer() <- true
}

func sendEvent(producer kafka.Producer, extracted event.ObservationExtracted) {
	bytes, error := schema.ObservationExtractedEvent.Marshal(extracted)
	if error != nil {
		panic(error)
	}
	producer.Output() <- bytes
}