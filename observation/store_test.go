package observation_test

import (
	"errors"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-observation-importer/observation/observationtest"
	"github.com/ONSdigital/dp-reporter-client/reporter/reportertest"
	neoErrors "github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/structures/messages"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var inputObservation = &observation.Observation{
	InstanceID: "123",
	Row:        "the,row,content",
	DimensionOptions: []*observation.DimensionOption{
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

		dbConnection := &observationtest.DBConnection{Result: observationtest.NewDBResult(1, 1, nil, nil)}
		errorReporterMock := reportertest.NewImportErrorReporterMock(nil)
		store := observation.NewStore(idCache, dbConnection, errorReporterMock)

		Convey("When save all is called", func() {

			results, err := store.SaveAll([]*observation.Observation{inputObservation})

			Convey("Then the DB is called with the expected query and parameters", func() {

				query := dbConnection.Query
				So(query, ShouldEqual, "UNWIND $rows AS row MATCH (`sex`:`_123_sex`), (`age`:`_123_age`) WHERE id(`sex`) = toInt(row.`sex`) AND id(`age`) = toInt(row.`age`) CREATE (o:`_123_observation` { value:row.v, rowIndex:row.i }), (o)-[:isValueOf]->(`sex`), (o)-[:isValueOf]->(`age`)")

				params := dbConnection.Params

				rows := params["rows"]
				row := rows.([]interface{})[0]
				rowMap, _ := row.(map[string]interface{})
				So(rowMap["v"], ShouldEqual, "the,row,content")
				So(rowMap["sex"], ShouldEqual, "333")
				So(rowMap["age"], ShouldEqual, "666")
			})

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

func TestStore_SaveAllExecError(t *testing.T) {
	Convey("Given a store with mock dimension ID cache and DB connection", t, func() {

		idCache := &observationtest.DimensionIDCache{IDs: ids}

		dbConnection := &observationtest.DBConnection{Result: nil, Error: mockError}
		errorReporterMock := reportertest.NewImportErrorReporterMock(nil)
		store := observation.NewStore(idCache, dbConnection, errorReporterMock)

		Convey("When dBConnection.Exec returns an error", func() {
			results, err := store.SaveAll([]*observation.Observation{inputObservation})

			Convey("Then no results and the expected error are returned", func() {
				So(results, ShouldBeNil)
				So(err, ShouldResemble, mockError)
			})

			Convey("And the error reporter is called once for each instance in the failed batch", func() {
				So(len(errorReporterMock.NotifyCalls()), ShouldEqual, 1)
				So(errorReporterMock.NotifyCalls()[0], ShouldResemble, reportertest.NotfiyParams{
					ID:         inputObservation.InstanceID,
					ErrContext: "observation batch insert failed",
					Err:        mockError,
				})
			})
		})
	})
}

func TestStore_SaveAll_ExecConstraintError(t *testing.T) {
	Convey("Given a store with mock dimension ID cache and DB connection", t, func() {

		idCache := &observationtest.DimensionIDCache{IDs: ids}

		constraintError := neoErrors.Wrap(messages.FailureMessage{Metadata: map[string]interface{}{"code": "Neo.ClientError.Schema.ConstraintValidationFailed"}}, "constraint error msg")

		dbConnection := &observationtest.DBConnection{Result: nil, Error: constraintError}
		errorReporterMock := reportertest.NewImportErrorReporterMock(nil)
		store := observation.NewStore(idCache, dbConnection, errorReporterMock)

		Convey("When dBConnection.Exec returns a neo4j constraint error", func() {
			results, err := store.SaveAll([]*observation.Observation{inputObservation})

			Convey("Then no results and a nil error are returned, as the error was ignored", func() {
				So(len(results), ShouldEqual, 0)
				So(err, ShouldEqual, nil)
			})

			Convey("And the error reporter is not called", func() {
				So(len(errorReporterMock.NotifyCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestStore_SaveAll_GetNodeIDError(t *testing.T) {

	Convey("Given a store with mock dimension ID cache that returns an error", t, func() {

		idCache := &observationtest.DimensionIDCache{IDs: nil, Error: mockError}

		errorReporterMock := reportertest.NewImportErrorReporterMock(nil)

		dbConnection := &observationtest.DBConnection{Result: observationtest.NewDBResult(1, 1, nil, nil)}
		store := observation.NewStore(idCache, dbConnection, errorReporterMock)

		Convey("When save all is called", func() {

			results, err := store.SaveAll([]*observation.Observation{inputObservation})

			Convey("The results have the expected values", func() {

				So(err, ShouldBeNil)
				So(results, ShouldNotBeNil)
				So(len(errorReporterMock.NotifyCalls()), ShouldEqual, 1)
				So(errorReporterMock.NotifyCalls()[0], ShouldResemble, reportertest.NotfiyParams{
					ID:         inputObservation.InstanceID,
					Err:        mockError,
					ErrContext: "failed to get dimension node id's",
				})

			})
		})
	})
}

func TestStore_SaveAll_NoNodeId(t *testing.T) {

	Convey("Given a store with mock dimension ID cache that has no entries", t, func() {

		idCache := &observationtest.DimensionIDCache{IDs: map[string]string{}}

		errorReporterMock := reportertest.NewImportErrorReporterMock(nil)

		dbConnection := &observationtest.DBConnection{Result: observationtest.NewDBResult(1, 1, nil, nil)}
		store := observation.NewStore(idCache, dbConnection, errorReporterMock)

		Convey("When save all is called", func() {

			results, err := store.SaveAll([]*observation.Observation{inputObservation})

			Convey("The results have the expected values", func() {

				So(err, ShouldBeNil)
				So(results, ShouldNotBeNil)
				So(len(errorReporterMock.NotifyCalls()), ShouldEqual, 1)
				So(errorReporterMock.NotifyCalls()[0], ShouldResemble, reportertest.NotfiyParams{
					ID:         inputObservation.InstanceID,
					Err:        errors.New("No nodeId found for 123_sex_Male"),
					ErrContext: "failed to create query parameters for batch query",
				})
			})
		})
	})
}
