package event

import "github.com/ONSdigital/go-ns/log"
import (
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/dp-observation-importer/errors"
)



type Batch struct {
	maxSize int
	events []*ObservationExtracted
	errorHandler errors.Handler
}

type Message interface {
	GetData() []byte
	Commit()
}

func NewBatch(batchSize int, errorHandler errors.Handler) *Batch {
	events := make([]*ObservationExtracted, 0, batchSize)

	return &Batch{
		maxSize:batchSize,
		events: events,
	}
}

func (batch *Batch) Add(message Message) {

	event, err := Unmarshal(message)
	if err != nil {
		batch.errorHandler.Handle(err, log.Data{"message": "failed to unmarshal event"})
	}

	batch.events = append(batch.events, event)
}

func (batch *Batch) Size() int {
	return len(batch.events)
}

func (batch *Batch) IsFull() bool {
	return len(batch.events) == batch.maxSize
}

func (batch *Batch) Events() []*ObservationExtracted {
	return batch.events
}

func (batch *Batch) IsEmpty() bool {
	return len(batch.events) == 0
}

func (batch *Batch) Commit() error {
	return nil
}

// Unmarshal converts an event instance to []byte.
func Unmarshal(message Message) (*ObservationExtracted, error) {
	var event ObservationExtracted
	err := schema.ObservationExtractedEvent.Unmarshal(message.GetData(), &event)
	return &event, err
}