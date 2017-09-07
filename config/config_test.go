package config_test

import (
	"github.com/ONSdigital/dp-observation-importer/config"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
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
				So(cfg.KafkaAddr[0], ShouldEqual, "localhost:9092")
				So(cfg.ObservationConsumerGroup, ShouldEqual, "observation-extracted")
				So(cfg.ObservationConsumerTopic, ShouldEqual, "observation-extracted")
				So(cfg.DatabaseAddress, ShouldEqual, "bolt://localhost:7687")
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.BatchSize, ShouldEqual, 1000)
				So(cfg.BatchWaitTime, ShouldEqual, time.Millisecond*200)
				So(cfg.ErrorProducerTopic, ShouldEqual, "import-error")
				So(cfg.ResultProducerTopic, ShouldEqual, "import-observations-inserted")
				So(cfg.CacheTTL, ShouldEqual, time.Minute*60)
				So(cfg.GracefulShutdownTimeout, ShouldEqual, time.Second*10)
			})
		})
	})
}
