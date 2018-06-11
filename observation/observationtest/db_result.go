package observationtest

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

var _ bolt.Result = (*DBResult)(nil)

// DBResult is a mock neo4j bolt.result.
type DBResult struct {
	lastInsertId int64
	rowsAffected int64
	metadata     map[string]interface{}
	err          error
}

// NewDBResult returns a new mock DBResult using the given parameter values as responses.
func NewDBResult(lastInsertId int64,
	rowsAffected int64,
	metadata map[string]interface{},
	err error) *DBResult {

	return &DBResult{
		lastInsertId: lastInsertId,
		metadata:     metadata,
		rowsAffected: rowsAffected,
		err:          err,
	}
}

// LastInsertId returns mock values.
func (result DBResult) LastInsertId() (int64, error) {
	return result.lastInsertId, result.err
}

// RowsAffected returns mock values.
func (result DBResult) RowsAffected() (int64, error) {
	return result.rowsAffected, result.err
}

// Metadata returns mock values.
func (result DBResult) Metadata() map[string]interface{} {
	return result.metadata
}
