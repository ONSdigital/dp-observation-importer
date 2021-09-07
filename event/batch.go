package event

import (
	"context"

	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

// Batch handles adding raw messages to a batch of ObservationExtracted events.
type Batch struct {
	maxSize  int
	events   []*ObservationExtracted
	messages []Message
}

// Message represents a single message to be added to the batch.
type Message interface {
	GetData() []byte
	Mark()
	Commit()
}

// NewBatch returns a new batch instance of the given size.
func NewBatch(batchSize int) *Batch {
	events := make([]*ObservationExtracted, 0, batchSize)

	return &Batch{
		maxSize: batchSize,
		events:  events,
	}
}

// Add a message to the batch.
func (batch *Batch) Add(ctx context.Context, message Message) {

	event, err := Unmarshal(message)
	if err != nil {
		log.Error(ctx, "failed to unmarshal event", err)
		return
	}

	batch.messages = append(batch.messages, message)
	batch.events = append(batch.events, event)

}

// Size returns the number of events currently in the batch.
func (batch *Batch) Size() int {
	return len(batch.events)
}

// IsFull returns true if the batch is full based on the configured maxSize.
func (batch *Batch) IsFull() bool {
	return len(batch.events) == batch.maxSize
}

// Events returns the events currenty in the batch.
func (batch *Batch) Events() []*ObservationExtracted {
	return batch.events
}

// IsEmpty returns true if the batch has no events in it.
func (batch *Batch) IsEmpty() bool {
	return len(batch.events) == 0
}

// Commit is called when the batch has been processed. The last message has been released already, so at this point we just need to commit
func (batch *Batch) Commit() {
	for i, msg := range batch.messages {
		if i < len(batch.messages)-1 {
			msg.Mark()
			continue
		}
		msg.Commit()
	}
	batch.Clear()
}

// Clear will reset to batch to contain no events.
func (batch *Batch) Clear() {
	batch.events = batch.events[0:0]
	batch.messages = batch.messages[0:0]
}

// Unmarshal converts an event instance to []byte.
func Unmarshal(message Message) (*ObservationExtracted, error) {
	var event ObservationExtracted
	err := schema.ObservationExtractedEvent.Unmarshal(message.GetData(), &event)
	return &event, err
}
