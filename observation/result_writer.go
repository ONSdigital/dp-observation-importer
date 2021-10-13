package observation

import (
	"context"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/log.go/v2/log"
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

		log.Info(ctx, "observations inserted, producing event message",
			log.Data{"event": event},
		)

		b, err := Marshal(event)
		if err != nil {
			log.Error(ctx, "failed to marshal observations inserted event", err, log.Data{
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
