package config

import (
	"github.com/ian-kent/gofigure"
)

// Config values for the application.
type Config struct {
	BindAddr                 string `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"The port to bind to"`
	KafkaAddr                string `env:"KAFKA_ADDR" flag:"kafka-addr" flagDesc:"The address of the Kafka instance"`
	ObservationConsumerGroup string `env:"OBSERVATION_CONSUMER_GROUP" flag:"observation-consumer-group" flagDesc:"The Kafka consumer group to consume observation messages from"`
	ObservationConsumerTopic string `env:"OBSERVATION_CONSUMER_TOPIC" flag:"observation-consumer-topic" flagDesc:"The Kafka topic to consume observation messages from"`
	DatabaseAddress          string `env:"DATABASE_ADDRESS" flag:"database-address" flagDesc:"The address of the database to store observations"`
	ImportAPIURL             string `env:"IMPORT_API_URL" flag:"import-api-url" flagDesc:"The URL of the import API"`
}

// Get the configuration values from the environment or provide the defaults.
func Get() (*Config, error) {

	cfg := Config{
		BindAddr:                 ":21700",
		KafkaAddr:                "http://localhost:9092",
		ObservationConsumerGroup: "observation-extracted",
		ObservationConsumerTopic: "observation-extracted",
		DatabaseAddress:          "bolt://localhost:7687",
		ImportAPIURL:             "http://localhost:21800",
	}

	err := gofigure.Gofigure(&cfg)

	return &cfg, err
}
