package observation_test

import (
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-observation-importer/observation/observationtest"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func TestMapper_Map(t *testing.T) {

	Convey("Given a mapper with a mock header cache", t, func() {

		csvHeader := strings.Split("V4_1,some other header,Time_codelist,Time,Geography_codelist,Geography,Aggregate_codelist,Aggregate", ",")
		mockOrderCache := observationtest.DimensionHeaderCache{DimensionOrder: csvHeader, Error: nil}
		mapper := observation.NewMapper(mockOrderCache)

		Convey("When map is called with an example csv line", func() {

			csvRow := "128,,Month,Aug-16,K02000001,,cpi1dim1A0,CPI (overall index)"
			rowIndex := int64(453)
			instanceID := "123321"
			observation, err := mapper.Map(csvRow, rowIndex, instanceID)

			Convey("The returned observation should be populated with the row data.", func() {
				So(err, ShouldBeNil)
				So(observation, ShouldNotBeNil)
				So(observation.InstanceID, ShouldEqual, instanceID)
				So(observation.Row, ShouldEqual, csvRow)
				So(observation.RowIndex, ShouldEqual, rowIndex)

				So(len(observation.DimensionOptions), ShouldEqual, 3)
				So(observation.DimensionOptions[0].DimensionName, ShouldEqual, "Time")
				So(observation.DimensionOptions[0].Name, ShouldEqual, "Aug-16")
				So(observation.DimensionOptions[1].DimensionName, ShouldEqual, "Geography")
				So(observation.DimensionOptions[1].Name, ShouldEqual, "K02000001")
				So(observation.DimensionOptions[2].DimensionName, ShouldEqual, "Aggregate")
				So(observation.DimensionOptions[2].Name, ShouldEqual, "cpi1dim1A0")
			})
		})
	})
}
