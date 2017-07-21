package event_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/ONSdigital/dp-observation-importer/event/eventtest"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/schema"
)

func TestIsEmpty(t *testing.T) {

	Convey("Given a batch that is not empty", t, func() {

		expectedEvent := event.ObservationExtracted{ InstanceID:"123", Row:"the,row,content" }
		message := &eventtest.Message{Data: []byte(Marshal(expectedEvent))}

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

		expectedEvent := event.ObservationExtracted{ InstanceID:"123", Row:"the,row,content" }
		message := &eventtest.Message{Data: []byte(Marshal(expectedEvent))}

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

func TestSize(t *testing.T) {

	Convey("Given a batch", t, func() {

		expectedEvent := event.ObservationExtracted{ InstanceID:"123", Row:"the,row,content" }
		message := &eventtest.Message{Data: []byte(Marshal(expectedEvent))}

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

		expectedEvent := event.ObservationExtracted{ InstanceID:"123", Row:"the,row,content" }
		message := &eventtest.Message{Data: []byte(Marshal(expectedEvent))}

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
		Row:        "some,row,content",
	}
	return expectedEvent
}
