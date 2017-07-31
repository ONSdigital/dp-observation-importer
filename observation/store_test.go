package observation_test

import (
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-observation-importer/observation/observationtest"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSpec(t *testing.T) {

	Convey("Given a store with mock dimension ID cache and DB connection", t, func() {

		inputObservation := &observation.Observation{
			InstanceID: "123",
			Row:        "the,row,content",
			DimensionOptions: []observation.DimensionOption{
				{DimensionName: "Sex", Name: "Male"},
				{DimensionName: "Age", Name: "45"},
			},
		}

		ids := map[string]string{
			"123_Sex_Male": "333",
			"123_Age_45":   "666",
		}
		idCache := &observationtest.DimensionIDCache{IDs: ids}

		dbConnection := &observationtest.DBConnection{Results: []bolt.Result{observationtest.NewDBResult(1, 1, nil, nil)}}
		store := observation.NewStore(idCache, dbConnection)

		Convey("When save all is called", func() {

			results, err := store.SaveAll([]*observation.Observation{inputObservation})

			Convey("Then the DB is called with the expected query and parameters", func() {

				So(len(dbConnection.Queries), ShouldEqual, 1)
				So(len(dbConnection.Params), ShouldEqual, 1)

				query := dbConnection.Queries[0]
				So(query, ShouldEqual, "UNWIND $rows AS row MATCH (Sex:_123_Sex), (Age:_123_Age) WHERE id(Sex) = toInt(row.Sex) AND id(Age) = toInt(row.Age) CREATE (o:_123_observation { value:row.v }), (o)-[:isValueOf]->(Sex), (o)-[:isValueOf]->(Age)")

				//var params map[string]interface{}
				params := dbConnection.Params[0]

				rows := params["rows"]
				row := rows.([]interface{})[0]
				rowMap, _ := row.(map[string]interface{})
				So(rowMap["v"], ShouldEqual, "the,row,content")
				So(rowMap["Sex"], ShouldEqual, "333")
				So(rowMap["Age"], ShouldEqual, "666")
			})

			Convey("The results have the expected values", func() {

				So(err, ShouldBeNil)
				So(results, ShouldNotBeNil)
				So(len(results), ShouldEqual, 1)
				So(results[0].InstanceID, ShouldEqual, inputObservation.InstanceID)
				So(results[0].ObservationsInserted, ShouldEqual, 1)
			})
		})
	})
}
