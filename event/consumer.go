package event

import (
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"sync"
	"time"
)

// MessageConsumer provides a generic interface for consuming []byte messages (from Kafka)
type MessageConsumer interface {
	Incoming() chan kafka.Message
}

// Handler represents a handler for processing a batch of events.
type Handler interface {
	Handle(events []*ObservationExtracted) error
}

type Consumer struct {
	closer chan bool
	wg     *sync.WaitGroup
}

// NewConsumerGroup returns a new consumer group using default configuration.
func NewConsumer() *Consumer {

	var wg sync.WaitGroup

	c := Consumer{
		closer: make(chan bool),
		wg:     &wg,
	}

	return &c
}

// Consume convert them to event instances, and pass the event to the provided handler.
func (consumer *Consumer) Consume(messageConsumer MessageConsumer,
	batchSize int,
	handler Handler,
	batchWaitTime time.Duration,
	exit chan error) {

	consumer.wg.Add(1)

	go func() {

		batch := NewBatch(batchSize)

		// Wait a batch full of messages.
		// If we do not get any messages for a time, just process the messages already in the batch.
		for {
			select {
			case msg := <-messageConsumer.Incoming():

				AddMessageToBatch(batch, msg, handler, exit)

			case <-time.After(batchWaitTime):

				if batch.IsEmpty() {
					continue
				}

				log.Debug("batch wait time reached. proceeding with batch", log.Data{"batchsize": batch.Size()})
				ProcessBatch(handler, batch, exit)

			case <-exit:
				return
			}
		}
	}()
}

// Close safely closes the consumer and releases all resources
func (consumer *Consumer) Close() (err error) {

	consumer.closer <- true
	consumer.wg.Wait()

	return nil
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
		return
	}

	batch.Commit()
}
