package event

import (
	"github.com/ONSdigital/dp-observation-importer/kafka"
	"time"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-observation-importer/errors"
)

// MessageConsumer provides a generic interface for consuming []byte messages
type MessageConsumer interface {
	Incoming() chan kafka.Message
	Closer() chan bool
}

// Handler represents a handler for processing a batch of events.
type Handler interface {
	Handle(events []*ObservationExtracted) (error)
}

// Consume convert them to event instances, and pass the event to the provided handler.
func Consume(messageConsumer MessageConsumer,
	batchSize int,
	errorHandler errors.Handler,
	handler Handler,
	batchWaitTime time.Duration,
	exit chan struct{}) {

	batch := NewBatch(batchSize, errorHandler)

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

			log.Debug("batch wait time reached. proceeding with batch", log.Data{"batchsize": batch.Size() })
			ProcessBatch(handler, batch, exit)

		case <-exit:
			return
		}
	}
}

func AddMessageToBatch(batch *Batch, msg kafka.Message, handler Handler, exit chan struct{})  {
	batch.Add(msg)
	if batch.IsFull() {
		log.Debug("batch is full - processing batch", log.Data{"batchsize": batch.Size() })
		ProcessBatch(handler, batch, exit)
	}
}

func ProcessBatch(handler Handler, batch *Batch, exit chan struct{}) {

	err := handler.Handle(batch.Events())
	if err != nil {
		close(exit)
		return
	}

	batch.Commit()
}
