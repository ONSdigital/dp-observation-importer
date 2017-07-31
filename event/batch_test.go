package event_test

import (
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/event/eventtest"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIsEmpty(t *testing.T) {

	Convey("Given a batch that is not empty", t, func() {

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage([]byte(Marshal(*expectedEvent)))

		batchSize := 1
		errorHandler := eventtest.ErrorHandler{}

		batch := event.NewBatch(batchSize, errorHandler)

		Convey("When the batch has no messages added", func() {
			Convey("The batch is empty", func() {
				So(batch.IsEmpty(), ShouldBeTrue)
			})
		})

		Convey("When the batch has a message added", func() {

			batch.Add(message)

			Convey("The batch is now empty", func() {
				So(batch.IsEmpty(), ShouldBeFalse)
			})
		})
	})
}

func TestAdd(t *testing.T) {

	Convey("Given a batch", t, func() {

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage([]byte(Marshal(*expectedEvent)))

		batchSize := 1
		errorHandler := eventtest.ErrorHandler{}

		batch := event.NewBatch(batchSize, errorHandler)

		Convey("When add is called with a valid message", func() {

			batch.Add(message)

			Convey("The batch contains the expected event.", func() {
				So(len(batch.Events()), ShouldEqual, 1)
				So(batch.Events()[0].Row, ShouldEqual, expectedEvent.Row)
				So(batch.Events()[0].InstanceID, ShouldEqual, expectedEvent.InstanceID)
			})
		})
	})
}

func TestCommit(t *testing.T) {

	Convey("Given a batch with two valid messages", t, func() {

		expectedEvent := event.ObservationExtracted{InstanceID: "123", Row: "the,row,content"}
		expectedLastEvent := event.ObservationExtracted{InstanceID: "123", Row: "last,row,content"}
		message := kafkatest.NewMessage([]byte(Marshal(expectedEvent)))
		lastMessage := kafkatest.NewMessage([]byte(Marshal(expectedLastEvent)))

		batchSize := 2
		errorHandler := eventtest.ErrorHandler{}

		batch := event.NewBatch(batchSize, errorHandler)

		batch.Add(message)
		batch.Add(lastMessage)

		batch.Commit()

		Convey("When commit is called", func() {

			Convey("The last message in the batch is committed", func() {
				So(message.Committed(), ShouldBeFalse)
				So(lastMessage.Committed(), ShouldBeTrue)
			})

			Convey("The batch is emptied.", func() {
				So(len(batch.Events()), ShouldEqual, 0)
				So(batch.IsEmpty(), ShouldBeTrue)
				So(batch.IsFull(), ShouldBeFalse)
				So(batch.Size(), ShouldEqual, 0)
			})

			Convey("The batch can be reused", func() {
				batch.Add(lastMessage)

				So(len(batch.Events()), ShouldEqual, 1)
				So(batch.IsEmpty(), ShouldBeFalse)
				So(batch.IsFull(), ShouldBeFalse)
				So(batch.Size(), ShouldEqual, 1)

				So(batch.Events()[0].Row, ShouldEqual, expectedLastEvent.Row)
				So(batch.Events()[0].InstanceID, ShouldEqual, expectedLastEvent.InstanceID)

				batch.Add(message)

				So(len(batch.Events()), ShouldEqual, 2)
				So(batch.IsEmpty(), ShouldBeFalse)
				So(batch.IsFull(), ShouldBeTrue)
				So(batch.Size(), ShouldEqual, 2)

				So(batch.Events()[1].Row, ShouldEqual, expectedEvent.Row)
				So(batch.Events()[1].InstanceID, ShouldEqual, expectedEvent.InstanceID)
			})
		})
	})
}

func TestSize(t *testing.T) {

	Convey("Given a batch", t, func() {

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage([]byte(Marshal(*expectedEvent)))

		batchSize := 1
		errorHandler := eventtest.ErrorHandler{}

		batch := event.NewBatch(batchSize, errorHandler)

		So(batch.Size(), ShouldEqual, 0)

		Convey("When add is called with a valid message", func() {

			batch.Add(message)

			Convey("The batch size should increase.", func() {
				So(batch.Size(), ShouldEqual, 1)
				batch.Add(message)
				So(batch.Size(), ShouldEqual, 2)
			})
		})
	})
}

func TestIsFull(t *testing.T) {

	Convey("Given a batch with a size of 2", t, func() {

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage([]byte(Marshal(*expectedEvent)))

		batchSize := 2
		errorHandler := eventtest.ErrorHandler{}

		batch := event.NewBatch(batchSize, errorHandler)

		So(batch.IsFull(), ShouldBeFalse)

		Convey("When the number of messages added equals the batch size", func() {

			batch.Add(message)
			So(batch.IsFull(), ShouldBeFalse)
			batch.Add(message)

			Convey("The batch should be full.", func() {
				So(batch.IsFull(), ShouldBeTrue)
			})
		})
	})
}

func TestToEvent(t *testing.T) {

	Convey("Given a event schema encoded using avro", t, func() {

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage([]byte(Marshal(*expectedEvent)))

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
		Row:        "some,row,content",
	}
	return expectedEvent
}
