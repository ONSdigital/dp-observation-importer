package feature

import (
	"context"
	"os"

	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/dp-observation-importer/cmd/dp-observation-importer/runner"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	initialiserMock "github.com/ONSdigital/dp-observation-importer/initialise/mock"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
)

type ImporterFeature struct {
	service        initialise.ExternalServiceList
	KafkaConsumer  kafka.IConsumerGroup
	FakeDatasetAPI *httpfake.HTTPFake
}

func NewObservationImporterFeature(url string) *ImporterFeature {

	f := &ImporterFeature{}

	f.FakeDatasetAPI = httpfake.New()
	os.Setenv("DATASET_API_URL", f.FakeDatasetAPI.ResolveURL(""))
	f.KafkaConsumer = kafkatest.NewMessageConsumer(false)

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
	return nil
}

func (f *ImporterFeature) theFollowingDataIsInsertedIntoTheGraph(arg1 *godog.DocString) error {
	return godog.ErrPending
}

func (f *ImporterFeature) thisObservationIsConsumed(messageContent *godog.DocString) error {
	runner.Run(context.Background(), config.Get(), f.service)
	observation := event.ObservationExtracted{InstanceID: "7", Row: "5,,sex,male,age,30"}
	bytes, err := schema.ObservationExtractedEvent.Marshal(observation)
	if err != nil {
		return err
	}
	message := kafkatest.NewMessage(bytes, 0)
	f.KafkaConsumer.Channels().Upstream <- message

	return nil
}

// func (f *ImporterFeature) InitialiseService() (http.Handler, error) {
// 	if err := f.svc.Run(context.Background(), "1", "", "", f.errorChan); err != nil {
// 		return nil, err
// 	}
// 	f.ServiceRunning = true
// 	return f.HTTPServer.Handler, nil
// }

func (f *ImporterFeature) DoGetGraphDB(ctx context.Context) (*graph.DB, error) {
	graphDB, err := graph.NewObservationStore(ctx)
	if err != nil {
		return nil, err
	}
	return graphDB, nil
}

func (f *ImporterFeature) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (healthcheck.HealthCheck, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return healthcheck.HealthCheck{}, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return hc, nil
}

func (f *ImporterFeature) DoGetImportErrorReporter(ObservationsImportedErrProducer reporter.KafkaProducer, serviceName string) (errorReporter reporter.ImportErrorReporter, err error) {
	errorReporter, err = reporter.NewImportErrorReporter(ObservationsImportedErrProducer, serviceName)
	return
}

func funcClose(ctx context.Context) error {
	return nil
}

func (f *ImporterFeature) DoGetProducer(ctx context.Context, kafkaBrokers []string, topic string, name initialise.KafkaProducerName, cfg *config.Config) (kafkaProducer kafka.IProducer, err error) {
	return &kafkatest.IProducerMock{
		ChannelsFunc: func() *kafka.ProducerChannels {
			return &kafka.ProducerChannels{}
		},
		CloseFunc: funcClose,
	}, nil
}

func (f *ImporterFeature) DoGetConsumer(ctx context.Context, cfg *config.Config) (kafkaConsumer kafka.IConsumerGroup, err error) {
	return f.KafkaConsumer, nil
}
