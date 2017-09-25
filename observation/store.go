package observation

import (
	"fmt"

	"github.com/ONSdigital/dp-observation-importer/errors"
	"github.com/ONSdigital/go-ns/log"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"strings"
)

// Store provides persistence for observations.
type Store struct {
	dimensionIDCache DimensionIDCache
	dBConnection     DBConnection
	errorHandler     errors.Handler
}

// DimensionIDCache provides database ID's of dimensions when inserting observations.
type DimensionIDCache interface {
	GetNodeIDs(instanceID string) (map[string]string, error)
}

// DBConnection provides a connection to the database.
type DBConnection interface {
	ExecPipeline(query []string, params ...map[string]interface{}) ([]bolt.Result, error)
}

// NewStore returns a new Observation store instance that uses the given dimension ID cache and db connection.
func NewStore(dimensionIDCache DimensionIDCache, dBConnection DBConnection, errorHandler errors.Handler) *Store {
	return &Store{
		dimensionIDCache: dimensionIDCache,
		dBConnection:     dBConnection,
		errorHandler:     errorHandler,
	}
}

// Result holds a single
type Result struct {
	InstanceID           string
	ObservationsInserted int32
}

// SaveAll the observations against the provided dimension options and instanceID.
func (store *Store) SaveAll(observations []*Observation) ([]*Result, error) {

	results := make([]*Result, 0)

	// handle the inserts separately for each instance in the batch.
	instanceObservations := mapObservationsToInstances(observations)

	var queries = make([]string, 0)                        // query for each different instance
	var pipelineParams = make([]map[string]interface{}, 0) // a set of parameters for each observation to insert

	for instanceID := range instanceObservations {

		dimensionIds, err := store.dimensionIDCache.GetNodeIDs(instanceID)
		if err != nil {
			store.errorHandler.Handle(instanceID, err, log.Data{"message": "failed to get dimension node id's"})
			continue
		}

		query := buildInsertObservationQuery(instanceID, instanceObservations[instanceID])
		queries = append(queries, query)

		params, err := createParams(instanceObservations[instanceID], dimensionIds)
		if err != nil {
			store.errorHandler.Handle(instanceID, err, log.Data{"message": "failed create params for batch query"})
			continue
		}

		pipelineParams = append(pipelineParams, params)

		// create a result placeholder with the instance ID
		results = append(results, &Result{InstanceID: instanceID})
	}

	pipelineResults, err := store.dBConnection.ExecPipeline(queries, pipelineParams...)
	if err != nil {
		return nil, err
	}

	results = store.processResults(pipelineResults, results, instanceObservations)

	return results, nil
}

func (store *Store) processResults(pipelineResults []bolt.Result, results []*Result, instanceObservations map[string][]*Observation) []*Result {

	// we have no context of which result is for which instance ID.
	// the only way we can align them is to use the same index into the result arrays.
	// a result array is passed in containing each instance ID, and the update count is injected into it when the
	// DB result is checked.
	resultIndex := 0

	for _, result := range pipelineResults {

		instanceID := results[resultIndex].InstanceID

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			store.errorHandler.Handle(instanceID, err, log.Data{"message": "error running observation insert statement"})
			continue
		}

		log.Debug("save result",
			log.Data{"rows affected": rowsAffected, "metadata": result.Metadata()})

		results[resultIndex].ObservationsInserted = int32(len(instanceObservations[instanceID]))
		resultIndex++
	}

	return results
}

// createParams creates parameters to inject into an insert query for each observation.
func createParams(observations []*Observation, dimensionIDs map[string]string) (map[string]interface{}, error) {

	rows := make([]interface{}, 0)

	for _, observation := range observations {

		row := map[string]interface{}{
			"v": observation.Row,
		}

		for _, option := range observation.DimensionOptions {

			dimensionName := strings.ToLower(option.DimensionName)

			dimensionLookUp := observation.InstanceID + "_" + dimensionName + "_" + option.Name

			nodeID, ok := dimensionIDs[dimensionLookUp]
			if !ok {
				return nil, fmt.Errorf("No nodeId found for %s", dimensionLookUp)
			}

			row[dimensionName] = nodeID
		}

		rows = append(rows, row)
	}

	return map[string]interface{}{"rows": rows}, nil
}

// buildInsertObservationQuery creates an instance specific insert query.
func buildInsertObservationQuery(instanceID string, observations []*Observation) string {

	query := "UNWIND $rows AS row"

	match := " MATCH "
	where := " WHERE "
	create := fmt.Sprintf(" CREATE (o:`_%s_observation` { value:row.v }), ", instanceID)

	index := 0

	for _, option := range observations[0].DimensionOptions {

		if index != 0 {
			match += ", "
			where += " AND "
			create += ", "
		}
		optionName := strings.ToLower(option.DimensionName)

		match += fmt.Sprintf("(`%s`:`_%s_%s`)", optionName, instanceID, optionName)
		where += fmt.Sprintf("id(`%s`) = toInt(row.`%s`)", optionName, optionName)
		create += fmt.Sprintf("(o)-[:isValueOf]->(`%s`)", optionName)
		index++
	}

	query += match + where + create

	return query
}

// mapObservationsToInstances separates observations by instance id.
func mapObservationsToInstances(observations []*Observation) map[string][]*Observation {

	var instanceObservations = make(map[string][]*Observation)

	for _, observation := range observations {

		// add the instance key and new list if it does not exist.
		if _, ok := instanceObservations[observation.InstanceID]; !ok {
			instanceObservations[observation.InstanceID] = make([]*Observation, 0)
		}

		value := instanceObservations[observation.InstanceID]

		instanceObservations[observation.InstanceID] = append(value, observation)
	}

	return instanceObservations
}
