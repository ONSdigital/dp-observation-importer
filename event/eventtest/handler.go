package eventtest

import (
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/go-ns/log"
)

var _ event.Handler = (*EventHandler)(nil)

// NewEventHandler returns a new mock event handler to capture event
func NewEventHandler() *EventHandler {

	events := make([]*event.ObservationExtracted, 0)

	return &EventHandler{
		Events: events,
	}
}

// EventHandler provides a mock implementation that captures events to check.
type EventHandler struct {
	Events []*event.ObservationExtracted
	Error  error
}

// Handle captures the given event and stores it for later assertions
func (handler *EventHandler) Handle(events []*event.ObservationExtracted) error {

	log.Debug("handle called", nil)
	handler.Events = events
	return handler.Error
}
