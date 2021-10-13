package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/kelseyhightower/envconfig"
)

// KafkaTLSProtocolFlag informs service to use TLS protocol for kafka
const KafkaTLSProtocolFlag = "TLS"

// Config values for the application
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	CacheTTL                   time.Duration `envconfig:"CACHE_TTL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"             json:"-"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	GraphDriverChoice          string        `envconfig:"GRAPH_DRIVER_TYPE"`
	EnableGetGraphDimensionID  bool          `envconfig:"ENABLE_GET_GRAPH_DIMENSION_ID"`
	KafkaConfig                KafkaConfig
}

// KafkaConfig contains the config required to connect to Kafka
type KafkaConfig struct {
	Brokers                  []string      `envconfig:"KAFKA_ADDR"                  json:"-"`
	Version                  string        `envconfig:"KAFKA_VERSION"`
	BatchSize                int           `envconfig:"BATCH_SIZE"` // number of kafka messages that will be batched
	BatchWaitTime            time.Duration `envconfig:"BATCH_WAIT_TIME"`
	MaxBytes                 int           `envconfig:"KAFKA_MAX_BYTES"`
	OffsetOldest             bool          `envconfig:"KAFKA_OFFSET_OLDEST"`
	SecProtocol              string        `envconfig:"KAFKA_SEC_PROTO"`
	SecClientKey             string        `envconfig:"KAFKA_SEC_CLIENT_KEY"        json:"-"`
	SecClientCert            string        `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	SecCACerts               string        `envconfig:"KAFKA_SEC_CA_CERTS"`
	SecSkipVerify            bool          `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	ObservationConsumerGroup string        `envconfig:"OBSERVATION_CONSUMER_GROUP"`
	ObservationConsumerTopic string        `envconfig:"OBSERVATION_CONSUMER_TOPIC"`
	ErrorProducerTopic       string        `envconfig:"ERROR_PRODUCER_TOPIC"`
	ResultProducerTopic      string        `envconfig:"RESULT_PRODUCER_TOPIC"`
}

func getDefaultConfig() *Config {
	return &Config{
		BindAddr:                   ":21700",
		DatasetAPIURL:              "http://localhost:22000",
		CacheTTL:                   time.Minute * 60,
		GracefulShutdownTimeout:    time.Second * 10,
		ServiceAuthToken:           "AA78C45F-DD64-4631-BED9-FEAE29200620",
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		GraphDriverChoice:          "neo4j",
		EnableGetGraphDimensionID:  true,
		KafkaConfig: KafkaConfig{
			Brokers:                  []string{"localhost:9092"},
			Version:                  "1.0.2",
			BatchSize:                100,
			BatchWaitTime:            time.Millisecond * 200,
			MaxBytes:                 0,
			OffsetOldest:             true,
			ErrorProducerTopic:       "report-events",
			ResultProducerTopic:      "import-observations-inserted",
			ObservationConsumerGroup: "observation-extracted",
			ObservationConsumerTopic: "observation-extracted",
		},
	}
}

// Get the configuration values from the environment or provide the defaults.
func Get(ctx context.Context) (*Config, error) {

	cfg := getDefaultConfig()
	if err := envconfig.Process("", cfg); err != nil {
		return cfg, err
	}

	errs := validateConfig(ctx, cfg)
	if len(errs) != 0 {
		err := fmt.Errorf("config validation errors: %v", strings.Join(errs, ", "))
		log.Error(ctx, "failed on config validation", err)
		return nil, err
	}

	cfg.ServiceAuthToken = "Bearer " + cfg.ServiceAuthToken

	return cfg, nil
}

func validateConfig(ctx context.Context, cfg *Config) []string {
	errs := []string{}

	if len(cfg.ServiceAuthToken) == 0 {
		errs = append(errs, "no SERVICE_AUTH_TOKEN given")
	}

	kafkaCfgErrs := validateKafkaValues(cfg.KafkaConfig)
	if len(kafkaCfgErrs) != 0 {
		log.Info(ctx, "failed kafka configuration validation")
		errs = append(errs, kafkaCfgErrs...)
	}

	return errs
}

func validateKafkaValues(kafkaConfig KafkaConfig) []string {
	errs := []string{}

	if len(kafkaConfig.Brokers) == 0 {
		errs = append(errs, "no KAFKA_ADDR given")
	}

	if kafkaConfig.BatchSize < 1 {
		errs = append(errs, "BATCH_SIZE is less than 1")
	}

	if len(kafkaConfig.Version) == 0 {
		errs = append(errs, "no KAFKA_VERSION given")
	}

	if kafkaConfig.SecProtocol != "" && kafkaConfig.SecProtocol != KafkaTLSProtocolFlag {
		errs = append(errs, "KAFKA_SEC_PROTO has invalid value")
	}

	// isKafkaClientCertSet xor isKafkaClientKeySet
	isKafkaClientCertSet := len(kafkaConfig.SecClientCert) != 0
	isKafkaClientKeySet := len(kafkaConfig.SecClientKey) != 0
	if (isKafkaClientCertSet || isKafkaClientKeySet) && !(isKafkaClientCertSet && isKafkaClientKeySet) {
		errs = append(errs, "only one of KAFKA_SEC_CLIENT_CERT or KAFKA_SEC_CLIENT_KEY has been set - requires both")
	}

	return errs
}
