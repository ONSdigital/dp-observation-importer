package config_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/ONSdigital/dp-observation-importer/config"
)

func TestSpec(t *testing.T) {

	cfg, err := config.Get()

	Convey("Given an environment with no environment variables set", t, func() {

		Convey("When the config values are retrieved", func() {

			Convey("There should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, ":21700")
				So(cfg.KafkaAddr, ShouldEqual, "http://localhost:9092")
				So(cfg.ObservationConsumerGroup, ShouldEqual, "observation-extracted")
				So(cfg.ObservationConsumerTopic, ShouldEqual, "observation-extracted")
				So(cfg.DatabaseAddress, ShouldEqual, "bolt://localhost:7687")
				So(cfg.ImportAPIURL, ShouldEqual, "http://localhost:21800")
			})
		})
	})
}