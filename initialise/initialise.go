package initialise

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/Shopify/sarama"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser

// ExternalServiceList represents a list of services
type ExternalServiceList struct {
	Consumer                        bool
	ObservationsImportedProducer    bool
	ObservationsImportedErrProducer bool
	Graph                           bool
	ErrorReporter                   bool
	HealthCheck                     bool
	Init                            Initialiser
}

type Initialiser interface {
	DoGetConsumer(context.Context, string, string, *config.KafkaConfig) (kafka.IConsumerGroup, error)
	DoGetProducer(context.Context, string, *config.KafkaConfig) (kafka.IProducer, error)
	DoGetImportErrorReporter(reporter.KafkaProducer, string) (reporter.ImportErrorReporter, error)
	DoGetHealthCheck(*config.Config, string, string, string) (healthcheck.HealthCheck, error)
	DoGetGraphDB(context.Context) (*graph.DB, error)
}

// KafkaProducerName : Type for kafka producer name used by iota constants
type KafkaProducerName int

// Possible names of Kafka Producers
const (
	ObservationsImported = iota
	ObservationsImportedErr
)

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) ExternalServiceList {
	return ExternalServiceList{
		Init: initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

var kafkaProducerNames = []string{"ObservationsImported", "ObservationsImportedErr"}

// Values of the kafka producers names
func (k KafkaProducerName) String() string {
	return kafkaProducerNames[k]
}

// GetConsumer returns a kafka consumer, which might not be initialised yet.
func (e *ExternalServiceList) GetConsumer(ctx context.Context, cfg *config.Config) (kafkaConsumer kafka.IConsumerGroup, err error) {
	kafkaConsumer, err = e.Init.DoGetConsumer(ctx, cfg.KafkaConfig.ObservationConsumerTopic, cfg.KafkaConfig.ObservationConsumerGroup, &cfg.KafkaConfig)
	if err != nil {
		return
	}
	e.Consumer = true
	return
}

// GetProducer returns a kafka producer, which might not be initialised yet.
func (e *ExternalServiceList) GetProducer(ctx context.Context, kafkaBrokers []string, topic string, name KafkaProducerName, cfg *config.Config) (kafkaProducer kafka.IProducer, err error) {
	kafkaProducer, err = e.Init.DoGetProducer(ctx, topic, &cfg.KafkaConfig)
	if err != nil {
		return
	}
	switch {
	case name == ObservationsImported:
		e.ObservationsImportedProducer = true
	case name == ObservationsImportedErr:
		e.ObservationsImportedErrProducer = true
	default:
		err = fmt.Errorf("kafka producer name not recognised: '%s'. Valid names: %v", name.String(), kafkaProducerNames)
	}

	return
}

// GetImportErrorReporter returns an ErrorImportReporter to send error reports to the import-reporter (only if ObservationsImportedErrProducer is available)
func (e *ExternalServiceList) GetImportErrorReporter(observationsImportedErrProducer reporter.KafkaProducer, serviceName string) (errorReporter reporter.ImportErrorReporter, err error) {
	if !e.ObservationsImportedErrProducer {
		return reporter.ImportErrorReporter{},
			fmt.Errorf("cannot create ImportErrorReporter because kafka producer '%s' is not available", kafkaProducerNames[ObservationsImportedErr])
	}

	errorReporter, err = e.Init.DoGetImportErrorReporter(observationsImportedErrProducer, serviceName)
	if err != nil {
		return
	}

	e.ErrorReporter = true
	return
}

// GetHealthCheck creates a healthcheck with versionInfo
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (healthcheck.HealthCheck, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return healthcheck.HealthCheck{}, err
	}
	e.HealthCheck = true
	return hc, nil
}

// GetGraphDB returns a graphDB
func (e *ExternalServiceList) GetGraphDB(ctx context.Context) (*graph.DB, error) {
	graphDB, err := e.Init.DoGetGraphDB(ctx)
	if err != nil {
		return nil, err
	}
	e.Graph = true
	return graphDB, err
}

func (i *Init) DoGetGraphDB(ctx context.Context) (*graph.DB, error) {
	graphDB, err := graph.NewObservationStore(ctx)
	if err != nil {
		return nil, err
	}
	return graphDB, nil
}

func (i *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (healthcheck.HealthCheck, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return healthcheck.HealthCheck{}, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return hc, nil
}

func (i *Init) DoGetImportErrorReporter(observationsImportedErrProducer reporter.KafkaProducer, serviceName string) (errorReporter reporter.ImportErrorReporter, err error) {
	errorReporter, err = reporter.NewImportErrorReporter(observationsImportedErrProducer, serviceName)
	return
}

func (i *Init) DoGetProducer(ctx context.Context, topic string, kafkaConfig *config.KafkaConfig) (kafkaProducer kafka.IProducer, err error) {
	pChannels := kafka.CreateProducerChannels()
	pConfig := &kafka.ProducerConfig{
		KafkaVersion:    &kafkaConfig.Version,
		MaxMessageBytes: &kafkaConfig.MaxBytes,
	}

	if kafkaConfig.SecProtocol == config.KafkaTLSProtocolFlag {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			kafkaConfig.SecCACerts,
			kafkaConfig.SecClientCert,
			kafkaConfig.SecClientKey,
			kafkaConfig.SecSkipVerify,
		)
	}

	kafkaProducer, err = kafka.NewProducer(ctx, kafkaConfig.Brokers, topic, pChannels, pConfig)
	return
}

func (i *Init) DoGetConsumer(ctx context.Context, topic, group string, kafkaConfig *config.KafkaConfig) (kafkaConsumer kafka.IConsumerGroup, err error) {
	cgChannels := kafka.CreateConsumerGroupChannels(kafkaConfig.BatchSize)

	offset := sarama.OffsetOldest
	if !kafkaConfig.OffsetOldest {
		offset = sarama.OffsetNewest
	}

	cgConfig := &kafka.ConsumerGroupConfig{
		Offset:       &offset,
		KafkaVersion: &kafkaConfig.Version,
	}

	if kafkaConfig.SecProtocol == config.KafkaTLSProtocolFlag {
		cgConfig.SecurityConfig = kafka.GetSecurityConfig(
			kafkaConfig.SecCACerts,
			kafkaConfig.SecClientCert,
			kafkaConfig.SecClientKey,
			kafkaConfig.SecSkipVerify,
		)
	}

	kafkaConsumer, err = kafka.NewConsumerGroup(
		ctx,
		kafkaConfig.Brokers,
		topic,
		group,
		cgChannels,
		cgConfig,
	)
	return
}
