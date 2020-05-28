package initialise

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-reporter-client/reporter"
)

// ExternalServiceList represents a list of services
type ExternalServiceList struct {
	Consumer                        bool
	ObservationsImportedProducer    bool
	ObservationsImportedErrProducer bool
	Graph                           bool
	ErrorReporter                   bool
	HealthCheck                     bool
}

// KafkaProducerName : Type for kafka producer name used by iota constants
type KafkaProducerName int

// Possible names of Kafa Producers
const (
	ObservationsImported = iota
	ObservationsImportedErr
)

var kafkaProducerNames = []string{"ObservationsImported", "ObservationsImportedErr"}

// Values of the kafka producers names
func (k KafkaProducerName) String() string {
	return kafkaProducerNames[k]
}

// GetConsumer returns a kafka consumer, which might not be initialised yet.
func (e *ExternalServiceList) GetConsumer(ctx context.Context, cfg *config.Config) (kafkaConsumer *kafka.ConsumerGroup, err error) {
	cgChannels := kafka.CreateConsumerGroupChannels(true)
	kafkaConsumer, err = kafka.NewConsumerGroup(
		ctx,
		cfg.Brokers,
		cfg.ObservationConsumerTopic,
		cfg.ObservationConsumerGroup,
		kafka.OffsetNewest,
		true,
		cgChannels,
	)
	if err != nil {
		return
	}

	e.Consumer = true
	return
}

// GetProducer returns a kafka producer, which might not be initialised yet.
func (e *ExternalServiceList) GetProducer(ctx context.Context, kafkaBrokers []string, topic string, name KafkaProducerName, envMax int) (kafkaProducer *kafka.Producer, err error) {
	pChannels := kafka.CreateProducerChannels()
	kafkaProducer, err = kafka.NewProducer(ctx, kafkaBrokers, topic, envMax, pChannels)
	if err != nil {
		return
	}

	switch {
	case name == ObservationsImported:
		e.ObservationsImportedProducer = true
	case name == ObservationsImportedErr:
		e.ObservationsImportedErrProducer = true
	default:
		err = fmt.Errorf("Kafka producer name not recognised: '%s'. Valid names: %v", name.String(), kafkaProducerNames)
	}

	return
}

// GetImportErrorReporter returns an ErrorImportReporter to send error reports to the import-reporter (only if ObservationsImportedErrProducer is available)
func (e *ExternalServiceList) GetImportErrorReporter(ObservationsImportedErrProducer reporter.KafkaProducer, serviceName string) (errorReporter reporter.ImportErrorReporter, err error) {
	if !e.ObservationsImportedErrProducer {
		return reporter.ImportErrorReporter{},
			fmt.Errorf("Cannot create ImportErrorReporter because kafka producer '%s' is not available", kafkaProducerNames[ObservationsImportedErr])
	}

	errorReporter, err = reporter.NewImportErrorReporter(ObservationsImportedErrProducer, serviceName)
	if err != nil {
		return
	}

	e.ErrorReporter = true
	return
}

// GetHealthCheck creates a healthcheck with versionInfo
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (healthcheck.HealthCheck, error) {

	// Create healthcheck object with versionInfo
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return healthcheck.HealthCheck{}, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	e.HealthCheck = true

	return hc, nil
}

// GetGraphDB returns a graphDB
func (e *ExternalServiceList) GetGraphDB(ctx context.Context) (*graph.DB, error) {

	graphDB, err := graph.NewObservationStore(ctx)
	if err != nil {
		return nil, err
	}

	e.Graph = true

	return graphDB, nil
}
