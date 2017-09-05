package event_test

import (
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/event/eventtest"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestConsume(t *testing.T) {

	Convey("Given a consumer with a mocked message producer with an expected message", t, func() {

		expectedEvent := event.ObservationExtracted{InstanceID: "123", Row: "the,row,content"}
		messageConsumer := newMockConsumer(expectedEvent)
		batchSize := 1
		eventHandler := eventtest.NewEventHandler()
		batchWaitTime := time.Second * 1
		exit := make(chan error, 1)

		consumer := event.NewConsumer()

		Convey("When consume is called", func() {

			go consumer.Consume(messageConsumer, batchSize, eventHandler, batchWaitTime, exit)

			waitForEventsToBeSentToHandler(eventHandler, exit)

			Convey("The expected event is sent to the handler", func() {
				So(len(eventHandler.Events), ShouldEqual, 1)

				event := eventHandler.Events[0]
				So(event.InstanceID, ShouldEqual, expectedEvent.InstanceID)
				So(event.Row, ShouldEqual, expectedEvent.Row)
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Given a consumer", t, func() {

		messages := make(chan kafka.Message, 3)
		messageConsumer := kafkatest.NewMessageConsumer(messages)
		batchSize := 1
		eventHandler := eventtest.NewEventHandler()
		batchWaitTime := time.Second * 1
		exit := make(chan error, 1)

		consumer := event.NewConsumer()

		go consumer.Consume(messageConsumer, batchSize, eventHandler, batchWaitTime, exit)

		Convey("When close is called", func() {

			err := consumer.Close(nil)

			Convey("The expected event is sent to the handler", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestConsume_Timeout(t *testing.T) {

	Convey("Given a consumer with a mocked message producer with an expected message", t, func() {

		expectedEvent := event.ObservationExtracted{InstanceID: "123", Row: "the,row,content"}
		messageConsumer := newMockConsumer(expectedEvent)
		batchSize := 2
		eventHandler := eventtest.NewEventHandler()
		batchWaitTime := time.Millisecond * 50
		exit := make(chan error, 1)

		consumer := event.NewConsumer()

		Convey("When consume is called with a batch size of 2, and no other messages are consumed", func() {

			go consumer.Consume(messageConsumer, batchSize, eventHandler, batchWaitTime, exit)

			waitForEventsToBeSentToHandler(eventHandler, exit)

			Convey("The consumer timeout is hit and the single event is sent to the handler anyway", func() {
				So(len(eventHandler.Events), ShouldEqual, 1)

				event := eventHandler.Events[0]
				So(event.InstanceID, ShouldEqual, expectedEvent.InstanceID)
				So(event.Row, ShouldEqual, expectedEvent.Row)
			})
		})

	})
}

func TestConsume_DelayedMessages(t *testing.T) {

	Convey("Given a consumer with a mocked message producer that produces messages every 20ms", t, func() {

		expectedEvent := event.ObservationExtracted{InstanceID: "123", Row: "the,row,content"}
		messages := make(chan kafka.Message, 3)
		messageConsumer := kafkatest.NewMessageConsumer(messages)

		batchSize := 3
		eventHandler := eventtest.NewEventHandler()
		batchWaitTime := time.Millisecond * 50
		exit := make(chan error, 1)

		messageDelay := time.Millisecond * 25
		message := kafkatest.NewMessage([]byte(marshal(expectedEvent)))

		consumer := event.NewConsumer()

		SendMessagesWithDelay(messages, message, messageDelay, 3)

		Convey("When consume is called", func() {

			go consumer.Consume(messageConsumer, batchSize, eventHandler, batchWaitTime, exit)

			waitForEventsToBeSentToHandler(eventHandler, exit)

			Convey("The expected events are sent to the handler in one batch - i.e. the timeout is not hit", func() {
				So(len(eventHandler.Events), ShouldEqual, 3)

				event := eventHandler.Events[2]
				So(event.InstanceID, ShouldEqual, expectedEvent.InstanceID)
				So(event.Row, ShouldEqual, expectedEvent.Row)
			})
		})

	})
}

func SendMessagesWithDelay(messages chan kafka.Message, message kafka.Message, messageDelay time.Duration, numberOfMessages int) {
	go func() {
		for i := 0; i < numberOfMessages; i++ {
			time.Sleep(messageDelay)
			messages <- message
		}
	}()
}

func newMockConsumer(expectedEvent event.ObservationExtracted) event.MessageConsumer {

	messages := make(chan kafka.Message, 1)
	messageConsumer := kafkatest.NewMessageConsumer(messages)
	message := kafkatest.NewMessage([]byte(marshal(expectedEvent)))
	messages <- message
	return messageConsumer

}

func waitForEventsToBeSentToHandler(eventHandler *eventtest.EventHandler, exit chan error) {

	start := time.Now()
	timeout := start.Add(time.Millisecond * 500)
	for {
		if len(eventHandler.Events) > 0 {
			log.Debug("events have been sent to the handler", nil)
			close(exit)
			break
		}

		if time.Now().After(timeout) {
			log.Debug("timeout hit", nil)
			close(exit)
			break
		}

		time.Sleep(time.Millisecond * 10)
	}
}
