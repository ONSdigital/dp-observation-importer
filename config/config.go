package config

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config values for the application.
type Config struct {
	BindAddr                 string        `envconfig:"BIND_ADDR"`
	KafkaAddr                []string      `envconfig:"KAFKA_ADDR"`
	ObservationConsumerGroup string        `envconfig:"OBSERVATION_CONSUMER_GROUP"`
	ObservationConsumerTopic string        `envconfig:"OBSERVATION_CONSUMER_TOPIC"`
	DatabaseAddress          string        `envconfig:"DATABASE_ADDRESS"`
	DatasetAPIURL            string        `envconfig:"DATASET_API_URL"`
	Neo4jPoolSize            int           `envconfig:"NEO4J_POOL_SIZE"`
	DatasetAPIAuthToken      string        `envconfig:"DATASET_API_AUTH_TOKEN"          json:"-"`
	BatchSize                int           `envconfig:"BATCH_SIZE"`
	BatchWaitTime            time.Duration `envconfig:"BATCH_WAIT_TIME"`
	ErrorProducerTopic       string        `envconfig:"ERROR_PRODUCER_TOPIC"`
	ResultProducerTopic      string        `envconfig:"RESULT_PRODUCER_TOPIC"`
	CacheTTL                 time.Duration `envconfig:"CACHE_TTL"`
	GracefulShutdownTimeout  time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	ServiceAuthToken         string        `envconfig:"SERVICE_AUTH_TOKEN"             json:"-"`
}

// Get the configuration values from the environment or provide the defaults.
func Get() (*Config, error) {

	cfg := &Config{
		BindAddr:                 ":21700",
		KafkaAddr:                []string{"localhost:9092"},
		ObservationConsumerGroup: "observation-extracted",
		ObservationConsumerTopic: "observation-extracted",
		DatabaseAddress:          "bolt://localhost:7687",
		DatasetAPIURL:            "http://localhost:22000",
		Neo4jPoolSize:            5,
		DatasetAPIAuthToken:      "FD0108EA-825D-411C-9B1D-41EF7727F465",
		BatchSize:                1000,
		BatchWaitTime:            time.Millisecond * 200,
		ErrorProducerTopic:       "report-events",
		ResultProducerTopic:      "import-observations-inserted",
		CacheTTL:                 time.Minute * 60,
		GracefulShutdownTimeout:  time.Second * 10,
		ServiceAuthToken:         "AA78C45F-DD64-4631-BED9-FEAE29200620",
	}

	return cfg, envconfig.Process("", cfg)
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
