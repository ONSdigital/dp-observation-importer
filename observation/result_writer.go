package observation

import (
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/go-ns/log"
)

// MessageWriter writes observations as messages
type MessageWriter struct {
	messageProducer MessageProducer
}

// MessageProducer dependency that writes messages
type MessageProducer interface {
	Output() chan []byte
	Closer() chan bool
}

// NewMessageWriter returns a new observation message writer.
func NewResultWriter(messageProducer MessageProducer) *MessageWriter {
	return &MessageWriter{
		messageProducer: messageProducer,
	}
}

// Write results as messages.
func (messageWriter MessageWriter) Write(results []*Result) {

	for _, result := range results {

		event := InsertedEvent{
			InstanceID:           result.InstanceID,
			ObservationsInserted: result.ObservationsInserted,
		}

		bytes, err := Marshal(event)
		if err != nil {
			log.Error(err, log.Data{
				"schema": "failed to marshal observations inserted event",
				"event":  event})
		}

		messageWriter.messageProducer.Output() <- bytes
	}
}

// Marshal converts the given ObservationsInsertedEvent to a []byte.
func Marshal(insertedEvent InsertedEvent) ([]byte, error) {
	bytes, err := schema.ObservationsInsertedEvent.Marshal(insertedEvent)
	return bytes, err
}
