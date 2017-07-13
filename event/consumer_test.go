package event_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/ONSdigital/dp-observation-importer/kafka"
	"github.com/ONSdigital/dp-observation-importer/event/eventtest"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/schema"
)

func TestConsume_UnmarshallError(t *testing.T) {
	Convey("Given an event consumer with an invalid schema and a valid schema", t, func() {

		messages := make(chan kafka.Message, 2)
		messageConsumer := eventtest.NewMessageConsumer(messages)
		handler := eventtest.NewEventHandler()

		expectedEvent := getExampleEvent()

		messages <- &eventtest.Message{Data: []byte("invalid schema")}
		messages <- &eventtest.Message{Data: Marshal(*expectedEvent)}
		close(messages)

		Convey("When consume messages is called", func() {

			event.Consume(messageConsumer, handler)

			Convey("Only the valid event is sent to the handler ", func() {
				So(len(handler.Events), ShouldEqual, 1)

				event := handler.Events[0]
				So(event.Row, ShouldEqual, expectedEvent.Row)
				So(event.InstanceID, ShouldEqual, expectedEvent.InstanceID)
			})
		})
	})
}

func TestConsume(t *testing.T) {

	Convey("Given an event consumer with a valid schema", t, func() {

		messages := make(chan kafka.Message, 1)
		messageConsumer := eventtest.NewMessageConsumer(messages)
		handler := eventtest.NewEventHandler()

		expectedEvent := getExampleEvent()

		message := &eventtest.Message{Data: Marshal(*expectedEvent)}

		messages <- message
		close(messages)

		Convey("When consume is called", func() {

			event.Consume(messageConsumer, handler)

			Convey("A event is sent to the handler ", func() {
				So(len(handler.Events), ShouldEqual, 1)

				event := handler.Events[0]
				So(event.Row, ShouldEqual, expectedEvent.Row)
				So(event.InstanceID, ShouldEqual, expectedEvent.InstanceID)
			})

			Convey("The message is committed", func() {
				So(message.Committed(), ShouldEqual, true)
			})
		})
	})
}

func TestToEvent(t *testing.T) {

	Convey("Given a event schema encoded using avro", t, func() {

		expectedEvent := getExampleEvent()
		message := &eventtest.Message{Data: Marshal(*expectedEvent)}

		Convey("When the expectedEvent is unmarshalled", func() {

			event, err := event.Unmarshal(message)

			Convey("The expectedEvent has the expected values", func() {
				So(err, ShouldBeNil)
				So(event.Row, ShouldEqual, expectedEvent.Row)
				So(event.InstanceID, ShouldEqual, expectedEvent.InstanceID)
			})
		})
	})
}

// Marshal helper method to marshal a event into a []byte
func Marshal(event event.ObservationExtracted) []byte {
	bytes, err := schema.ObservationExtractedEvent.Marshal(event)
	So(err, ShouldBeNil)
	return bytes
}

func getExampleEvent() *event.ObservationExtracted {
	expectedEvent := &event.ObservationExtracted{
		InstanceID: "1234",
		Row:    "some,row,content",
	}
	return expectedEvent
}
