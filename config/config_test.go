package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/ONSdigital/dp-observation-importer/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {

	os.Clearenv()
	cfg, err := config.Get()

	Convey("Given an environment with no environment variables set", t, func() {

		Convey("When the config values are retrieved", func() {

			Convey("There should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, ":21700")
				So(cfg.Brokers[0], ShouldEqual, "localhost:9092")
				So(cfg.ObservationConsumerGroup, ShouldEqual, "observation-extracted")
				So(cfg.ObservationConsumerTopic, ShouldEqual, "observation-extracted")
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.BatchSize, ShouldEqual, 100)
				So(cfg.BatchWaitTime, ShouldEqual, time.Millisecond*200)
				So(cfg.ErrorProducerTopic, ShouldEqual, "report-events")
				So(cfg.ResultProducerTopic, ShouldEqual, "import-observations-inserted")
				So(cfg.CacheTTL, ShouldEqual, time.Minute*60)
				So(cfg.GracefulShutdownTimeout, ShouldEqual, time.Second*10)
				So(cfg.ServiceAuthToken, ShouldEqual, "Bearer AA78C45F-DD64-4631-BED9-FEAE29200620")
				So(cfg.ZebedeeURL, ShouldEqual, "http://localhost:8082")
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
			})
		})
	})
}
