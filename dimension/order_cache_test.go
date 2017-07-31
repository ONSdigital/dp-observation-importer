package dimension

import (
	"github.com/ONSdigital/dp-observation-importer/dimension/dimensiontest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestHeaderCache_GetOrder(t *testing.T) {

	cache := NewOrderCache(dimensiontest.MockOrderStore{}, time.Minute)

	Convey("Given a valid instanceId", t, func() {

		Convey("When a request for a cached CSV header", func() {
			headers, error := cache.GetOrder("1")
			Convey("The CSV headers is returned", func() {
				So(error, ShouldBeNil)
				So(headers, ShouldContain, "V4_1")
				So(headers, ShouldContain, "data_marking")
				So(headers, ShouldContain, "time_codelist")
			})
		})
	})
}

func TestHeaderCache_GetOrder_ReturnErrors(t *testing.T) {

	cache := NewOrderCache(dimensiontest.MockOrderStore{ReturnError: true}, time.Minute)

	Convey("Given a invalid instanceId", t, func() {

		Convey("When a request for the CSV header", func() {
			headers, error := cache.GetOrder("1")
			Convey("An error is returned", func() {
				So(error, ShouldNotBeNil)
				So(headers, ShouldBeEmpty)
			})
		})
	})
}
