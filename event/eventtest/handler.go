package eventtest

import (
	"context"

	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/log.go/log"
)

var _ event.Handler = (*EventHandler)(nil)

// NewEventHandler returns a new mock event handler to capture event
func NewEventHandler() *EventHandler {

	events := make([]*event.ObservationExtracted, 0)
	eventUpdated := make(chan bool)

	return &EventHandler{
		Events:       events,
		EventUpdated: eventUpdated,
	}
}

// EventHandler provides a mock implementation that captures events to check.
type EventHandler struct {
	Events       []*event.ObservationExtracted
	Error        error
	EventUpdated chan bool
}

// Handle captures the given event and stores it for later assertions
func (handler *EventHandler) Handle(ctx context.Context, events []*event.ObservationExtracted) error {
	log.Event(ctx, "handle called", log.INFO)
	handler.Events = events

	handler.EventUpdated <- true
	return handler.Error
}
