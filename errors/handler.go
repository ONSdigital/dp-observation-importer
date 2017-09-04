package errors

import (
	"time"

	eventhandler "github.com/ONSdigital/dp-import-reporter/handler"
	eventSchema "github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
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
	// todo - marshsal error + data to avro encoded message and send to handler.messageProducer.Output()

	log.Error(err, data)
	eventReport := eventhandler.EventReport{
		InstanceID: instanceID,
		EventType:  "error",
		EventMsg:   err.Error(),
	}

	producer := eventProducer()
	avroBytes, err := eventSchema.ReportedEventSchema.Marshal(&eventReport)
	if err != nil {
		log.Error(err, nil)
		return
	}

	producer.Output() <- avroBytes
	time.Sleep(time.Duration(5000 * time.Millisecond))
	producer.Closer() <- true

}

func eventProducer() kafka.Producer {
	// producerTopic := flag.String("topic", "event-reporter", "producer topic")
	brokers := []string{"localhost:9092"}
	producer := kafka.NewProducer(brokers, "event-reporter", int(2000000))

	return producer
}
