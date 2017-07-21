package eventtest

import (
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/kafka"
)

var _ event.MessageConsumer = (*MessageConsumer)(nil)

// NewMessageConsumer creates a consumer using the given channel.
func NewMessageConsumer(messages chan kafka.Message) *MessageConsumer {
	return &MessageConsumer{
		messages: messages,
		closer:   nil,
	}
}

// MessageConsumer is a mock that provides the stored schema channel.
type MessageConsumer struct {
	messages chan kafka.Message
	closer   chan bool
}

// Incoming returns the stored schema channel.
func (consumer *MessageConsumer) Incoming() chan kafka.Message {
	return consumer.messages
}

// Closer returns the stored closer channel.
func (consumer *MessageConsumer) Closer() chan bool {
	return consumer.closer
}
