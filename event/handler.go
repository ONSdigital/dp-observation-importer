package event

import (
	"github.com/ONSdigital/dp-observation-importer/observation"
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
}

// ObservationMapper handles the conversion from row data to observation instances.
type ObservationMapper interface {
	Map(row string, instanceID string) (*observation.Observation, error)
}

// ObservationStore handles the persistence of observations.
type ObservationStore interface {
	SaveAll(observations []*observation.Observation) ([]*observation.Result, error)
}

// ResultWriter dependency that outputs results
type ResultWriter interface {
	Write(results []*observation.Result)
}

// NewBatchHandler returns a new BatchHandler to use the given observation mapper / store.
func NewBatchHandler(observationMapper ObservationMapper, observationStore ObservationStore, resultWriter ResultWriter) *BatchHandler {
	return &BatchHandler{
		observationMapper: observationMapper,
		observationStore:  observationStore,
		resultWriter:      resultWriter,
	}
}

// Handle the given slice of ObservationExtracted events.
func (handler BatchHandler) Handle(events []*ObservationExtracted) error {
	observations := make([]*observation.Observation, 0, len(events))

	for _, event := range events {
		observation, err := handler.observationMapper.Map(event.Row, event.InstanceID)
		if err != nil {
			return err
		}

		observations = append(observations, observation)
	}

	results, err := handler.observationStore.SaveAll(observations)
	if err != nil {
		return err
	}

	handler.resultWriter.Write(results)

	return nil
}
