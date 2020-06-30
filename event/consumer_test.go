package event_test

import (
	"testing"
	"time"

	"github.com/ONSdigital/dp-kafka/kafkatest"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/event/eventtest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConsume(t *testing.T) {

	Convey("Given a consumer with a mocked message producer with an expected message", t, func() {

		expectedEvent := event.ObservationExtracted{InstanceID: "123", Row: "the,row,content"}
		messageConsumer := kafkatest.NewMessageConsumer(false)
		batchSize := 1
		eventHandler := eventtest.NewEventHandler()
		batchWaitTime := time.Second
		exit := make(chan error, 1)

		consumer := event.NewConsumer()

		Convey("When consume is called", func() {

			go consumer.Consume(messageConsumer, batchSize, eventHandler, batchWaitTime, exit)

			message := kafkatest.NewMessage(marshal(expectedEvent), 0)
			messageConsumer.Channels().Upstream <- message

			<-eventHandler.EventUpdated

			Convey("The consumer is released to consume the next message", func() {
				So(len(messageConsumer.ReleaseCalls()), ShouldEqual, 1)
			})

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
		batchWaitTime := time.Second
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
		messageConsumer := kafkatest.NewMessageConsumer(false)
		batchSize := 2
		eventHandler := eventtest.NewEventHandler()
		batchWaitTime := 50 * time.Millisecond
		exit := make(chan error, 1)

		consumer := event.NewConsumer()

		Convey("When consume is called with a batch size of 2, and no other messages are consumed", func() {

			go consumer.Consume(messageConsumer, batchSize, eventHandler, batchWaitTime, exit)

			message := kafkatest.NewMessage(marshal(expectedEvent), 0)
			messageConsumer.Channels().Upstream <- message
			<-messageConsumer.Channels().UpstreamDone

			<-eventHandler.EventUpdated

			Convey("The consumer is released to consume the next message", func() {
				So(len(messageConsumer.ReleaseCalls()), ShouldEqual, 1)
			})

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
		messageConsumer := kafkatest.NewMessageConsumer(true)

		// batchsize is also the maximum number of messages which would get processed
		// before assertions are being made in this test due to the mocked version of event handler
		batchSize := 3
		eventHandler := eventtest.NewEventHandler()
		batchWaitTime := 150 * time.Millisecond
		exit := make(chan error, 1)

		messageDelay := time.Millisecond * 25
		consumer := event.NewConsumer()

		Convey("When consume is called", func() {

			go consumer.Consume(messageConsumer, batchSize, eventHandler, batchWaitTime, exit)

			message := kafkatest.NewMessage(marshal(expectedEvent), 0)
			go SendMessagesWithDelay(messageConsumer, []*kafkatest.Message{message, message, message}, messageDelay)

			// wait for event updated channel to receive event from the mock event handler
			<-eventHandler.EventUpdated

			Convey("The consumer is released to consume the next message", func() {
				So(len(messageConsumer.ReleaseCalls()), ShouldEqual, 3)
			})

			Convey("The expected events are sent to the handler in one batch - i.e. the timeout is not hit", func() {
				So(len(eventHandler.Events), ShouldEqual, 3)

				event := eventHandler.Events[0]
				So(event.InstanceID, ShouldEqual, expectedEvent.InstanceID)
				So(event.Row, ShouldEqual, expectedEvent.Row)
			})
		})

	})
}

func SendMessagesWithDelay(messageConsumer *kafkatest.MessageConsumer, messages []*kafkatest.Message, messageDelay time.Duration) {
	for _, message := range messages {
		time.Sleep(messageDelay)
		messageConsumer.Channels().Upstream <- message
		<-messageConsumer.Channels().UpstreamDone
	}
}
