package event_test

import (
	"testing"

	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/event/eventtest"
	"github.com/ONSdigital/dp-observation-importer/observation"
	mock "github.com/ONSdigital/go-ns/errorhandler/mock"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var mockError = errors.New("Mapping failed")

var instanceID = "123321"
var row = "31749353,Quarter,May-Jul 2016,,,,In Employment,,16+,,People,,Non Seasonal Adjusted"

var expectedEvent = &event.ObservationExtracted{
	InstanceID: instanceID,
	Row:        row,
}

var expectedObservation = &observation.Observation{
	InstanceID: instanceID,
	Row:        row,
	DimensionOptions: []*observation.DimensionOption{
		{
			Name:          "Male",
			DimensionName: "Gender",
		},
		{
			Name:          "45",
			DimensionName: "Age",
		},
	},
}

var expectedResult = &observation.Result{
	InstanceID:           instanceID,
	ObservationsInserted: 1,
}

func TestBatchHandler_Handle(t *testing.T) {

	Convey("Given a handler configured with a mock mapper and store", t, func() {

		mockObservationMapper := &eventtest.ObservationMapperMock{
			MapFunc: func(row string, instanceID string) (*observation.Observation, error) {
				return expectedObservation, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			SaveAllFunc: func(observations []*observation.Observation) ([]*observation.Result, error) {
				return []*observation.Result{expectedResult}, nil
			},
		}

		mockResultWriter := &eventtest.ResultWriterMock{
			WriteFunc: func(results []*observation.Result) {},
		}

		handler := event.NewBatchHandler(mockObservationMapper, mockObservationStore, mockResultWriter, nil)

		Convey("When handle is called", func() {

			err := handler.Handle([]*event.ObservationExtracted{expectedEvent})

			Convey("The expected calls to the observation mapper, store, and result writer happen", func() {
				So(err, ShouldBeNil)

				So(len(mockObservationMapper.MapCalls()), ShouldEqual, 1)
				So(mockObservationMapper.MapCalls()[0].Row, ShouldEqual, expectedEvent.Row)
				So(mockObservationMapper.MapCalls()[0].InstanceID, ShouldEqual, expectedEvent.InstanceID)

				So(len(mockObservationStore.SaveAllCalls()), ShouldEqual, 1)
				So(mockObservationStore.SaveAllCalls()[0].Observations[0], ShouldEqual, expectedObservation)

				So(len(mockResultWriter.WriteCalls()), ShouldEqual, 1)
				So(mockResultWriter.WriteCalls()[0].Results[0], ShouldEqual, expectedResult)
			})
		})
	})
}

func TestBatchHandler_Handle_MapperError(t *testing.T) {

	Convey("Given a handler configured with a mock mapper that returns an error", t, func() {

		mockObservationMapper := &eventtest.ObservationMapperMock{
			MapFunc: func(row string, instanceID string) (*observation.Observation, error) {
				return nil, mockError
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			SaveAllFunc: func(observations []*observation.Observation) ([]*observation.Result, error) {
				return []*observation.Result{expectedResult}, nil
			},
		}

		mockResultWriter := &eventtest.ResultWriterMock{
			WriteFunc: func(results []*observation.Result) {},
		}

		mockErrorHandler := &mock.HandlerMock{
			HandleFunc: func(instanceID string, err error) {},
		}

		handler := event.NewBatchHandler(mockObservationMapper, mockObservationStore, mockResultWriter, mockErrorHandler)

		Convey("When handle is called", func() {

			err := handler.Handle([]*event.ObservationExtracted{expectedEvent})

			Convey("The error from the mapper is sent to the error handler", func() {
				So(err, ShouldBeNil)
				So(len(mockErrorHandler.HandleCalls()), ShouldEqual, 1)
				So(mockErrorHandler.HandleCalls()[0].InstanceID, ShouldEqual, instanceID)
			})
		})
	})
}

func TestBatchHandler_Handle_StoreError(t *testing.T) {

	Convey("Given a handler configured with a mock mapper, and mock store that returns an error", t, func() {

		mockObservationMapper := &eventtest.ObservationMapperMock{
			MapFunc: func(row string, instanceID string) (*observation.Observation, error) {
				return expectedObservation, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			SaveAllFunc: func(observations []*observation.Observation) ([]*observation.Result, error) {
				return nil, mockError
			},
		}

		handler := event.NewBatchHandler(mockObservationMapper, mockObservationStore, nil, nil)

		Convey("When handle is called", func() {

			err := handler.Handle([]*event.ObservationExtracted{expectedEvent})

			Convey("The error returned from the store is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, mockError)
			})
		})
	})
}
