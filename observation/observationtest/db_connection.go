package observationtest

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

// DBConnection provides a connection to the database.
type DBConnection struct {
	Results []bolt.Result
	Error   error
	Queries []string
	Params  []map[string]interface{}
}

// ExecPipeline captures the parameters given and returns the stored responses.
func (connection *DBConnection) ExecPipeline(queries []string, params ...map[string]interface{}) ([]bolt.Result, error) {
	connection.Queries = queries
	connection.Params = params

	return connection.Results, connection.Error
}
