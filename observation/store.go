package observation

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"fmt"
	"github.com/ONSdigital/go-ns/log"
)

// Store provides persistence for observations.
type Store struct {
	dimensionIDCache DimensionIDCache
	dBConnection     DBConnection
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
func NewStore(dimensionIDCache DimensionIDCache, dBConnection DBConnection) *Store {
	return &Store{
		dimensionIDCache: dimensionIDCache,
		dBConnection:     dBConnection,
	}
}

// SaveAll the observations against the provided dimension options and instanceID.
func (store *Store) SaveAll(observations []*Observation) error {

	// handle the inserts separately for each instance in the batch.
	instanceObservations := mapObservationsToInstances(observations)

	var queries = make([]string, 0)                        // query for each different instance
	var pipelineParams = make([]map[string]interface{}, 0) // a set of parameters for each observation to insert

	for instanceID := range instanceObservations {

		dimensionIds, err := store.dimensionIDCache.GetNodeIDs(instanceID)
		if err != nil {
			log.Error(err,log.Data{"message": "Failed to get dimension node ID's", "instance":instanceID})
			continue
		}

		query := buildInsertObservationQuery(instanceID, instanceObservations[instanceID])
		queries = append(queries, query)

		params, err := createParams(instanceObservations[instanceID], dimensionIds)
		pipelineParams = append(pipelineParams, params)

	}

	results, err := store.dBConnection.ExecPipeline(queries, pipelineParams...)
	if err != nil {
		return err
	}

	for _, result := range results {

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Error(err,log.Data{"message": "Error running observation insert statement"})
		}

		log.Debug("Save result",
			log.Data{"rows affected": rowsAffected, "metadata": result.Metadata()})

		// todo: check for error results
		// handle constraint violation - retry
		// exponential back off
	}

	return nil
}

// createParams creates parameters to inject into an insert query for each observation.
func createParams(observations []*Observation, dimensionIDs map[string]string) (map[string]interface{}, error) {

	rows := make([]interface{}, 0)

	for _, observation := range observations {

		row := map[string]interface{}{
			"v": observation.Row,
		}

		for _, option := range observation.DimensionOptions {

			dimensionLookUp := observation.InstanceID + "_" + option.DimensionName + "_" + option.Name

			nodeID, ok := dimensionIDs[dimensionLookUp]
			if ! ok {
				return nil, fmt.Errorf("No nodeId found for %s", dimensionLookUp)
			}

			//row["name"] = option
			row[option.DimensionName] = nodeID
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
	create := fmt.Sprintf(" CREATE (o:_%s_observation { value:row.v }), ", instanceID)

	index := 0

	for _, option := range observations[0].DimensionOptions {

		if index != 0 {
			match += ", "
			where += " AND "
			create += ", "
		}

		match += fmt.Sprintf("(%s:_%s_%s)", option.DimensionName, instanceID, option.DimensionName)
		where += fmt.Sprintf("id(%s) = toInt(row.%s)", option.DimensionName, option.DimensionName)
		create += fmt.Sprintf("(o)-[:isValueOf]->(%s)", option.DimensionName)
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
		if _, hasKey := instanceObservations[observation.InstanceID]; !hasKey {
			instanceObservations[observation.InstanceID] = make([]*Observation, 0)
		}

		value := instanceObservations[observation.InstanceID]

		instanceObservations[observation.InstanceID] = append(value, observation)
	}

	return instanceObservations
}
