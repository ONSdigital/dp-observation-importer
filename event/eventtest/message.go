package eventtest

import (
	"github.com/ONSdigital/dp-observation-extractor/kafka"
)

var _ kafka.Message = (*Message)(nil)

// Message allows a mock message to return the configured data, and capture whether commit has been called.
type Message struct {
	Data      []byte
	committed bool
}

// GetData returns the data that was added to the struct.
func (m Message) GetData() []byte {
	return m.Data
}

// Commit captures the fact that the method was called.
func (m Message) Commit() {
	m.committed = true
}

// Committed returns true if commit was called on this message.
func (m Message) Committed() bool {
	return m.committed
}
