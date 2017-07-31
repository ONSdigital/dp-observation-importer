package dimension

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/ONSdigital/dp-observation-importer/dimension/dimensiontest"
)

func TestDimensionMemoryCache_GetNodeIDs(t *testing.T) {

	cache := NewIDCache(dimensiontest.MockIDStore{})

	Convey("Given a valid instanceId", t, func() {

		Convey("When a request for a cached dimensions", func() {
			dimensions, error := cache.GetNodeIDs("1")
			Convey("The dimensions are returned", func() {
				So(error, ShouldBeNil)
				So(dimensions["age_55"], ShouldEqual, "123")
			})
		})
	})
}

func TestDimensionMemoryCache_GetNodeIDsReturnsError(t *testing.T) {

	cache := NewIDCache(dimensiontest.MockIDStore{ReturnError:true})

	Convey("Given a invalid instanceId", t, func() {

		Convey("When a request for a cached dimensions", func() {
			_, error := cache.GetNodeIDs("1")
			Convey("An error is returned", func() {
				So(error, ShouldNotBeNil)

			})
		})
	})
}
