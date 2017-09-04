package errors

import (
	"github.com/ONSdigital/go-ns/log"
)

var _ Handler = (*KafkaHandler)(nil)

//go:generate moq -out errorstest/handler.go -pkg errorstest . Handler

// Handler defines a generic interface for handling errors.
type Handler interface {
	Handle(instanceID string, err error, data log.Data)
}

// KafkaHandler provides an error handler that writes to a Kafka error topic.
type KafkaHandler struct {
	messageProducer MessageProducer
}

// NewKafkaHandler returns a new KafkaHadler that sends error messages to the given messageProducer.
func NewKafkaHandler(messageProducer MessageProducer) *KafkaHandler {
	return &KafkaHandler{
		messageProducer: messageProducer,
	}
}

// MessageProducer dependency that writes messages
type MessageProducer interface {
	Output() chan []byte
}

// Handle logs the error and sends is as a Kafka message.
func (handler *KafkaHandler) Handle(instanceID string, err error, data log.Data) {

	if data == nil {
		data = log.Data{}
	}

	data["instance_id"] = instanceID
	log.Error(err, data)

	// todo - marshsal error + data to avro encoded message and send to handler.messageProducer.Output()
}
