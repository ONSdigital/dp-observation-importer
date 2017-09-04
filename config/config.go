package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

// Config values for the application.
type Config struct {
	BindAddr                 string        `envconfig:"BIND_ADDR"`
	KafkaAddr                []string      `envconfig:"KAFKA_ADDR"`
	ObservationConsumerGroup string        `envconfig:"OBSERVATION_CONSUMER_GROUP"`
	ObservationConsumerTopic string        `envconfig:"OBSERVATION_CONSUMER_TOPIC"`
	DatabaseAddress          string        `envconfig:"DATABASE_ADDRESS"`
	ImportAPIURL             string        `envconfig:"IMPORT_API_URL"`
	BatchSize                int           `envconfig:"BATCH_SIZE"`
	BatchWaitTime            time.Duration `envconfig:"BATCH_WAIT_TIME_MS"`
	ErrorProducerTopic       string        `envconfig:"ERROR_PRODUCER_TOPIC"`
	ResultProducerTopic      string        `envconfig:"RESULT_PRODUCER_TOPIC"`
	CacheTTL                 time.Duration `envconfig:"CACHE_TTL"`
}

// Get the configuration values from the environment or provide the defaults.
func Get() (*Config, error) {

	cfg := &Config{
		BindAddr:                 ":21700",
		KafkaAddr:                []string{"localhost:9092"},
		ObservationConsumerGroup: "observation-extracted",
		ObservationConsumerTopic: "observation-extracted",
		DatabaseAddress:          "bolt://localhost:7687",
		ImportAPIURL:             "http://localhost:21800",
		BatchSize:                1000,
		BatchWaitTime:            time.Millisecond * 200,
		ErrorProducerTopic:       "import-error",
		ResultProducerTopic:      "import-observations-inserted",
		CacheTTL:                 time.Minute * 60,
	}

	return cfg, envconfig.Process("", cfg)
}
