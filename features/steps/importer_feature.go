package feature

import (
	"context"
	"encoding/json"
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
	"github.com/ONSdigital/dp-observation-importer/dimension"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	initialiserMock "github.com/ONSdigital/dp-observation-importer/initialise/mock"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-observation-importer/observation/mock"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/dp-reporter-client/reporter"
	featuretest "github.com/armakuni/dp-go-featuretest"
	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
)

type ImporterFeature struct {
	ErrorFeature   featuretest.ErrorFeature
	service        initialise.ExternalServiceList
	KafkaConsumer  kafka.IConsumerGroup
	KafkaProducer  kafka.IProducer
	FakeDatasetAPI *httpfake.HTTPFake
	ObservationDB  *mock.ObservationMock
	killChannel    chan os.Signal
}

func NewObservationImporterFeature(url string) *ImporterFeature {

	f := &ImporterFeature{}

	f.FakeDatasetAPI = httpfake.New()
	os.Setenv("DATASET_API_URL", f.FakeDatasetAPI.ResolveURL(""))
	os.Setenv("GRAPH_DRIVER_TYPE", "mock")
	os.Setenv("BATCH_SIZE", "1")
	f.KafkaConsumer = kafkatest.NewMessageConsumer(false)

	f.ObservationDB = &mock.ObservationMock{
		InsertObservationBatchFunc: func(ctx context.Context, attempt int, instanceID string, observations []*models.Observation, dimensionIDs map[string]string) error {
			fmt.Println("inside insert function")
			fmt.Println("observations: ", observations)
			fmt.Println("dimensions received: ", dimensionIDs)
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

	f.service = initialise.NewServiceList(initMock)

	return f
}

func (f *ImporterFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^dataset instance "([^"]*)" has dimensions:$`, f.datasetInstanceHasDimensions)
	ctx.Step(`^dataset instance "([^"]*)" has no dimensions$`, f.datasetInstanceHasNoDimensions)
	ctx.Step(`^dataset instance "([^"]*)" has headers:$`, f.datasetInstanceHasHeaders)
	ctx.Step(`^observation "([^"]*)" is inserted into the graph for instance ID "([^"]*)"$`, f.observationIsInsertedIntoTheGraphForInstanceID)
	ctx.Step(`^dimension key "([^"]*)" is mapped to "([^"]*)"$`, f.dimensionKeyIsMappedTo)
	ctx.Step(`^these observations are consumed:$`, f.theseObservationsAreConsumed)
	ctx.Step(`^a message stating "([^"]*)" observation\(s\) inserted for instance ID "([^"]*)" is sent$`, f.aMessageStatingObservationsInsertedForInstanceIDIsSent)
}

func (f *ImporterFeature) Close() {
	f.FakeDatasetAPI.Close()
	// f.KafkaConsumer.Close(context.Background())
}

func (f *ImporterFeature) Reset() {
	f.FakeDatasetAPI.Reset()
	// f.KafkaConsumer = kafkatest.NewMessageConsumer(true)
}

func (f *ImporterFeature) datasetInstanceHasDimensions(instanceID string, table *godog.Table) error {
	assist := assistdog.NewDefault()
	var dimensionItem = &dimension.Dimension{}
	dimensions, err := assist.CreateSlice(dimensionItem, table)
	if err != nil {
		return err
	}
	resultJSON, err := json.Marshal(dimensions)
	if err != nil {
		return err
	}
	f.FakeDatasetAPI.NewHandler().Get("/instances/" + instanceID + "/dimensions").Reply(200).BodyString("{\"Items\": " + string(resultJSON) + "}")
	return nil
}

func (f *ImporterFeature) datasetInstanceHasNoDimensions(instanceID string) error {
	f.FakeDatasetAPI.NewHandler().Get("/instances/" + instanceID + "/dimensions").Reply(200).BodyString("{\"Items\": [] }")
	return nil
}

func (f *ImporterFeature) datasetInstanceHasHeaders(instanceID string, headers *godog.DocString) error {
	f.FakeDatasetAPI.NewHandler().Get("/instances/" + instanceID).Reply(200).BodyString(headers.Content)
	return nil
}

func (f *ImporterFeature) observationIsInsertedIntoTheGraphForInstanceID(data string, ID string) error {
	calls := f.ObservationDB.InsertObservationBatchCalls()
	assert.Equal(&f.ErrorFeature, data, calls[0].Observations[0].Row)
	if f.ErrorFeature.StepError() != nil {
		return f.ErrorFeature.StepError()
	}
	assert.Equal(&f.ErrorFeature, ID, calls[0].Observations[0].InstanceID)

	return f.ErrorFeature.StepError()
}

func (f *ImporterFeature) dimensionKeyIsMappedTo(key, value string) error {
	calls := f.ObservationDB.InsertObservationBatchCalls()
	assert.Contains(&f.ErrorFeature, calls[0].DimensionIDs, key)
	if f.ErrorFeature.StepError() != nil {
		return f.ErrorFeature.StepError()
	}
	assert.Equal(&f.ErrorFeature, value, calls[0].DimensionIDs[key])

	return f.ErrorFeature.StepError()
}

func (f *ImporterFeature) theseObservationsAreConsumed(table *godog.Table) error {
	config, err := config.Get()
	if err != nil {
		return err
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		runner.Run(context.Background(), config, f.service, signals)
	}()

	assist := assistdog.NewDefault()
	observation := &event.ObservationExtracted{}
	events, err := assist.CreateSlice(observation, table)
	if err != nil {
		return err
	}

	for _, event := range events.([]*event.ObservationExtracted) {
		bytes, err := schema.ObservationExtractedEvent.Marshal(event)
		if err != nil {
			return err
		}
		message := kafkatest.NewMessage(bytes, 0)

		f.KafkaConsumer.Channels().Upstream <- message
	}

	time.Sleep(300 * time.Millisecond)

	signals <- os.Interrupt

	return nil
}

func (f *ImporterFeature) aMessageStatingObservationsInsertedForInstanceIDIsSent(count int32, instanceID string) error {

	message := <-f.KafkaProducer.Channels().Output
	var unmarshalledMessage observation.InsertedEvent

	schema.ObservationsInsertedEvent.Unmarshal(message, &unmarshalledMessage)

	assert.Equal(&f.ErrorFeature, instanceID, unmarshalledMessage.InstanceID)
	assert.Equal(&f.ErrorFeature, count, unmarshalledMessage.ObservationsInserted)
	return f.ErrorFeature.StepError()
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
	return f.KafkaProducer, nil
}

func (f *ImporterFeature) DoGetConsumer(ctx context.Context, cfg *config.Config) (kafkaConsumer kafka.IConsumerGroup, err error) {
	return f.KafkaConsumer, nil
}
