package errorstest

import "github.com/ONSdigital/dp-observation-extractor/observation"

var _ observation.MessageProducer = (*MessageProducer)(nil)

// MessageProducer provides a mock that allows injection of the required output channel.
type MessageProducer struct {
	observation.MessageProducer
	OutputChannel chan []byte
}

// Output returns the injected output channel for testing.
func (messageProducer MessageProducer) Output() chan []byte {
	return messageProducer.OutputChannel
}
