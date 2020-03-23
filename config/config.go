package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config values for the application.
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	Brokers                    []string      `envconfig:"KAFKA_ADDR"`
	ObservationConsumerGroup   string        `envconfig:"OBSERVATION_CONSUMER_GROUP"`
	ObservationConsumerTopic   string        `envconfig:"OBSERVATION_CONSUMER_TOPIC"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	BatchSize                  int           `envconfig:"BATCH_SIZE"`
	BatchWaitTime              time.Duration `envconfig:"BATCH_WAIT_TIME"`
	ErrorProducerTopic         string        `envconfig:"ERROR_PRODUCER_TOPIC"`
	ResultProducerTopic        string        `envconfig:"RESULT_PRODUCER_TOPIC"`
	KafkaMaxBytes              string        `envconfig:"KAFKA_MAX_BYTES"`
	CacheTTL                   time.Duration `envconfig:"CACHE_TTL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"              json:"-"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

// Get the configuration values from the environment or provide the defaults.
func Get() (*Config, error) {

	cfg := &Config{
		BindAddr:                   ":21700",
		Brokers:                    []string{"localhost:9092"},
		ObservationConsumerGroup:   "observation-extracted",
		ObservationConsumerTopic:   "observation-extracted",
		DatasetAPIURL:              "http://localhost:22000",
		BatchSize:                  1000,
		BatchWaitTime:              time.Millisecond * 200,
		ErrorProducerTopic:         "report-events",
		ResultProducerTopic:        "import-observations-inserted",
		KafkaMaxBytes:              "0",
		CacheTTL:                   time.Minute * 60,
		GracefulShutdownTimeout:    time.Second * 10,
		ServiceAuthToken:           "AA78C45F-DD64-4631-BED9-FEAE29200620",
		ZebedeeURL:                 "http://localhost:8082",
		HealthCheckInterval:        10 * time.Second,
		HealthCheckCriticalTimeout: 1 * time.Minute,
	}

	if err := envconfig.Process("", cfg); err != nil {
		return cfg, err
	}

	cfg.ServiceAuthToken = "Bearer " + cfg.ServiceAuthToken

	return cfg, nil
}
