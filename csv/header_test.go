package csv_test

import (
	"github.com/ONSdigital/dp-observation-importer/csv"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestHeader_DimensionOffset_0(t *testing.T) {

	Convey("Given an example CSV header with a zero offset", t, func() {

		header := csv.NewHeader([]string{"V4_0"})
		So(header, ShouldNotBeNil)

		Convey("When DimensionOffset is called", func() {

			offset, err := header.DimensionOffset()

			Convey("1 is returned - accounting for the observation column", func() {
				So(err, ShouldBeNil)
				So(offset, ShouldEqual, 1)
			})
		})
	})
}

func TestHeader_DimensionOffset_1(t *testing.T) {

	Convey("Given an example CSV header with a zero offset", t, func() {

		header := csv.NewHeader([]string{"V4_1"})
		So(header, ShouldNotBeNil)

		Convey("When DimensionOffset is called", func() {

			offset, err := header.DimensionOffset()

			Convey("2 is returned - accounting for the observation column", func() {
				So(err, ShouldBeNil)
				So(offset, ShouldEqual, 2)
			})
		})
	})
}

func TestHeader_DimensionOffset_3(t *testing.T) {

	Convey("Given an example CSV header with a zero offset", t, func() {

		header := csv.NewHeader([]string{"V4_3"})
		So(header, ShouldNotBeNil)

		Convey("When DimensionOffset is called", func() {

			offset, err := header.DimensionOffset()

			Convey("4 is returned - accounting for the observation column", func() {
				So(err, ShouldBeNil)
				So(offset, ShouldEqual, 4)
			})
		})
	})
}

func TestHeader_DimensionOffset_Error(t *testing.T) {

	Convey("Given an example CSV header with an invalid offset", t, func() {

		header := csv.NewHeader([]string{"V4_nooffset"})
		So(header, ShouldNotBeNil)

		Convey("When DimensionOffset is called", func() {

			_, err := header.DimensionOffset()

			Convey("An error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
