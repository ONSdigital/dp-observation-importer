package feature

import (
	"context"

	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/cucumber/godog"
)

type ImporterFeature struct {
	service initialise.ExternalServiceList
}

type InitMock struct{}

func NewObservationImporterFeature(url string) *ImporterFeature {

	f := &ImporterFeature{}

	initMock := &InitMock{}

	f.service = initialise.NewServiceList(initMock)

	return f
}

func (f *ImporterFeature) RegisterSteps(context *godog.ScenarioContext) {

}

func (f *ImporterFeature) Close() {

}

func (f *ImporterFeature) Reset() {

}

// func (f *ImporterFeature) InitialiseService() (http.Handler, error) {
// 	if err := f.svc.Run(context.Background(), "1", "", "", f.errorChan); err != nil {
// 		return nil, err
// 	}
// 	f.ServiceRunning = true
// 	return f.HTTPServer.Handler, nil
// }

func (i *InitMock) DoGetGraphDB(ctx context.Context) (*graph.DB, error) {
	graphDB, err := graph.NewObservationStore(ctx)
	if err != nil {
		return nil, err
	}
	return graphDB, nil
}

func (i *InitMock) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (healthcheck.HealthCheck, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return healthcheck.HealthCheck{}, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return hc, nil
}

func (i *InitMock) DoGetImportErrorReporter(ObservationsImportedErrProducer reporter.KafkaProducer, serviceName string) (errorReporter reporter.ImportErrorReporter, err error) {
	errorReporter, err = reporter.NewImportErrorReporter(ObservationsImportedErrProducer, serviceName)
	return
}

func (i *InitMock) DoGetProducer(ctx context.Context, kafkaBrokers []string, topic string, name initialise.KafkaProducerName, cfg *config.Config) (kafkaProducer *kafka.Producer, err error) {
	return &kafka.Producer{}, nil
}

func (i *InitMock) DoGetConsumer(ctx context.Context, cfg *config.Config) (kafkaConsumer *kafka.ConsumerGroup, err error) {
	cgChannels := kafka.CreateConsumerGroupChannels(cfg.BatchSize)
	var kafkaOffset = kafka.OffsetOldest

	cgConfig := &kafka.ConsumerGroupConfig{
		Offset:       &kafkaOffset,
		KafkaVersion: &cfg.KafkaVersion,
	}
	kafkaConsumer, err = kafka.NewConsumerGroup(
		ctx,
		cfg.Brokers,
		cfg.ObservationConsumerTopic,
		cfg.ObservationConsumerGroup,
		cgChannels,
		cgConfig,
	)
	return
}
