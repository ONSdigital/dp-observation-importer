package observation

import (
	"fmt"

	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/ONSdigital/go-ns/log"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/structures/messages"
	"strings"
)

const constraintError = "Neo.ClientError.Schema.ConstraintValidationFailed"

// Store provides persistence for observations.
type Store struct {
	dimensionIDCache DimensionIDCache
	dBConnection     DBConnection
	errorReporter    reporter.ErrorReporter
}

// DimensionIDCache provides database ID's of dimensions when inserting observations.
type DimensionIDCache interface {
	GetNodeIDs(instanceID string) (map[string]string, error)
}

// DBConnection provides a connection to the database.
type DBConnection interface {
	ExecNeo(query string, params map[string]interface{}) (bolt.Result, error)
}

// NewStore returns a new Observation store instance that uses the given dimension ID cache and db connection.
func NewStore(dimensionIDCache DimensionIDCache, dBConnection DBConnection, errorReporter reporter.ErrorReporter) *Store {
	return &Store{
		dimensionIDCache: dimensionIDCache,
		dBConnection:     dBConnection,
		errorReporter:    errorReporter,
	}
}

// Result holds the result for an individual instance
type Result struct {
	InstanceID           string
	ObservationsInserted int32
}

// SaveAll the observations against the provided dimension options and instanceID.
func (store *Store) SaveAll(observations []*Observation) ([]*Result, error) {

	results := make([]*Result, 0)

	// handle the inserts separately for each instance in the batch.
	instanceObservations := mapObservationsToInstances(observations)

	for instanceID := range instanceObservations {

		query := buildInsertObservationQuery(instanceID, instanceObservations[instanceID])

		dimensionIds, err := store.dimensionIDCache.GetNodeIDs(instanceID)
		if err != nil {
			store.reportError(instanceID, "failed to get dimension node id's", err)
			continue
		}

		queryParameters, err := createParams(instanceObservations[instanceID], dimensionIds)
		if err != nil {
			store.reportError(instanceID, "failed to create query parameters for batch query", err)
			continue
		}

		queryResult, err := store.dBConnection.ExecNeo(query, queryParameters)
		if err != nil {

			if neo4jErrorCode(err) == constraintError {

				log.Debug("constraint error identified - skipping the inserts for this instance",
					log.Data{"instance_id": instanceID})

				continue
			}

			for instanceID, _ := range instanceObservations {
				store.reportError(instanceID, "observation batch insert failed", err)
			}
			return nil, err
		}

		rowsAffected, err := queryResult.RowsAffected()
		if err != nil {
			store.reportError(instanceID, "error attempting to get number of rows affected in query result", err)
			continue
		}

		log.Debug("save result",
			log.Data{"rows affected": rowsAffected, "metadata": queryResult.Metadata()})

		observationsInserted := int32(len(instanceObservations[instanceID]))

		result := &Result{
			InstanceID:           instanceID,
			ObservationsInserted: observationsInserted,
		}

		results = append(results, result)
	}

	return results, nil
}

func neo4jErrorCode(err error) interface{} {
	if err, ok := err.(*errors.Error); ok {
		if bolterr, ok := err.InnerMost().(*messages.FailureMessage); ok {
			return bolterr.Metadata["code"]
		}
	}

	return ""
}

func (store *Store) reportError(instanceID string, context string, cause error) {
	if err := store.errorReporter.Notify(instanceID, context, cause); err != nil {
		log.ErrorC("errorReporter.Notify returned unexpected error while attempting to report error", err, log.Data{
			"reportedError": context,
		})
	}
}

// createParams creates parameters to inject into an insert query for each observation.
func createParams(observations []*Observation, dimensionIDs map[string]string) (map[string]interface{}, error) {

	rows := make([]interface{}, 0)

	for _, observation := range observations {

		row := map[string]interface{}{
			"v": observation.Row,
			"i": observation.RowIndex,
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
	create := fmt.Sprintf(" CREATE (o:`_%s_observation` { value:row.v, rowIndex:row.i }), ", instanceID)

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
