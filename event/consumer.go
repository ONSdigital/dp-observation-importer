package event

import (
	"github.com/ONSdigital/dp-observation-importer/kafka"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/go-ns/log"
)

// MessageConsumer provides a generic interface for consuming []byte messages
type MessageConsumer interface {
	Incoming() chan kafka.Message
	Closer() chan bool
}

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(event *ObservationExtracted) error
}

// Consume convert them to event instances, and pass the event to the provided handler.
func Consume(messageConsumer MessageConsumer, handler Handler) {
	for message := range messageConsumer.Incoming() {

		event, err := Unmarshal(message)
		if err != nil {
			log.Error(err, log.Data{"message": "failed to unmarshal event"})
			continue
		}

		log.Debug("event received", log.Data{"event": event})

		err = handler.Handle(event)
		if err != nil {
			log.Error(err, log.Data{"message": "failed to handle event"})
			continue
		}

		log.Debug("event processed - committing message", log.Data{"event": event})
		message.Commit()
		log.Debug("message committed", log.Data{"event": event})
	}
}

// Unmarshal converts an event instance to []byte.
func Unmarshal(message kafka.Message) (*ObservationExtracted, error) {
	var event ObservationExtracted
	err := schema.ObservationExtractedEvent.Unmarshal(message.GetData(), &event)
	return &event, err
}
