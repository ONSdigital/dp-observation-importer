package observationtest

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

// DBConnection provides a connection to the database.
type DBConnection struct {
	Result bolt.Result
	Error  error
	Query  string
	Params map[string]interface{}
}

// ExecNeo captures the parameters given and returns the stored responses.
func (connection *DBConnection) ExecNeo(query string, params map[string]interface{}) (bolt.Result, error) {
	connection.Query = query
	connection.Params = params

	return connection.Result, connection.Error
}
