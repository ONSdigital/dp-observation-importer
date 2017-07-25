package observation_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-observation-importer/observation/observationtest"
	"github.com/ONSdigital/dp-observation-importer/dimension"
)

func TestSpec(t *testing.T) {

	Convey("Given a store with mock dimension ID cache and DB connection", t, func() {

		inputObservation := &observation.Observation{
			InstanceID: "123",
			Row:        "the,row,content",
			DimensionOptions: []observation.DimensionOption{
				{DimensionName: "_123_Sex", NodeID: "333", NodeAlias: "Sex" },
				{DimensionName: "_123_Age", NodeID: "444", NodeAlias: "Age" },
			},
		}

		ids := dimension.IDs{
			//dimension.Dimension{"Sex", "Male", "333"},
			//dimension.Dimension{"Age", "Old", "444"},
		}

		idCache := &observationtest.DimensionIDCache{IDs: ids }
		dbConnection := &observationtest.DBConnection{}
		store := observation.NewStore(idCache, dbConnection)

		Convey("When save all is called", func() {

			store.SaveAll([]*observation.Observation{inputObservation })

			Convey("Then the DB is called with the expected query and parameters", func() {

				So(len(dbConnection.Queries), ShouldEqual, 1)
				So(len(dbConnection.Params), ShouldEqual, 1)

				query := dbConnection.Queries[0]
				So(query, ShouldEqual, "UNWIND $rows AS row MATCH (Sex:_123_Sex), (Age:_123_Age) WHERE id(Sex) = toInt(row._123_Sex) AND id(Age) = toInt(row._123_Age) CREATE (o:_123_observation { value:row.v }), (o)-[:isValueOf]->(Sex), (o)-[:isValueOf]->(Age)")

				//params := dbConnection.Params[0]
				//rows := params["rows"]
				//rowsMap, _ := rows.(map[string]interface{})
				//So(rowsMap["v"], ShouldEqual, "the,row,content")
				//So(rowsMap["Sex"], ShouldEqual, "333")
				//So(rowsMap["Age"], ShouldEqual, "666")
			})
		})
	})
}
