// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package observationtest

import (
	"database/sql/driver"
	"sync"
	"time"

	"github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

var (
	lockDBConnectionMockBegin           sync.RWMutex
	lockDBConnectionMockClose           sync.RWMutex
	lockDBConnectionMockExecNeo         sync.RWMutex
	lockDBConnectionMockExecPipeline    sync.RWMutex
	lockDBConnectionMockPrepareNeo      sync.RWMutex
	lockDBConnectionMockPreparePipeline sync.RWMutex
	lockDBConnectionMockQueryNeo        sync.RWMutex
	lockDBConnectionMockQueryNeoAll     sync.RWMutex
	lockDBConnectionMockQueryPipeline   sync.RWMutex
	lockDBConnectionMockSetChunkSize    sync.RWMutex
	lockDBConnectionMockSetTimeout      sync.RWMutex
)

// DBConnectionMock is a mock implementation of DBConnection.
//
//     func TestSomethingThatUsesDBConnection(t *testing.T) {
//
//         // make and configure a mocked DBConnection
//         mockedDBConnection := &DBConnectionMock{
//             BeginFunc: func() (driver.Tx, error) {
// 	               panic("TODO: mock out the Begin method")
//             },
//             CloseFunc: func() error {
// 	               panic("TODO: mock out the Close method")
//             },
//             ExecNeoFunc: func(query string, params map[string]interface{}) (golangNeo4jBoltDriver.Result, error) {
// 	               panic("TODO: mock out the ExecNeo method")
//             },
//             ExecPipelineFunc: func(query []string, params ...map[string]interface{}) ([]golangNeo4jBoltDriver.Result, error) {
// 	               panic("TODO: mock out the ExecPipeline method")
//             },
//             PrepareNeoFunc: func(query string) (golangNeo4jBoltDriver.Stmt, error) {
// 	               panic("TODO: mock out the PrepareNeo method")
//             },
//             PreparePipelineFunc: func(query ...string) (golangNeo4jBoltDriver.PipelineStmt, error) {
// 	               panic("TODO: mock out the PreparePipeline method")
//             },
//             QueryNeoFunc: func(query string, params map[string]interface{}) (golangNeo4jBoltDriver.Rows, error) {
// 	               panic("TODO: mock out the QueryNeo method")
//             },
//             QueryNeoAllFunc: func(query string, params map[string]interface{}) ([][]interface{}, map[string]interface{}, map[string]interface{}, error) {
// 	               panic("TODO: mock out the QueryNeoAll method")
//             },
//             QueryPipelineFunc: func(query []string, params ...map[string]interface{}) (golangNeo4jBoltDriver.PipelineRows, error) {
// 	               panic("TODO: mock out the QueryPipeline method")
//             },
//             SetChunkSizeFunc: func(in1 uint16)  {
// 	               panic("TODO: mock out the SetChunkSize method")
//             },
//             SetTimeoutFunc: func(in1 time.Duration)  {
// 	               panic("TODO: mock out the SetTimeout method")
//             },
//         }
//
//         // TODO: use mockedDBConnection in code that requires DBConnection
//         //       and then make assertions.
//
//     }
type DBConnectionMock struct {
	// BeginFunc mocks the Begin method.
	BeginFunc func() (driver.Tx, error)

	// CloseFunc mocks the Close method.
	CloseFunc func() error

	// ExecNeoFunc mocks the ExecNeo method.
	ExecNeoFunc func(query string, params map[string]interface{}) (golangNeo4jBoltDriver.Result, error)

	// ExecPipelineFunc mocks the ExecPipeline method.
	ExecPipelineFunc func(query []string, params ...map[string]interface{}) ([]golangNeo4jBoltDriver.Result, error)

	// PrepareNeoFunc mocks the PrepareNeo method.
	PrepareNeoFunc func(query string) (golangNeo4jBoltDriver.Stmt, error)

	// PreparePipelineFunc mocks the PreparePipeline method.
	PreparePipelineFunc func(query ...string) (golangNeo4jBoltDriver.PipelineStmt, error)

	// QueryNeoFunc mocks the QueryNeo method.
	QueryNeoFunc func(query string, params map[string]interface{}) (golangNeo4jBoltDriver.Rows, error)

	// QueryNeoAllFunc mocks the QueryNeoAll method.
	QueryNeoAllFunc func(query string, params map[string]interface{}) ([][]interface{}, map[string]interface{}, map[string]interface{}, error)

	// QueryPipelineFunc mocks the QueryPipeline method.
	QueryPipelineFunc func(query []string, params ...map[string]interface{}) (golangNeo4jBoltDriver.PipelineRows, error)

	// SetChunkSizeFunc mocks the SetChunkSize method.
	SetChunkSizeFunc func(in1 uint16)

	// SetTimeoutFunc mocks the SetTimeout method.
	SetTimeoutFunc func(in1 time.Duration)

	// calls tracks calls to the methods.
	calls struct {
		// Begin holds details about calls to the Begin method.
		Begin []struct {
		}
		// Close holds details about calls to the Close method.
		Close []struct {
		}
		// ExecNeo holds details about calls to the ExecNeo method.
		ExecNeo []struct {
			// Query is the query argument value.
			Query string
			// Params is the params argument value.
			Params map[string]interface{}
		}
		// ExecPipeline holds details about calls to the ExecPipeline method.
		ExecPipeline []struct {
			// Query is the query argument value.
			Query []string
			// Params is the params argument value.
			Params []map[string]interface{}
		}
		// PrepareNeo holds details about calls to the PrepareNeo method.
		PrepareNeo []struct {
			// Query is the query argument value.
			Query string
		}
		// PreparePipeline holds details about calls to the PreparePipeline method.
		PreparePipeline []struct {
			// Query is the query argument value.
			Query []string
		}
		// QueryNeo holds details about calls to the QueryNeo method.
		QueryNeo []struct {
			// Query is the query argument value.
			Query string
			// Params is the params argument value.
			Params map[string]interface{}
		}
		// QueryNeoAll holds details about calls to the QueryNeoAll method.
		QueryNeoAll []struct {
			// Query is the query argument value.
			Query string
			// Params is the params argument value.
			Params map[string]interface{}
		}
		// QueryPipeline holds details about calls to the QueryPipeline method.
		QueryPipeline []struct {
			// Query is the query argument value.
			Query []string
			// Params is the params argument value.
			Params []map[string]interface{}
		}
		// SetChunkSize holds details about calls to the SetChunkSize method.
		SetChunkSize []struct {
			// In1 is the in1 argument value.
			In1 uint16
		}
		// SetTimeout holds details about calls to the SetTimeout method.
		SetTimeout []struct {
			// In1 is the in1 argument value.
			In1 time.Duration
		}
	}
}

// Begin calls BeginFunc.
func (mock *DBConnectionMock) Begin() (driver.Tx, error) {
	if mock.BeginFunc == nil {
		panic("moq: DBConnectionMock.BeginFunc is nil but DBConnection.Begin was just called")
	}
	callInfo := struct {
	}{}
	lockDBConnectionMockBegin.Lock()
	mock.calls.Begin = append(mock.calls.Begin, callInfo)
	lockDBConnectionMockBegin.Unlock()
	return mock.BeginFunc()
}

// BeginCalls gets all the calls that were made to Begin.
// Check the length with:
//     len(mockedDBConnection.BeginCalls())
func (mock *DBConnectionMock) BeginCalls() []struct {
} {
	var calls []struct {
	}
	lockDBConnectionMockBegin.RLock()
	calls = mock.calls.Begin
	lockDBConnectionMockBegin.RUnlock()
	return calls
}

// Close calls CloseFunc.
func (mock *DBConnectionMock) Close() error {
	if mock.CloseFunc == nil {
		panic("moq: DBConnectionMock.CloseFunc is nil but DBConnection.Close was just called")
	}
	callInfo := struct {
	}{}
	lockDBConnectionMockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	lockDBConnectionMockClose.Unlock()
	return mock.CloseFunc()
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//     len(mockedDBConnection.CloseCalls())
func (mock *DBConnectionMock) CloseCalls() []struct {
} {
	var calls []struct {
	}
	lockDBConnectionMockClose.RLock()
	calls = mock.calls.Close
	lockDBConnectionMockClose.RUnlock()
	return calls
}

// ExecNeo calls ExecNeoFunc.
func (mock *DBConnectionMock) ExecNeo(query string, params map[string]interface{}) (golangNeo4jBoltDriver.Result, error) {
	if mock.ExecNeoFunc == nil {
		panic("moq: DBConnectionMock.ExecNeoFunc is nil but DBConnection.ExecNeo was just called")
	}
	callInfo := struct {
		Query  string
		Params map[string]interface{}
	}{
		Query:  query,
		Params: params,
	}
	lockDBConnectionMockExecNeo.Lock()
	mock.calls.ExecNeo = append(mock.calls.ExecNeo, callInfo)
	lockDBConnectionMockExecNeo.Unlock()
	return mock.ExecNeoFunc(query, params)
}

// ExecNeoCalls gets all the calls that were made to ExecNeo.
// Check the length with:
//     len(mockedDBConnection.ExecNeoCalls())
func (mock *DBConnectionMock) ExecNeoCalls() []struct {
	Query  string
	Params map[string]interface{}
} {
	var calls []struct {
		Query  string
		Params map[string]interface{}
	}
	lockDBConnectionMockExecNeo.RLock()
	calls = mock.calls.ExecNeo
	lockDBConnectionMockExecNeo.RUnlock()
	return calls
}

// ExecPipeline calls ExecPipelineFunc.
func (mock *DBConnectionMock) ExecPipeline(query []string, params ...map[string]interface{}) ([]golangNeo4jBoltDriver.Result, error) {
	if mock.ExecPipelineFunc == nil {
		panic("moq: DBConnectionMock.ExecPipelineFunc is nil but DBConnection.ExecPipeline was just called")
	}
	callInfo := struct {
		Query  []string
		Params []map[string]interface{}
	}{
		Query:  query,
		Params: params,
	}
	lockDBConnectionMockExecPipeline.Lock()
	mock.calls.ExecPipeline = append(mock.calls.ExecPipeline, callInfo)
	lockDBConnectionMockExecPipeline.Unlock()
	return mock.ExecPipelineFunc(query, params...)
}

// ExecPipelineCalls gets all the calls that were made to ExecPipeline.
// Check the length with:
//     len(mockedDBConnection.ExecPipelineCalls())
func (mock *DBConnectionMock) ExecPipelineCalls() []struct {
	Query  []string
	Params []map[string]interface{}
} {
	var calls []struct {
		Query  []string
		Params []map[string]interface{}
	}
	lockDBConnectionMockExecPipeline.RLock()
	calls = mock.calls.ExecPipeline
	lockDBConnectionMockExecPipeline.RUnlock()
	return calls
}

// PrepareNeo calls PrepareNeoFunc.
func (mock *DBConnectionMock) PrepareNeo(query string) (golangNeo4jBoltDriver.Stmt, error) {
	if mock.PrepareNeoFunc == nil {
		panic("moq: DBConnectionMock.PrepareNeoFunc is nil but DBConnection.PrepareNeo was just called")
	}
	callInfo := struct {
		Query string
	}{
		Query: query,
	}
	lockDBConnectionMockPrepareNeo.Lock()
	mock.calls.PrepareNeo = append(mock.calls.PrepareNeo, callInfo)
	lockDBConnectionMockPrepareNeo.Unlock()
	return mock.PrepareNeoFunc(query)
}

// PrepareNeoCalls gets all the calls that were made to PrepareNeo.
// Check the length with:
//     len(mockedDBConnection.PrepareNeoCalls())
func (mock *DBConnectionMock) PrepareNeoCalls() []struct {
	Query string
} {
	var calls []struct {
		Query string
	}
	lockDBConnectionMockPrepareNeo.RLock()
	calls = mock.calls.PrepareNeo
	lockDBConnectionMockPrepareNeo.RUnlock()
	return calls
}

// PreparePipeline calls PreparePipelineFunc.
func (mock *DBConnectionMock) PreparePipeline(query ...string) (golangNeo4jBoltDriver.PipelineStmt, error) {
	if mock.PreparePipelineFunc == nil {
		panic("moq: DBConnectionMock.PreparePipelineFunc is nil but DBConnection.PreparePipeline was just called")
	}
	callInfo := struct {
		Query []string
	}{
		Query: query,
	}
	lockDBConnectionMockPreparePipeline.Lock()
	mock.calls.PreparePipeline = append(mock.calls.PreparePipeline, callInfo)
	lockDBConnectionMockPreparePipeline.Unlock()
	return mock.PreparePipelineFunc(query...)
}

// PreparePipelineCalls gets all the calls that were made to PreparePipeline.
// Check the length with:
//     len(mockedDBConnection.PreparePipelineCalls())
func (mock *DBConnectionMock) PreparePipelineCalls() []struct {
	Query []string
} {
	var calls []struct {
		Query []string
	}
	lockDBConnectionMockPreparePipeline.RLock()
	calls = mock.calls.PreparePipeline
	lockDBConnectionMockPreparePipeline.RUnlock()
	return calls
}

// QueryNeo calls QueryNeoFunc.
func (mock *DBConnectionMock) QueryNeo(query string, params map[string]interface{}) (golangNeo4jBoltDriver.Rows, error) {
	if mock.QueryNeoFunc == nil {
		panic("moq: DBConnectionMock.QueryNeoFunc is nil but DBConnection.QueryNeo was just called")
	}
	callInfo := struct {
		Query  string
		Params map[string]interface{}
	}{
		Query:  query,
		Params: params,
	}
	lockDBConnectionMockQueryNeo.Lock()
	mock.calls.QueryNeo = append(mock.calls.QueryNeo, callInfo)
	lockDBConnectionMockQueryNeo.Unlock()
	return mock.QueryNeoFunc(query, params)
}

// QueryNeoCalls gets all the calls that were made to QueryNeo.
// Check the length with:
//     len(mockedDBConnection.QueryNeoCalls())
func (mock *DBConnectionMock) QueryNeoCalls() []struct {
	Query  string
	Params map[string]interface{}
} {
	var calls []struct {
		Query  string
		Params map[string]interface{}
	}
	lockDBConnectionMockQueryNeo.RLock()
	calls = mock.calls.QueryNeo
	lockDBConnectionMockQueryNeo.RUnlock()
	return calls
}

// QueryNeoAll calls QueryNeoAllFunc.
func (mock *DBConnectionMock) QueryNeoAll(query string, params map[string]interface{}) ([][]interface{}, map[string]interface{}, map[string]interface{}, error) {
	if mock.QueryNeoAllFunc == nil {
		panic("moq: DBConnectionMock.QueryNeoAllFunc is nil but DBConnection.QueryNeoAll was just called")
	}
	callInfo := struct {
		Query  string
		Params map[string]interface{}
	}{
		Query:  query,
		Params: params,
	}
	lockDBConnectionMockQueryNeoAll.Lock()
	mock.calls.QueryNeoAll = append(mock.calls.QueryNeoAll, callInfo)
	lockDBConnectionMockQueryNeoAll.Unlock()
	return mock.QueryNeoAllFunc(query, params)
}

// QueryNeoAllCalls gets all the calls that were made to QueryNeoAll.
// Check the length with:
//     len(mockedDBConnection.QueryNeoAllCalls())
func (mock *DBConnectionMock) QueryNeoAllCalls() []struct {
	Query  string
	Params map[string]interface{}
} {
	var calls []struct {
		Query  string
		Params map[string]interface{}
	}
	lockDBConnectionMockQueryNeoAll.RLock()
	calls = mock.calls.QueryNeoAll
	lockDBConnectionMockQueryNeoAll.RUnlock()
	return calls
}

// QueryPipeline calls QueryPipelineFunc.
func (mock *DBConnectionMock) QueryPipeline(query []string, params ...map[string]interface{}) (golangNeo4jBoltDriver.PipelineRows, error) {
	if mock.QueryPipelineFunc == nil {
		panic("moq: DBConnectionMock.QueryPipelineFunc is nil but DBConnection.QueryPipeline was just called")
	}
	callInfo := struct {
		Query  []string
		Params []map[string]interface{}
	}{
		Query:  query,
		Params: params,
	}
	lockDBConnectionMockQueryPipeline.Lock()
	mock.calls.QueryPipeline = append(mock.calls.QueryPipeline, callInfo)
	lockDBConnectionMockQueryPipeline.Unlock()
	return mock.QueryPipelineFunc(query, params...)
}

// QueryPipelineCalls gets all the calls that were made to QueryPipeline.
// Check the length with:
//     len(mockedDBConnection.QueryPipelineCalls())
func (mock *DBConnectionMock) QueryPipelineCalls() []struct {
	Query  []string
	Params []map[string]interface{}
} {
	var calls []struct {
		Query  []string
		Params []map[string]interface{}
	}
	lockDBConnectionMockQueryPipeline.RLock()
	calls = mock.calls.QueryPipeline
	lockDBConnectionMockQueryPipeline.RUnlock()
	return calls
}

// SetChunkSize calls SetChunkSizeFunc.
func (mock *DBConnectionMock) SetChunkSize(in1 uint16) {
	if mock.SetChunkSizeFunc == nil {
		panic("moq: DBConnectionMock.SetChunkSizeFunc is nil but DBConnection.SetChunkSize was just called")
	}
	callInfo := struct {
		In1 uint16
	}{
		In1: in1,
	}
	lockDBConnectionMockSetChunkSize.Lock()
	mock.calls.SetChunkSize = append(mock.calls.SetChunkSize, callInfo)
	lockDBConnectionMockSetChunkSize.Unlock()
	mock.SetChunkSizeFunc(in1)
}

// SetChunkSizeCalls gets all the calls that were made to SetChunkSize.
// Check the length with:
//     len(mockedDBConnection.SetChunkSizeCalls())
func (mock *DBConnectionMock) SetChunkSizeCalls() []struct {
	In1 uint16
} {
	var calls []struct {
		In1 uint16
	}
	lockDBConnectionMockSetChunkSize.RLock()
	calls = mock.calls.SetChunkSize
	lockDBConnectionMockSetChunkSize.RUnlock()
	return calls
}

// SetTimeout calls SetTimeoutFunc.
func (mock *DBConnectionMock) SetTimeout(in1 time.Duration) {
	if mock.SetTimeoutFunc == nil {
		panic("moq: DBConnectionMock.SetTimeoutFunc is nil but DBConnection.SetTimeout was just called")
	}
	callInfo := struct {
		In1 time.Duration
	}{
		In1: in1,
	}
	lockDBConnectionMockSetTimeout.Lock()
	mock.calls.SetTimeout = append(mock.calls.SetTimeout, callInfo)
	lockDBConnectionMockSetTimeout.Unlock()
	mock.SetTimeoutFunc(in1)
}

// SetTimeoutCalls gets all the calls that were made to SetTimeout.
// Check the length with:
//     len(mockedDBConnection.SetTimeoutCalls())
func (mock *DBConnectionMock) SetTimeoutCalls() []struct {
	In1 time.Duration
} {
	var calls []struct {
		In1 time.Duration
	}
	lockDBConnectionMockSetTimeout.RLock()
	calls = mock.calls.SetTimeout
	lockDBConnectionMockSetTimeout.RUnlock()
	return calls
}
