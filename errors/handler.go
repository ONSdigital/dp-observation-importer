package errors

import (
	"github.com/ONSdigital/go-ns/log"
)

var _ Handler = (*KafkaHandler)(nil)

// Handler defines a generic interface for handling errors.
type Handler interface {
	Handle(err error, data log.Data)
}

// KafkaHandler provides an error handler that writes to a Kafka error topic.
type KafkaHandler struct {
	messageProducer MessageProducer
}

// MessageProducer dependency that writes messages
type MessageProducer interface {
	Output() chan []byte
	Closer() chan bool
}

// Handle logs the error and sends is as a Kafka message.
func (handler *KafkaHandler) Handle(err error, data log.Data) {
	log.Error(err, data)

	// todo - marshsal error + data to avro encoded message and send to handler.messageProducer.Output()
}
