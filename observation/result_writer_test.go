package observation_test

import (
	"context"
	"testing"

	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-kafka/kafkatest"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-observation-importer/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestResultWriter_Write(t *testing.T) {
	ctx := context.Background()

	Convey("Given an observation inserted event", t, func() {
		expectedInstanceID := "123abc"
		result := &observation.Result{ObservationsInserted: 2, InstanceID: expectedInstanceID}
		expectedEvent := observation.InsertedEvent{ObservationsInserted: 2, InstanceID: expectedInstanceID}

		pChannels := kafka.CreateProducerChannels()
		mockMessageProducer := kafkatest.NewMessageProducerWithChannels(&pChannels, false)
		observationMessageWriter := observation.NewResultWriter(mockMessageProducer)

		Convey("When write is called on the result writer", func() {

			Convey("The schema producer has the observation on its output channel", func(c C) {
				go func() {
					messageBytes := <-pChannels.Output

					observationEvent, err := Unmarshal(messageBytes)
					c.So(err, ShouldBeNil)
					c.So(observationEvent.InstanceID, ShouldEqual, expectedEvent.InstanceID)
					c.So(observationEvent.ObservationsInserted, ShouldEqual, expectedEvent.ObservationsInserted)

				}()
				observationMessageWriter.Write(ctx, []*observation.Result{result})

				// Closing channels, kafkatest producer channels should contain close function
				// which does this for us so we can do mockMessageProducer.Close()
				close(mockMessageProducer.Channels().Init)
				close(mockMessageProducer.Channels().Output)
				close(mockMessageProducer.Channels().Errors)
				close(mockMessageProducer.Channels().Closer)
				close(mockMessageProducer.Channels().Closed)
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

				actualEvent, err := Unmarshal(bytes)
				So(err, ShouldBeNil)
				So(actualEvent.InstanceID, ShouldEqual, expectedEvent.InstanceID)
				So(actualEvent.ObservationsInserted, ShouldEqual, expectedEvent.ObservationsInserted)
			})
		})
	})
}

// Unmarshal converts observation events to []byte.
func Unmarshal(bytes []byte) (*observation.InsertedEvent, error) {
	event := &observation.InsertedEvent{}
	err := schema.ObservationsInsertedEvent.Unmarshal(bytes, event)

	return event, err
}
