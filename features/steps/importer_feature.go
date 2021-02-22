package feature

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"os/signal"

	graphConfig "github.com/ONSdigital/dp-graph/v2/config"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-graph/v2/graph/driver"
	graphMock "github.com/ONSdigital/dp-graph/v2/mock"
	"github.com/ONSdigital/dp-graph/v2/models"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	runner "github.com/ONSdigital/dp-observation-importer/cmd/dp-observation-importer"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	initialiserMock "github.com/ONSdigital/dp-observation-importer/initialise/mock"
	"github.com/ONSdigital/dp-observation-importer/observation/mock"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
)

type ImporterFeature struct {
	service        initialise.ExternalServiceList
	KafkaConsumer  kafka.IConsumerGroup
	FakeDatasetAPI *httpfake.HTTPFake
	ObservationDB  *mock.ObservationMock
}

func NewObservationImporterFeature(url string) *ImporterFeature {

	f := &ImporterFeature{}

	f.FakeDatasetAPI = httpfake.New()
	os.Setenv("DATASET_API_URL", f.FakeDatasetAPI.ResolveURL(""))
	os.Setenv("GRAPH_DRIVER_TYPE", "mock")
	f.KafkaConsumer = kafkatest.NewMessageConsumer(false)

	f.ObservationDB = &mock.ObservationMock{
		InsertObservationBatchFunc: func(ctx context.Context, attempt int, instanceID string, observations []*models.Observation, dimensionIDs map[string]string) error {
			fmt.Println("inside insert function")
			return nil
		},
	}

	initMock := &initialiserMock.InitialiserMock{
		DoGetConsumerFunc:            f.DoGetConsumer,
		DoGetProducerFunc:            f.DoGetProducer,
		DoGetImportErrorReporterFunc: f.DoGetImportErrorReporter,
		DoGetHealthCheckFunc:         f.DoGetHealthCheck,
		DoGetGraphDBFunc:             f.DoGetGraphDB,
	}

	f.service = initialise.NewServiceList(initMock)

	return f
}

func (f *ImporterFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^for instance ID "([^"]*)" the dataset api has headers$`, f.forInstanceIDTheDatasetApiHasHeaders)
	ctx.Step(`^the following data is inserted into the graph$`, f.theFollowingDataIsInsertedIntoTheGraph)
	ctx.Step(`^this observation is consumed:$`, f.thisObservationIsConsumed)
}

func (f *ImporterFeature) Close() {

}

func (f *ImporterFeature) Reset() {

}

func (f *ImporterFeature) forInstanceIDTheDatasetApiHasHeaders(instanceId string, headersString *godog.DocString) error {
	f.FakeDatasetAPI.NewHandler().Get("/instances/" + instanceId).Reply(200).BodyString(headersString.Content)
	f.FakeDatasetAPI.NewHandler().Get("/instances/" + instanceId + "/dimensions").Reply(200).BodyString(headersString.Content)
	f.FakeDatasetAPI.NewHandler().Get("/health").Reply(200)
	f.FakeDatasetAPI.NewHandler().Get("/healthcheck").Reply(200)
	return nil
}

func (f *ImporterFeature) theFollowingDataIsInsertedIntoTheGraph(data *godog.DocString) error {
	fmt.Println(f.ObservationDB.InsertObservationBatchCalls()[0].Observations[0].Row)
	return godog.ErrPending
}

func (f *ImporterFeature) thisObservationIsConsumed(messageContent *godog.DocString) error {
	config, err := config.Get()

	observation := event.ObservationExtracted{InstanceID: "7", Row: "5,,sex,male,age,30"}
	bytes, err := schema.ObservationExtractedEvent.Marshal(observation)
	if err != nil {
		return err
	}
	message := kafkatest.NewMessage(bytes, 0)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		runner.Run(context.Background(), config, f.service, signals)
	}()

	f.KafkaConsumer.Channels().Upstream <- message

	time.Sleep(1 * time.Second)

	signals <- os.Interrupt

	return nil
}

func fakeInsert(ctx context.Context, attempt int, instanceID string, observations []*models.Observation, dimensionIDs map[string]string) error {
	fmt.Println("inside insert function")
	return nil
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

func funcClose(ctx context.Context) error {
	return nil
}

func funcCheck(ctx context.Context, state *healthcheck.CheckState) error {
	return nil
}

func (f *ImporterFeature) DoGetProducer(ctx context.Context, kafkaBrokers []string, topic string, name initialise.KafkaProducerName, cfg *config.Config) (kafkaProducer kafka.IProducer, err error) {
	return &kafkatest.IProducerMock{
		ChannelsFunc: func() *kafka.ProducerChannels {
			return &kafka.ProducerChannels{}
		},
		CloseFunc:   funcClose,
		CheckerFunc: funcCheck,
	}, nil
}

func (f *ImporterFeature) DoGetConsumer(ctx context.Context, cfg *config.Config) (kafkaConsumer kafka.IConsumerGroup, err error) {
	return f.KafkaConsumer, nil
}
