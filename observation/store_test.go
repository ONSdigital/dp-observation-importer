package observation_test

import (
	"errors"
	"testing"

	"github.com/ONSdigital/dp-graph/graph"
	"github.com/ONSdigital/dp-observation-importer/models"
	store "github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-observation-importer/observation/observationtest"
	"github.com/ONSdigital/dp-reporter-client/reporter/reportertest"
	. "github.com/smartystreets/goconvey/convey"
)

var inputObservation = &models.Observation{
	InstanceID: "123",
	Row:        "the,row,content",
	DimensionOptions: []*models.DimensionOption{
		{DimensionName: "Sex", Name: "Male"},
		{DimensionName: "Age", Name: "45"},
	},
}

var ids = map[string]string{
	"123_sex_Male": "333",
	"123_age_45":   "666",
}

var mockError = errors.New("Broken")

func TestStore_SaveAll(t *testing.T) {

	Convey("Given a store with mock dimension ID cache and DB connection", t, func() {

		idCache := &observationtest.DimensionIDCache{IDs: ids}

		db := graph.Test(true, true, true)

		errorReporterMock := reportertest.NewImportErrorReporterMock(nil)
		s := store.NewStore(idCache, db, errorReporterMock)

		Convey("When save all is called", func() {

			results, err := s.SaveAll([]*models.Observation{inputObservation})

			Convey("The results have the expected values", func() {

				So(err, ShouldBeNil)
				So(results, ShouldNotBeNil)
				So(len(results), ShouldEqual, 1)
				So(results[0].InstanceID, ShouldEqual, inputObservation.InstanceID)
				So(results[0].ObservationsInserted, ShouldEqual, 1)
			})

			Convey("And the error reporter is never called", func() {
				So(len(errorReporterMock.NotifyCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestStore_SaveAll_GetNodeIDError(t *testing.T) {

	Convey("Given a store with mock dimension ID cache that returns an error", t, func() {

		idCache := &observationtest.DimensionIDCache{IDs: nil, Error: mockError}

		errorReporterMock := reportertest.NewImportErrorReporterMock(nil)

		db := graph.Test(true, true, true)

		s := store.NewStore(idCache, db, errorReporterMock)

		Convey("When save all is called", func() {

			results, err := s.SaveAll([]*models.Observation{inputObservation})

			Convey("The results have the expected values", func() {

				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, mockError)

				So(results, ShouldNotBeNil)
				So(len(errorReporterMock.NotifyCalls()), ShouldEqual, 1)
				So(errorReporterMock.NotifyCalls()[0], ShouldResemble, reportertest.NotfiyParams{
					ID:         inputObservation.InstanceID,
					Err:        mockError,
					ErrContext: "failed to get dimension node id's for batch",
				})

			})
		})
	})
}
