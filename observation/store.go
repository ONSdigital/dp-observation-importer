package observation

import (
	"fmt"

	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/ONSdigital/go-ns/log"
	bolt "github.com/ONSdigital/golang-neo4j-bolt-driver"
	neoErrors "github.com/ONSdigital/golang-neo4j-bolt-driver/errors"
	"github.com/ONSdigital/golang-neo4j-bolt-driver/structures/messages"
	"strings"
)

const constraintError = "Neo.ClientError.Schema.ConstraintValidationFailed"

// Store provides persistence for observations.
type Store struct {
	dimensionIDCache DimensionIDCache
	pool             DBPool
	errorReporter    reporter.ErrorReporter
}

// DimensionIDCache provides database ID's of dimensions when inserting observations.
type DimensionIDCache interface {
	GetNodeIDs(instanceID string) (map[string]string, error)
}

//go:generate moq -out observationtest/db_pool.go -pkg observationtest . DBPool

// DBPool used to require bolt connections
type DBPool interface {
	OpenPool() (bolt.Conn, error)
}

//go:generate moq -out observationtest/db_conn.go -pkg observationtest . DBConnection
// DBConnection is only used to generate mocked bolt connections
type DBConnection bolt.Conn

// NewStore returns a new Observation store instance that uses the given dimension ID cache and db connection.
func NewStore(dimensionIDCache DimensionIDCache, pool DBPool, errorReporter reporter.ErrorReporter) *Store {
	return &Store{
		dimensionIDCache: dimensionIDCache,
		pool:     pool,
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

		conn, err := store.pool.OpenPool()
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		queryResult, err := conn.ExecNeo(query, queryParameters)
		if err != nil {

			if neo4jErrorCode(err) == constraintError {

				log.Debug("constraint error identified - skipping the inserts for this instance",
					log.Data{"instance_id": instanceID})

				continue
			}

			for instanceID := range instanceObservations {
				store.reportError(instanceID, "observation batch insert failed", err)
			}

			// todo: add retry logic and identify fatal errors that should be returned.
			// any error will currently be sent to the reporter and mark the import as failed
			// we do not return the error as we want the message to be consumed.
			// returning an error causes the service to shutdown and not consume the message.
			continue
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
	if boltErr, ok := err.(*neoErrors.Error); ok {
		if failureMessage, ok := boltErr.Inner().(messages.FailureMessage); ok {
			return failureMessage.Metadata["code"]
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
