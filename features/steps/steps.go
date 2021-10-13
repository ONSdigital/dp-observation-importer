package feature

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/dimension"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-observation-importer/schema"
	"github.com/ONSdigital/dp-observation-importer/service"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
)

func (f *ImporterFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^these observations are consumed:$`, f.theseObservationsAreConsumed)
	ctx.Step(`^these dimensions should be inserted into the database for batch "([^"]*)":$`, f.theseDimensionsShouldBeInsertedIntoTheDatabaseForBatch)
	ctx.Step(`^these observations should be inserted into the database for batch "([^"]*)":$`, f.theseObservationsShouldBeInsertedIntoTheDatabaseForBatch)
	ctx.Step(`^a message stating "([^"]*)" observation\(s\) inserted for instance ID "([^"]*)" is produced$`, f.aMessageStatingObservationsInsertedForInstanceIDIsProduced)
	ctx.Step(`^the observation batch size is set to "([^"]*)"$`, f.theObservationBatchSizeIsSetTo)

	ctx.Step(`^instance "([^"]*)" on dataset-api has dimensions:$`, f.instanceOnDatasetapiHasDimensions)
	ctx.Step(`^instance "([^"]*)" on dataset-api has headers "([^"]*)"$`, f.instanceOnDatasetapiHasHeaders)
	ctx.Step(`^instance "([^"]*)" on dataset-api has no dimensions$`, f.instanceOnDatasetapiHasNoDimensions)
}

func (f *ImporterFeature) instanceOnDatasetapiHasDimensions(instanceID string, table *godog.Table) error {
	dimensions, err := f.convertToDimensionsJson(table)
	if err != nil {
		return err
	}

	f.FakeDatasetAPI.NewHandler().Get("/instances/" + instanceID + "/dimensions").Reply(200).BodyString("{\"Items\": " + string(dimensions) + "}")
	return nil
}

func (f *ImporterFeature) instanceOnDatasetapiHasNoDimensions(instanceID string) error {
	f.FakeDatasetAPI.NewHandler().Get("/instances/" + instanceID + "/dimensions").Reply(200).BodyString("{\"Items\": [] }")
	return nil
}

func (f *ImporterFeature) instanceOnDatasetapiHasHeaders(instanceID string, headers string) error {
	type csvHeaders struct {
		Headers []string `json:"headers"`
	}
	data := csvHeaders{
		Headers: strings.Split(headers, ","),
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	f.FakeDatasetAPI.NewHandler().Get("/instances/" + instanceID).Reply(200).BodyString(string(b))
	return nil
}

func (f *ImporterFeature) theseObservationsAreConsumed(table *godog.Table) error {

	observationEvents, err := f.convertToObservationEvents(table)
	if err != nil {
		return err
	}

	signals := f.registerInterrupt()

	cfg, err := config.Get(context.Background())
	if err != nil {
		return err
	}

	// run application in separate goroutine
	go func() {
		_ = service.Run(context.Background(), cfg, f.serviceList, signals, "", "", "")
	}()

	// consume extracted observations
	for _, e := range observationEvents {
		if err := f.sendToConsumer(e); err != nil {
			return err
		}
	}

	time.Sleep(300 * time.Millisecond)

	// kill application
	signals <- os.Interrupt

	return nil
}

func (f *ImporterFeature) registerInterrupt() chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	return signals
}

func (f *ImporterFeature) sendToConsumer(e *event.ObservationExtracted) error {
	bytes, err := schema.ObservationExtractedEvent.Marshal(e)
	if err != nil {
		return err
	}

	f.KafkaConsumer.Channels().Upstream <- kafkatest.NewMessage(bytes, 0)
	return nil
}

func (f *ImporterFeature) theseObservationsShouldBeInsertedIntoTheDatabaseForBatch(batch int, observasionsJson *godog.DocString) error {

	actualObservations := f.ObservationDB.InsertObservationBatchCalls()[batch].Observations

	actualObservationsJson, err := json.Marshal(actualObservations)
	if err != nil {
		return err
	}
	assert.JSONEq(&f.ErrorFeature, observasionsJson.Content, string(actualObservationsJson))

	return f.ErrorFeature.StepError()
}

func (f *ImporterFeature) theseDimensionsShouldBeInsertedIntoTheDatabaseForBatch(batch int, table *godog.Table) error {
	type dimension struct {
		Dimension string
		NodeID    string
	}
	assist := assistdog.NewDefault()
	dimensions, err := assist.CreateSlice(&dimension{}, table)
	if err != nil {
		return err
	}

	for _, dim := range dimensions.([]*dimension) {
		calls := f.ObservationDB.InsertObservationBatchCalls()
		assert.Contains(&f.ErrorFeature, calls[batch].DimensionIDs, dim.Dimension)
		if f.ErrorFeature.StepError() != nil {
			return f.ErrorFeature.StepError()
		}
		assert.Equal(&f.ErrorFeature, dim.NodeID, calls[batch].DimensionIDs[dim.Dimension])
		if f.ErrorFeature.StepError() != nil {
			return f.ErrorFeature.StepError()
		}
	}
	return f.ErrorFeature.StepError()
}

func (f *ImporterFeature) aMessageStatingObservationsInsertedForInstanceIDIsProduced(count int32, instanceID string) error {

	message := <-f.KafkaProducer.Channels().Output
	var unmarshalledMessage observation.InsertedEvent

	schema.ObservationsInsertedEvent.Unmarshal(message, &unmarshalledMessage)

	assert.Equal(&f.ErrorFeature, instanceID, unmarshalledMessage.InstanceID)
	assert.Equal(&f.ErrorFeature, count, unmarshalledMessage.ObservationsInserted)
	return f.ErrorFeature.StepError()
}

func (f *ImporterFeature) convertToDimensionsJson(table *godog.Table) ([]byte, error) {
	assist := assistdog.NewDefault()
	var dimensionItem = &dimension.Dimension{}
	dimensions, err := assist.CreateSlice(dimensionItem, table)
	if err != nil {
		return nil, err
	}

	return json.Marshal(dimensions)
}

func (f *ImporterFeature) convertToObservationEvents(table *godog.Table) ([]*event.ObservationExtracted, error) {
	assist := assistdog.NewDefault()
	events, err := assist.CreateSlice(&event.ObservationExtracted{}, table)
	if err != nil {
		return nil, err
	}
	return events.([]*event.ObservationExtracted), nil
}

func (f *ImporterFeature) theObservationBatchSizeIsSetTo(batchSize string) error {
	os.Setenv("BATCH_SIZE", batchSize)

	return nil
}
