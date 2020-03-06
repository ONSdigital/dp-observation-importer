package event_test

import (
	"testing"
	"time"

	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-kafka/kafkatest"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/event/eventtest"
	. "github.com/smartystreets/goconvey/convey"
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

			go consumer.Consume(ctx, messageConsumer, batchSize, eventHandler, batchWaitTime, exit)
			<-eventHandler.EventUpdated

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

		messageConsumer := kafkatest.NewMessageConsumer(false)
		batchSize := 1
		eventHandler := eventtest.NewEventHandler()
		batchWaitTime := time.Second * 1
		exit := make(chan error, 1)

		consumer := event.NewConsumer()

		go consumer.Consume(ctx, messageConsumer, batchSize, eventHandler, batchWaitTime, exit)

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

			go consumer.Consume(ctx, messageConsumer, batchSize, eventHandler, batchWaitTime, exit)
			<-eventHandler.EventUpdated

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
		cChannels := kafka.CreateConsumerGroupChannels(true)
		messageConsumer := kafkatest.NewMessageConsumerWithChannels(&cChannels, false)

		batchSize := 3
		eventHandler := eventtest.NewEventHandler()
		batchWaitTime := time.Millisecond * 50
		exit := make(chan error, 1)

		messageDelay := time.Millisecond * 25
		message := kafkatest.NewMessage([]byte(marshal(expectedEvent)), 0)

		consumer := event.NewConsumer()

		SendMessagesWithDelay(cChannels.Upstream, message, messageDelay, 3)

		Convey("When consume is called", func() {

			go consumer.Consume(ctx, messageConsumer, batchSize, eventHandler, batchWaitTime, exit)

			<-eventHandler.EventUpdated

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

	cChannels := kafka.CreateConsumerGroupChannels(true)
	messageConsumer := kafkatest.NewMessageConsumerWithChannels(&cChannels, false)
	message := kafkatest.NewMessage([]byte(marshal(expectedEvent)), 0)
	cChannels.Upstream <- message

	return messageConsumer
}
