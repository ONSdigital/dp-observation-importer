package event

import (
	"context"
	"errors"
	"time"

	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

// MessageConsumer provides a generic interface for consuming []byte messages (from Kafka)
type MessageConsumer interface {
	Incoming() chan kafka.Message
}

// Handler represents a handler for processing a batch of events.
type Handler interface {
	Handle(events []*ObservationExtracted) error
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
func (consumer *Consumer) Consume(messageConsumer MessageConsumer,
	batchSize int,
	handler Handler,
	batchWaitTime time.Duration,
	error chan error) {

	go func() {
		defer close(consumer.closed)

		batch := NewBatch(batchSize)

		// Wait a batch full of messages.
		// If we do not get any messages for a time, just process the messages already in the batch.
		for {
			select {
			case msg := <-messageConsumer.Incoming():

				AddMessageToBatch(batch, msg, handler, error)

			case <-time.After(batchWaitTime):

				if batch.IsEmpty() {
					continue
				}

				log.Debug("batch wait time reached. proceeding with batch", log.Data{"batchsize": batch.Size()})
				ProcessBatch(handler, batch, error)

			case <-consumer.closing:
				log.Info("closing event consumer loop", nil)
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
		log.Info("successfully closed event consumer", nil)
		return nil
	case <-ctx.Done():
		log.Info("shutdown context time exceeded, skipping graceful shutdown of event consumer", nil)
		return errors.New("Shutdown context timed out")
	}
}

// AddMessageToBatch will attempt to add the message to the batch and determine if it should be processed.
func AddMessageToBatch(batch *Batch, msg kafka.Message, handler Handler, error chan error) {
	batch.Add(msg)
	if batch.IsFull() {
		log.Debug("batch is full - processing batch", log.Data{"batchsize": batch.Size()})
		ProcessBatch(handler, batch, error)
	}
}

// ProcessBatch will attempt to handle and commit the batch, or shutdown if something goes horribly wrong.
func ProcessBatch(handler Handler, batch *Batch, error chan error) {
	err := handler.Handle(batch.Events())
	if err != nil {
		log.Error(err, log.Data{})
		error <- err
		// If the error type is non retriable then we should commit the message batch,
		// because we know it will never succeed
		if _, ok := err.(observation.ErrNonRetriable); ok {
			batch.Commit()
		}
		return
	}

	batch.Commit()
}
