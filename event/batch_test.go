package event_test

import (
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIsEmpty(t *testing.T) {

	Convey("Given a batch that is not empty", t, func() {

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage([]byte(marshal(*expectedEvent)))

		batchSize := 1
		batch := event.NewBatch(batchSize)

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
		message := kafkatest.NewMessage([]byte(marshal(*expectedEvent)))

		batchSize := 1
		batch := event.NewBatch(batchSize)

		Convey("When add is called with a valid message", func() {

			batch.Add(message)

			Convey("The batch contains the expected event.", func() {
				So(batch.Size(), ShouldEqual, 1)
				So(batch.Events()[0].Row, ShouldEqual, expectedEvent.Row)
				So(batch.Events()[0].InstanceID, ShouldEqual, expectedEvent.InstanceID)
			})
		})
	})
}

func TestCommit(t *testing.T) {

	Convey("Given a batch with two valid messages", t, func() {

		expectedEvent1 := event.ObservationExtracted{InstanceID: "123", Row: "the,row,content"}
		expectedEvent2 := event.ObservationExtracted{InstanceID: "123", Row: "the,second,content"}
		expectedEvent3 := event.ObservationExtracted{InstanceID: "123", Row: "the,third,content"}
		expectedEvent4 := event.ObservationExtracted{InstanceID: "123", Row: "the,fourth,content"}
		message1 := kafkatest.NewMessage([]byte(marshal(expectedEvent1)))
		message2 := kafkatest.NewMessage([]byte(marshal(expectedEvent2)))
		message3 := kafkatest.NewMessage([]byte(marshal(expectedEvent3)))
		message4 := kafkatest.NewMessage([]byte(marshal(expectedEvent4)))

		batchSize := 2
		batch := event.NewBatch(batchSize)

		batch.Add(message1)
		batch.Add(message2)

		batch.Commit()

		Convey("When commit is called", func() {

			Convey("The last message in the batch is committed", func() {
				So(message1.Committed(), ShouldBeFalse)
				So(message2.Committed(), ShouldBeTrue)
			})

			Convey("The batch is emptied.", func() {
				So(batch.IsEmpty(), ShouldBeTrue)
				So(batch.IsFull(), ShouldBeFalse)
				So(batch.Size(), ShouldEqual, 0)
			})

			Convey("The batch can be reused", func() {
				batch.Add(message3)

				So(batch.IsEmpty(), ShouldBeFalse)
				So(batch.IsFull(), ShouldBeFalse)
				So(batch.Size(), ShouldEqual, 1)

				So(batch.Events()[0].Row, ShouldEqual, expectedEvent3.Row)
				So(batch.Events()[0].InstanceID, ShouldEqual, expectedEvent3.InstanceID)

				batch.Add(message4)

				So(batch.IsEmpty(), ShouldBeFalse)
				So(batch.IsFull(), ShouldBeTrue)
				So(batch.Size(), ShouldEqual, 2)

				So(batch.Events()[1].Row, ShouldEqual, expectedEvent4.Row)
				So(batch.Events()[1].InstanceID, ShouldEqual, expectedEvent4.InstanceID)
			})
		})
	})
}

func TestSize(t *testing.T) {

	Convey("Given a batch", t, func() {

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage([]byte(marshal(*expectedEvent)))

		batchSize := 1
		batch := event.NewBatch(batchSize)

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
		message := kafkatest.NewMessage([]byte(marshal(*expectedEvent)))

		batchSize := 2
		batch := event.NewBatch(batchSize)

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
		message := kafkatest.NewMessage([]byte(marshal(*expectedEvent)))

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
func marshal(event event.ObservationExtracted) []byte {
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
