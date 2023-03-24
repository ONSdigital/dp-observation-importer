package event

import (
	"context"
	"errors"
	"time"

	graph "github.com/ONSdigital/dp-graph/v2/graph/driver"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
)

// MessageConsumer provides a generic interface for consuming []byte messages (from Kafka)
type MessageConsumer interface {
	Channels() *kafka.ConsumerGroupChannels
}

// Handler represents a handler for processing a batch of events.
type Handler interface {
	Handle(ctx context.Context, events []*ObservationExtracted) error
}

// Consumer consumes event messages.
type Consumer struct {
	closing chan eventClose
	closed  chan bool
}

type eventClose struct {
	ctx context.Context
}

// NewConsumer returns a new consumer instance.
func NewConsumer() *Consumer {
	return &Consumer{
		closing: make(chan eventClose),
		closed:  make(chan bool),
	}
}

// Consume convert them to event instances, and pass the event to the provided handler.
func (consumer *Consumer) Consume(messageConsumer MessageConsumer,
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
			delay := time.NewTimer(batchWaitTime)
			select {
			case msg := <-messageConsumer.Channels().Upstream:
				// Ensure timer is stopped and its resources are freed
				if !delay.Stop() {
					// if the timer has been stopped then read from the channel
					<-delay.C
				}
				ctx := context.Background()

				AddMessageToBatch(ctx, batch, msg, handler, errChan)
				msg.Release()

			case <-delay.C:
				if batch.IsEmpty() {
					continue
				}

				ctx := context.Background()

				log.Info(ctx, "batch wait time reached. proceeding with batch", log.Data{"batchsize": batch.Size()})
				ProcessBatch(ctx, handler, batch, errChan)

			case eventClose := <-consumer.closing:
				// Ensure timer is stopped and its resources are freed
				if !delay.Stop() {
					// if the timer has been stopped then read from the channel
					<-delay.C
				}
				log.Info(eventClose.ctx, "closing event consumer loop")
				close(consumer.closing)
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

	consumer.closing <- eventClose{ctx: ctx}

	select {
	case <-consumer.closed:
		log.Info(ctx, "successfully closed event consumer")
		return nil
	case <-ctx.Done():
		log.Info(ctx, "shutdown context time exceeded, skipping graceful shutdown of event consumer")
		return errors.New("Shutdown context timed out")
	}
}

// AddMessageToBatch will attempt to add the message to the batch and determine if it should be processed.
func AddMessageToBatch(ctx context.Context, batch *Batch, msg kafka.Message, handler Handler, errChan chan error) {
	batch.Add(ctx, msg)
	if batch.IsFull() {
		log.Info(ctx, "batch is full - processing batch", log.Data{"batchsize": batch.Size()})
		ProcessBatch(ctx, handler, batch, errChan)
	}
}

// ProcessBatch will attempt to handle and commit the batch, or shutdown if something goes horribly wrong.
func ProcessBatch(ctx context.Context, handler Handler, batch *Batch, errChan chan error) {
	err := handler.Handle(ctx, batch.Events())
	if err != nil {
		log.Error(ctx, "error processing batch", err)
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
