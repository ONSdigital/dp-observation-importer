package event_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/event/eventtest"
	"github.com/ONSdigital/dp-observation-importer/models"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-reporter-client/reporter/reportertest"
	. "github.com/smartystreets/goconvey/convey"
)

// var ctx = context.Background()
var mockError = errors.New("Mapping failed")

var instanceID = "123321"
var row = "31749353,Quarter,May-Jul 2016,,,,In Employment,,16+,,People,,Non Seasonal Adjusted"

var expectedEvent = &event.ObservationExtracted{
	InstanceID: instanceID,
	Row:        row,
	RowIndex:   int64(765),
}

var expectedObservation = &models.Observation{
	InstanceID: instanceID,
	Row:        row,
	DimensionOptions: []*models.DimensionOption{
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
	ctx := context.Background()

	Convey("Given a handler configured with a mock mapper and store", t, func() {

		mockObservationMapper := &eventtest.ObservationMapperMock{
			MapFunc: func(ctx context.Context, row string, rowIndex int64, instanceID string) (*models.Observation, error) {
				return expectedObservation, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			SaveAllFunc: func(ctx context.Context, observations []*models.Observation) ([]*observation.Result, error) {
				return []*observation.Result{expectedResult}, nil
			},
		}

		mockResultWriter := &eventtest.ResultWriterMock{
			WriteFunc: func(ctx context.Context, results []*observation.Result) {},
		}

		handler := event.NewBatchHandler(mockObservationMapper, mockObservationStore, mockResultWriter, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(ctx, []*event.ObservationExtracted{expectedEvent})

			Convey("The expected calls to the observation mapper, store, and result writer happen", func() {
				So(err, ShouldBeNil)

				So(len(mockObservationMapper.MapCalls()), ShouldEqual, 1)
				So(mockObservationMapper.MapCalls()[0].Row, ShouldEqual, expectedEvent.Row)
				So(mockObservationMapper.MapCalls()[0].RowIndex, ShouldEqual, expectedEvent.RowIndex)
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
	ctx := context.Background()

	Convey("Given a handler configured with a mock mapper that returns an error", t, func() {

		mockObservationMapper := &eventtest.ObservationMapperMock{
			MapFunc: func(ctx context.Context, row string, rowIndex int64, instanceID string) (*models.Observation, error) {
				return nil, mockError
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			SaveAllFunc: func(ctx context.Context, observations []*models.Observation) ([]*observation.Result, error) {
				return []*observation.Result{expectedResult}, nil
			},
		}

		mockResultWriter := &eventtest.ResultWriterMock{
			WriteFunc: func(ctx context.Context, results []*observation.Result) {},
		}

		mockErrorReporter := reportertest.NewImportErrorReporterMock(nil)

		handler := event.NewBatchHandler(mockObservationMapper, mockObservationStore, mockResultWriter, mockErrorReporter)

		Convey("When handle is called", func() {

			err := handler.Handle(ctx, []*event.ObservationExtracted{expectedEvent})

			Convey("The error from the mapper is sent to the error handler", func() {
				So(err, ShouldBeNil)
				So(len(mockErrorReporter.NotifyCalls()), ShouldEqual, 1)
				So(mockErrorReporter.NotifyCalls()[0], ShouldResemble, reportertest.NotfiyParams{
					ID:         instanceID,
					ErrContext: "error while attempting to convert from row data to observation instances",
					Err:        mockError,
				})
			})
		})
	})
}

func TestBatchHandler_Handle_StoreError(t *testing.T) {
	ctx := context.Background()

	Convey("Given a handler configured with a mock mapper, and mock store that returns an error", t, func() {

		mockObservationMapper := &eventtest.ObservationMapperMock{
			MapFunc: func(ctx context.Context, row string, rowIndex int64, instanceID string) (*models.Observation, error) {
				return expectedObservation, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			SaveAllFunc: func(ctx context.Context, observations []*models.Observation) ([]*observation.Result, error) {
				return nil, mockError
			},
		}

		handler := event.NewBatchHandler(mockObservationMapper, mockObservationStore, nil, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(ctx, []*event.ObservationExtracted{expectedEvent})

			Convey("The error returned from the store is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, mockError)
			})
		})
	})
}
