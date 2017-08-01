package observation_test

import (
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestResultWriter_Write(t *testing.T) {

	Convey("Given an observation inserted event", t, func() {

		expectedInstanceID := "123abc"
		result := &observation.Result{ObservationsInserted: 2, InstanceID: expectedInstanceID}
		expectedEvent := observation.InsertedEvent{ObservationsInserted: 2, InstanceID: expectedInstanceID}

		// mock schema producer contains the output channel to capture messages sent.
		outputChannel := make(chan []byte, 1)
		mockMessageProducer := kafkatest.NewMessageProducer(outputChannel, nil, nil)

		observationMessageWriter := observation.NewResultWriter(mockMessageProducer)

		Convey("When write is called on the result writer", func() {

			observationMessageWriter.Write([]*observation.Result{result})

			Convey("The schema producer has the observation on its output channel", func() {

				messageBytes := <-outputChannel
				close(outputChannel)
				observationEvent := Unmarshal(messageBytes)
				So(observationEvent.InstanceID, ShouldEqual, expectedEvent.InstanceID)
				So(observationEvent.ObservationsInserted, ShouldEqual, expectedEvent.ObservationsInserted)
			})
		})
	})
}

func TestMessageWriter_Marshal(t *testing.T) {

	Convey("Given an example observation inserted event", t, func() {

		expectedInstanceID := "123abc"
		expectedEvent := observation.InsertedEvent{ObservationsInserted: 2, InstanceID: expectedInstanceID}

		Convey("When Marshal is called", func() {

			bytes, err := observation.Marshal(expectedEvent)
			So(err, ShouldBeNil)

			Convey("The observation inserted event can be unmarshalled and has the expected values", func() {

				actualEvent := Unmarshal(bytes)
				So(actualEvent.InstanceID, ShouldEqual, expectedEvent.InstanceID)
				So(actualEvent.ObservationsInserted, ShouldEqual, expectedEvent.ObservationsInserted)
			})
		})
	})
}

// Unmarshal converts observation events to []byte.
func Unmarshal(bytes []byte) *observation.InsertedEvent {
	event := &observation.InsertedEvent{}
	err := schema.ObservationsInsertedEvent.Unmarshal(bytes, event)
	So(err, ShouldBeNil)
	return event
}
