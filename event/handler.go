package event

import (
	"context"
	"github.com/ONSdigital/dp-graph/v2/models"

	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/ONSdigital/log.go/log"
)

//go:generate moq -out eventtest/observation_mapper.go -pkg eventtest . ObservationMapper
//go:generate moq -out eventtest/observation_store.go -pkg eventtest . ObservationStore
//go:generate moq -out eventtest/result_writer.go -pkg eventtest . ResultWriter

var _ Handler = (*BatchHandler)(nil)

// BatchHandler handles batches of ObservationExtracted events that contain CSV row data.
type BatchHandler struct {
	observationMapper ObservationMapper
	observationStore  ObservationStore
	resultWriter      ResultWriter
	errorReporter     reporter.ErrorReporter
}

// ObservationMapper handles the conversion from row data to observation instances.
type ObservationMapper interface {
	Map(ctx context.Context, row string, rowIndex int64, instanceID string) (*models.Observation, error)
}

// ObservationStore handles the persistence of observations.
type ObservationStore interface {
	SaveAll(ctx context.Context, observations []*models.Observation) ([]*observation.Result, error)
}

// ResultWriter dependency that outputs results
type ResultWriter interface {
	Write(ctx context.Context, results []*observation.Result)
}

// NewBatchHandler returns a new BatchHandler to use the given observation mapper / store.
func NewBatchHandler(
	observationMapper ObservationMapper,
	observationStore ObservationStore,
	resultWriter ResultWriter,
	errorReporter reporter.ErrorReporter) *BatchHandler {

	return &BatchHandler{
		observationMapper: observationMapper,
		observationStore:  observationStore,
		resultWriter:      resultWriter,
		errorReporter:     errorReporter,
	}
}

// Handle the given slice of ObservationExtracted events.
func (handler BatchHandler) Handle(ctx context.Context, events []*ObservationExtracted) error {
	observations := make([]*models.Observation, 0, len(events))

	for _, event := range events {
		observation, err := handler.observationMapper.Map(ctx, event.Row, event.RowIndex, event.InstanceID)
		if err != nil {
			if err := handler.errorReporter.Notify(event.InstanceID, "error while attempting to convert from row data to observation instances", err); err != nil {
				log.Event(ctx, "error reporter notify returned an unexpected error", log.ERROR, log.Error(err))
			}
			continue // do not add this error'd event to the batch
		}

		observations = append(observations, observation)
	}

	results, err := handler.observationStore.SaveAll(ctx, observations)
	if err != nil {
		return err
	}

	handler.resultWriter.Write(ctx, results)

	return nil
}
