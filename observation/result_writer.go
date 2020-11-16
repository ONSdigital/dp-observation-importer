package observation

import (
	"context"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/log.go/log"
)

// MessageWriter writes observations as messages
type MessageWriter struct {
	MessageProducer MessageProducer
}

// MessageProducer dependency that writes messages
type MessageProducer interface {
	Channels() *kafka.ProducerChannels
}

// NewResultWriter returns a new observation message writer.
func NewResultWriter(messageProducer MessageProducer) *MessageWriter {
	return &MessageWriter{
		MessageProducer: messageProducer,
	}
}

// Write results as messages.
func (messageWriter MessageWriter) Write(ctx context.Context, results []*Result) {
	for _, result := range results {

		event := InsertedEvent{
			InstanceID:           result.InstanceID,
			ObservationsInserted: result.ObservationsInserted,
		}

		log.Event(ctx, "observations inserted, producing event message", log.INFO,
			log.Data{"event": event},
		)

		b, err := Marshal(event)
		if err != nil {
			log.Event(ctx, "failed to marshal observations inserted event", log.ERROR, log.Error(err), log.Data{
				"schema": "failed to marshal observations inserted event",
				"event":  event})
			continue
		}

		messageWriter.MessageProducer.Channels().Output <- b
	}
}

// Marshal converts the given ObservationsInsertedEvent to a []byte.
func Marshal(insertedEvent InsertedEvent) ([]byte, error) {
	bytes, err := schema.ObservationsInsertedEvent.Marshal(insertedEvent)
	return bytes, err
}
