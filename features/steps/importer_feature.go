package feature

import (
	"context"
	"os"

	graphConfig "github.com/ONSdigital/dp-graph/v2/config"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-graph/v2/graph/driver"
	graphMock "github.com/ONSdigital/dp-graph/v2/mock"
	"github.com/ONSdigital/dp-graph/v2/models"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	initialiserMock "github.com/ONSdigital/dp-observation-importer/initialise/mock"
	"github.com/ONSdigital/dp-observation-importer/observation/mock"
	"github.com/ONSdigital/dp-reporter-client/reporter"
	featuretest "github.com/armakuni/dp-go-featuretest"
	"github.com/maxcnunes/httpfake"
)

type ImporterFeature struct {
	ErrorFeature   featuretest.ErrorFeature
	serviceList    initialise.ExternalServiceList
	KafkaConsumer  kafka.IConsumerGroup
	KafkaProducer  kafka.IProducer
	FakeDatasetAPI *httpfake.HTTPFake
	ObservationDB  *mock.ObservationMock
	killChannel    chan os.Signal
}

func NewObservationImporterFeature() *ImporterFeature {

	f := &ImporterFeature{}

	f.FakeDatasetAPI = httpfake.New()
	os.Setenv("DATASET_API_URL", f.FakeDatasetAPI.ResolveURL(""))
	os.Setenv("GRAPH_DRIVER_TYPE", "mock")
	consumer := kafkatest.NewMessageConsumer(false)
	consumer.CheckerFunc = funcCheck
	f.KafkaConsumer = consumer

	// setup fake InsertObservationBatch function so we can record calls made to it
	f.ObservationDB = &mock.ObservationMock{
		InsertObservationBatchFunc: func(ctx context.Context, attempt int, instanceID string, observations []*models.Observation, dimensionIDs map[string]string) error {
			return nil
		},
	}

	channels := &kafka.ProducerChannels{
		Output: make(chan []byte),
	}
	f.KafkaProducer = &kafkatest.IProducerMock{
		ChannelsFunc: func() *kafka.ProducerChannels {
			return channels
		},
		CloseFunc:   funcClose,
		CheckerFunc: funcCheck,
	}

	initMock := &initialiserMock.InitialiserMock{
		DoGetConsumerFunc:            f.DoGetConsumer,
		DoGetProducerFunc:            f.DoGetProducer,
		DoGetImportErrorReporterFunc: f.DoGetImportErrorReporter,
		DoGetHealthCheckFunc:         f.DoGetHealthCheck,
		DoGetGraphDBFunc:             f.DoGetGraphDB,
	}

	f.serviceList = initialise.NewServiceList(initMock)

	return f
}

func (f *ImporterFeature) Close() {
	f.FakeDatasetAPI.Close()
}

func (f *ImporterFeature) Reset() {
	f.FakeDatasetAPI.Reset()
}

func (f *ImporterFeature) DoGetGraphDB(ctx context.Context) (*graph.DB, error) {
	errs := make(chan error)

	cfg, err := graphConfig.Get(errs)
	if err != nil {
		return nil, err
	}

	cfg.Driver = &graphMock.Mock{
		IsBackendReachable: true,
		IsQueryValid:       true,
		IsContentFound:     true,
	}

	var codelist driver.CodeList
	var hierarchy driver.Hierarchy
	var instance driver.Instance

	var observation driver.Observation = f.ObservationDB
	var dimension driver.Dimension

	return &graph.DB{
		Driver:      cfg.Driver,
		CodeList:    codelist,
		Hierarchy:   hierarchy,
		Instance:    instance,
		Observation: observation,
		Dimension:   dimension,
		Errors:      errs,
	}, nil
}

func (f *ImporterFeature) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (healthcheck.HealthCheck, error) {
	versionInfo, err := healthcheck.NewVersionInfo("1234", "gitCommit", "version")
	if err != nil {
		return healthcheck.HealthCheck{}, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	hc.Status = "200"
	return hc, nil
}

func (f *ImporterFeature) DoGetImportErrorReporter(ObservationsImportedErrProducer reporter.KafkaProducer, serviceName string) (errorReporter reporter.ImportErrorReporter, err error) {
	errorReporter, err = reporter.NewImportErrorReporter(ObservationsImportedErrProducer, serviceName)
	return
}

func (f *ImporterFeature) DoGetProducer(ctx context.Context, kafkaBrokers []string, topic string, name initialise.KafkaProducerName, cfg *config.Config) (kafkaProducer kafka.IProducer, err error) {
	return f.KafkaProducer, nil
}

func (f *ImporterFeature) DoGetConsumer(ctx context.Context, cfg *config.Config) (kafkaConsumer kafka.IConsumerGroup, err error) {
	return f.KafkaConsumer, nil
}

func funcClose(ctx context.Context) error {
	return nil
}

func funcCheck(ctx context.Context, state *healthcheck.CheckState) error {
	return nil
}
