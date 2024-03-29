package observation

import (
	"context"

	"github.com/ONSdigital/dp-graph/v2/models"

	graph "github.com/ONSdigital/dp-graph/v2/graph/driver"
	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/ONSdigital/log.go/v2/log"
)

// DimensionIDCache provides database ID's of dimensions when inserting observations.
type DimensionIDCache interface {
	GetNodeIDs(ctx context.Context, instanceID string) (map[string]string, error)
}

// Store provides persistence for observations.
type Store struct {
	dimensionIDCache    DimensionIDCache
	graph               graph.Observation
	errorReporter       reporter.ErrorReporter
	getGraphDimensionID bool
}

// NewStore returns a new Observation store instance that uses the given dimension ID cache and db connection.
func NewStore(dimensionIDCache DimensionIDCache, db graph.Observation, errorReporter reporter.ErrorReporter, getGraphDimensionID bool) *Store {
	return &Store{
		dimensionIDCache:    dimensionIDCache,
		graph:               db,
		errorReporter:       errorReporter,
		getGraphDimensionID: getGraphDimensionID,
	}
}

// Result holds the result for an individual instance
type Result struct {
	InstanceID           string
	ObservationsInserted int32
}

// SaveAll the observations against the provided dimension options and instanceID.
func (store *Store) SaveAll(ctx context.Context, observations []*models.Observation) ([]*Result, error) {
	results := make([]*Result, 0)

	// handle the inserts separately for each instance in the batch.
	instanceObservations := mapObservationsToInstances(observations)

	for instanceID, observations := range instanceObservations {
		var dimensionIds map[string]string
		var err error
		if store.getGraphDimensionID {
			dimensionIds, err = store.dimensionIDCache.GetNodeIDs(ctx, instanceID)
			if err != nil {
				store.reportError(ctx, instanceID, "failed to get dimension node id's for batch", err)
				return results, err
			}
		}

		if err := store.graph.InsertObservationBatch(context.Background(), 1, instanceID, observations, dimensionIds); err != nil {
			store.reportError(ctx, instanceID, "failed to insert observation batch to graph", err)
			continue
		}

		observationsInserted := int32(len(observations))

		result := &Result{
			InstanceID:           instanceID,
			ObservationsInserted: observationsInserted,
		}

		results = append(results, result)
	}

	return results, nil
}

func (store *Store) reportError(ctx context.Context, instanceID string, message string, cause error) {
	if err := store.errorReporter.Notify(instanceID, message, cause); err != nil {
		log.Error(ctx, "errorReporter.Notify returned unexpected error while attempting to report error", err, log.Data{
			"reportedError": message,
		})
	}
}

// mapObservationsToInstances separates observations by instance id.
func mapObservationsToInstances(observations []*models.Observation) map[string][]*models.Observation {

	var instanceObservations = make(map[string][]*models.Observation)

	for _, observation := range observations {

		// add the instance key and new list if it does not exist.
		if _, ok := instanceObservations[observation.InstanceID]; !ok {
			instanceObservations[observation.InstanceID] = make([]*models.Observation, 0)
		}

		value := instanceObservations[observation.InstanceID]

		instanceObservations[observation.InstanceID] = append(value, observation)
	}

	return instanceObservations
}
