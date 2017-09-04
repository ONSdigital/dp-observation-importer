package errors

import (
	"time"

	eventhandler "github.com/ONSdigital/dp-import-reporter/handler"
	eventSchema "github.com/ONSdigital/dp-import-reporter/schema"
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
	Closer() chan bool
}

// Handle logs the error and sends is as a Kafka message.
func (handler *KafkaHandler) Handle(instanceID string, err error, data log.Data) {

	if data == nil {
		data = log.Data{}
	}

	data["instance_id"] = instanceID

	log.Error(err, data)
	eventReport := eventhandler.EventReport{
		InstanceID: instanceID,
		EventType:  "error",
		EventMsg:   err.Error(),
	}

	errMsg, err := eventSchema.ReportedEventSchema.Marshal(&eventReport)
	if err != nil {
		log.Error(err, nil)
		return
	}

	handler.messageProducer.Output() <- errMsg
	time.Sleep(time.Duration(1000 * time.Millisecond))

}
