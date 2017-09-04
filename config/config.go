package config

import (
	"time"

	"github.com/ian-kent/gofigure"
)

// Config values for the application.
type Config struct {
	BindAddr                 string        `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"The port to bind to"`
	KafkaAddr                []string      `env:"KAFKA_ADDR" flag:"kafka-addr" flagDesc:"The addresses of Kafka instances"`
	ObservationConsumerGroup string        `env:"OBSERVATION_CONSUMER_GROUP" flag:"observation-consumer-group" flagDesc:"The Kafka consumer group to consume observation messages from"`
	ObservationConsumerTopic string        `env:"OBSERVATION_CONSUMER_TOPIC" flag:"observation-consumer-topic" flagDesc:"The Kafka topic to consume observation messages from"`
	DatabaseAddress          string        `env:"DATABASE_ADDRESS" flag:"database-address" flagDesc:"The address of the database to store observations"`
	ImportAPIURL             string        `env:"IMPORT_API_URL" flag:"import-api-url" flagDesc:"The URL of the import API"`
	BatchSize                int           `env:"BATCH_SIZE" flag:"batch-size" flagDesc:"The number of messages to process in each batch"`
	BatchWaitTime            time.Duration `env:"BATCH_WAIT_TIME_MS" flag:"batch-wait-time-ms" flagDesc:"The number of MS to wait before processing a partially full batch of messages"`
	ErrorProducerTopic       string        `env:"ERROR_PRODUCER_TOPIC" flag:"error-producer-topic" flagDesc:"The Kafka topic to send error messages to"`
	ResultProducerTopic      string        `env:"RESULT_PRODUCER_TOPIC" flag:"result-producer-topic" flagDesc:"The Kafka topic to send result messages to"`
	CacheTTL                 time.Duration `env:"CACHE_TTL" flag:"cache-clear-time" flagDesc:"The amount of time to wait before clearing the cache (In minutes)"`
}

// Get the configuration values from the environment or provide the defaults.
func Get() (*Config, error) {

	cfg := Config{
		BindAddr:                 ":21700",
		KafkaAddr:                []string{"localhost:9092"},
		ObservationConsumerGroup: "observation-extracted",
		ObservationConsumerTopic: "observation-extracted",
		DatabaseAddress:          "bolt://localhost:7687",
		ImportAPIURL:             "http://localhost:21800",
		BatchSize:                1000,
		BatchWaitTime:            time.Millisecond * 200,
		ErrorProducerTopic:       "event-reporter",
		ResultProducerTopic:      "import-observations-inserted",
		CacheTTL:                 time.Minute * 60,
	}

	err := gofigure.Gofigure(&cfg)

	return &cfg, err
}
