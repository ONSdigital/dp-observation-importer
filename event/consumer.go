package event

import (
	"context"
	"errors"
	"time"

	graph "github.com/ONSdigital/dp-graph/graph/driver"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/log.go/log"
)

// MessageConsumer provides a generic interface for consuming []byte messages (from Kafka)
type MessageConsumer interface {
	Channels() *kafka.ConsumerGroupChannels
	CommitAndRelease(msg kafka.Message)
}

// Handler represents a handler for processing a batch of events.
type Handler interface {
	Handle(ctx context.Context, events []*ObservationExtracted) error
}

// Consumer consumes event messages.
type Consumer struct {
	closing chan bool
	closed  chan bool
}

// NewConsumer returns a new consumer instance.
func NewConsumer() *Consumer {
	return &Consumer{
		closing: make(chan bool),
		closed:  make(chan bool),
	}
}

// Consume convert them to event instances, and pass the event to the provided handler.
func (consumer *Consumer) Consume(ctx context.Context, messageConsumer MessageConsumer,
	batchSize int,
	handler Handler,
	batchWaitTime time.Duration,
	errChan chan error) {

	go func() {
		defer close(consumer.closed)

		batch := NewBatch(batchSize)

		// Wait a batch full of messages.
		// If we do not get any messages for a time, just process the messages already in the batch.
		for {
			select {
			case msg := <-messageConsumer.Channels().Upstream:

				AddMessageToBatch(ctx, batch, msg, handler, errChan)
				messageConsumer.CommitAndRelease(msg)

			case <-time.After(batchWaitTime):
				if batch.IsEmpty() {
					continue
				}

				log.Event(ctx, "batch wait time reached. proceeding with batch", log.INFO, log.Data{"batchsize": batch.Size()})
				ProcessBatch(ctx, handler, batch, errChan)

			case <-consumer.closing:
				log.Event(ctx, "closing event consumer loop", log.INFO)
				return
			}
		}
	}()
}

// Close safely closes the consumer and releases all resources
func (consumer *Consumer) Close(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	close(consumer.closing)

	select {
	case <-consumer.closed:
		log.Event(ctx, "successfully closed event consumer", log.INFO)
		return nil
	case <-ctx.Done():
		log.Event(ctx, "shutdown context time exceeded, skipping graceful shutdown of event consumer", log.INFO)
		return errors.New("Shutdown context timed out")
	}
}

// AddMessageToBatch will attempt to add the message to the batch and determine if it should be processed.
func AddMessageToBatch(ctx context.Context, batch *Batch, msg kafka.Message, handler Handler, errChan chan error) {
	batch.Add(ctx, msg)
	if batch.IsFull() {
		log.Event(ctx, "batch is full - processing batch", log.INFO, log.Data{"batchsize": batch.Size()})
		ProcessBatch(ctx, handler, batch, errChan)
	}
}

// ProcessBatch will attempt to handle and commit the batch, or shutdown if something goes horribly wrong.
func ProcessBatch(ctx context.Context, handler Handler, batch *Batch, errChan chan error) {
	err := handler.Handle(ctx, batch.Events())
	if err != nil {
		log.Event(ctx, "error processing batch", log.ERROR, log.Error(err))
		errChan <- err
		// If the error type is non retriable then we should commit the message batch,
		// because we know it will never succeed
		if _, ok := err.(graph.ErrNonRetriable); ok {
			batch.Commit()
		}
		return
	}

	batch.Commit()
}
