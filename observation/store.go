package observation

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"strings"

	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/ONSdigital/go-ns/log"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	neoErrors "github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/structures/messages"
)

const (
	constraintError      = "Neo.ClientError.Schema.ConstraintValidationFailed"
	statementErrorPrefix = "Neo.ClientError.Statement"
)

// Store provides persistence for observations.
type Store struct {
	dimensionIDCache DimensionIDCache
	pool             DBPool
	errorReporter    reporter.ErrorReporter
	maxRetries       int
}

// ErrAttemptsExceededLimit is returned when the number of attempts has reaced
// the maximum permitted
type ErrAttemptsExceededLimit struct {
	WrappedErr error
}

func (e ErrAttemptsExceededLimit) Error() string {
	return fmt.Sprintf("number of attempts to save observations exceeded: %s", e.WrappedErr.Error())
}

// ErrNonRetriable is returned when the wrapped error type is not retriable
type ErrNonRetriable struct {
	WrappedErr error
}

func (e ErrNonRetriable) Error() string {
	return fmt.Sprintf("received a non retriable error from neo4j: %s", e.WrappedErr.Error())
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
func NewStore(dimensionIDCache DimensionIDCache, pool DBPool, errorReporter reporter.ErrorReporter, maxRetries int) *Store {
	return &Store{
		dimensionIDCache: dimensionIDCache,
		pool:             pool,
		errorReporter:    errorReporter,
		maxRetries:       maxRetries,
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

	conn, err := store.pool.OpenPool()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error(err, nil)
		}
	}()

	for instanceID, observations := range instanceObservations {

		err := store.save(1, store.maxRetries, conn, instanceID, observations)
		if err != nil {
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

func (store *Store) save(attempt, maxAttempts int, conn bolt.Conn, instanceID string, observations []*Observation) error {
	query := buildInsertObservationQuery(instanceID, observations)

	dimensionIds, err := store.dimensionIDCache.GetNodeIDs(instanceID)
	if err != nil {
		store.reportError(instanceID, "failed to get dimension node id's for batch", err)
		return err
	}

	queryParameters, err := createParams(observations, dimensionIds)
	if err != nil {
		store.reportError(instanceID, "failed to create query parameters for batch query", err)
		return err
	}

	queryResult, err := conn.ExecNeo(query, queryParameters)
	if err != nil {

		if neoErr, ok := checkNeo4jError(err); ok {
			log.Info("received an error from neo4j that cannot be retried",
				log.Data{"instance_id": instanceID, "error": neoErr})

			return ErrNonRetriable{err}
		}

		time.Sleep(getSleepTime(attempt, 20*time.Millisecond))

		if attempt >= maxAttempts {
			store.reportError(instanceID, "observation batch save failed", ErrAttemptsExceededLimit{err})

			return ErrAttemptsExceededLimit{err}
		}

		log.Info("got an error when saving observations, attempting to retry", log.Data{
			"instance_id":  instanceID,
			"retry_number": attempt,
			"max_attempts": maxAttempts,
		})

		// Retry until it is successful or runs out of retries. TODO: the ability to retry
		// neo queries should be built into go-ns, similar to the rchttp package.
		return store.save(attempt+1, maxAttempts, conn, instanceID, observations)
	}

	rowsAffected, err := queryResult.RowsAffected()
	if err != nil {
		store.reportError(instanceID, "error attempting to get number of rows affected in query result", err)
		return err
	}

	log.Info("successfully saved observation batch", log.Data{"rows_affected": rowsAffected, "instance_id": instanceID})

	return nil

}

func checkNeo4jError(err error) (string, bool) {
	var neoErr string
	var boltErr *neoErrors.Error
	var ok bool

	if boltErr, ok = err.(*neoErrors.Error); !ok {
		return "", false
	}

	if failureMessage, ok := boltErr.Inner().(messages.FailureMessage); ok {
		if neoErr, ok = failureMessage.Metadata["code"].(string); !ok {
			return "", false
		}
	}

	if neoErr != constraintError && !strings.Contains(neoErr, statementErrorPrefix) {
		return "", false
	}

	s := strings.Split(err.Error(), "\n")

	//get the first 5 useful lines of a neo stack error
	var shortErr string
	c := 0
	for _, l := range s {
		if c < 5 && l != "" {
			shortErr += l + "\n"
			c++
		}
	}

	return shortErr, true
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

// getSleepTime will return a sleep time based on the attempt and initial retry time.
// It uses the algorithm 2^n where n is the attempt number (double the previous) and
// a randomization factor of between 0-5ms so that the server isn't being hit constantly
// at the same time by many clients
func getSleepTime(attempt int, retryTime time.Duration) time.Duration {
	n := (math.Pow(2, float64(attempt)))
	rand.Seed(time.Now().Unix())
	rnd := time.Duration(rand.Intn(4)+1) * time.Millisecond
	return (time.Duration(n) * retryTime) - rnd
}
