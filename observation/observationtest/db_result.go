package observationtest

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

var _ bolt.Result = (*DBResult)(nil)

type DBResult struct {
	lastInsertId int64
	rowsAffected int64
	metadata     map[string]interface{}
	err          error
}

func NewDBResult(lastInsertId int64,
	rowsAffected int64,
	metadata map[string]interface{},
	err error) *DBResult {

	return &DBResult{
		lastInsertId:lastInsertId,
		metadata:metadata,
		rowsAffected:rowsAffected,
		err:err,
	}
}

func (result DBResult) LastInsertId() (int64, error) {
	return result.lastInsertId, result.err
}
func (result DBResult) RowsAffected() (int64, error) {
	return result.rowsAffected, result.err
}
func (result DBResult) Metadata() map[string]interface{} {
	return result.metadata
}