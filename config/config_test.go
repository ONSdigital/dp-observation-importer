package config

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {

	Convey("Given an environment with no environment variables set", t, func() {

		os.Clearenv()

		Convey("When the default config is retrieved", func() {
			cfg, err := Get(context.Background())

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, ":21700")
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.CacheTTL, ShouldEqual, time.Minute*60)
				So(cfg.GracefulShutdownTimeout, ShouldEqual, time.Second*10)
				So(cfg.ServiceAuthToken, ShouldEqual, "Bearer AA78C45F-DD64-4631-BED9-FEAE29200620")
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.GraphDriverChoice, ShouldEqual, "neo4j")
				So(cfg.EnableGetGraphDimensionID, ShouldBeTrue)

				So(cfg.KafkaConfig.Brokers[0], ShouldEqual, "localhost:9092")
				So(cfg.KafkaConfig.ObservationConsumerGroup, ShouldEqual, "observation-extracted")
				So(cfg.KafkaConfig.ObservationConsumerTopic, ShouldEqual, "observation-extracted")
				So(cfg.KafkaConfig.BatchSize, ShouldEqual, 100)
				So(cfg.KafkaConfig.BatchWaitTime, ShouldEqual, time.Millisecond*200)
				So(cfg.KafkaConfig.ErrorProducerTopic, ShouldEqual, "report-events")
				So(cfg.KafkaConfig.ResultProducerTopic, ShouldEqual, "import-observations-inserted")
				So(cfg.KafkaConfig.OffsetOldest, ShouldBeTrue)
				So(cfg.KafkaConfig.SecProtocol, ShouldEqual, "")
				So(cfg.KafkaConfig.SecClientKey, ShouldEqual, "")
				So(cfg.KafkaConfig.SecClientCert, ShouldEqual, "")
				So(cfg.KafkaConfig.SecCACerts, ShouldEqual, "")
				So(cfg.KafkaConfig.SecSkipVerify, ShouldBeFalse)

			})
		})

		Convey("When configuration is called with an invalid security setting", func() {
			os.Setenv("KAFKA_SEC_PROTO", "ssl")
			cfg, err := Get(context.Background())

			Convey("Then an error should be returned", func() {
				So(cfg, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("config validation errors: KAFKA_SEC_PROTO has invalid value"))
			})
		})

		Convey("When more than one config is invalid", func() {
			os.Setenv("SERVICE_AUTH_TOKEN", "")
			os.Setenv("KAFKA_SEC_CLIENT_CERT", "notempty")
			cfg, err := Get(context.Background())

			Convey("Then an error should be returned", func() {
				So(cfg, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("config validation errors: no SERVICE_AUTH_TOKEN given, only one of KAFKA_SEC_CLIENT_CERT or KAFKA_SEC_CLIENT_KEY has been set - requires both"))
			})
		})

	})
}
