// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package eventtest

import (
	"github.com/ONSdigital/dp-observation-importer/observation"
	"sync"
)

var (
	lockResultWriterMockWrite sync.RWMutex
)

// ResultWriterMock is a mock implementation of ResultWriter.
//
//     func TestSomethingThatUsesResultWriter(t *testing.T) {
//
//         // make and configure a mocked ResultWriter
//         mockedResultWriter := &ResultWriterMock{
//             WriteFunc: func(results []*observation.Result) error {
// 	               panic("TODO: mock out the Write method")
//             },
//         }
//
//         // TODO: use mockedResultWriter in code that requires ResultWriter
//         //       and then make assertions.
//
//     }
type ResultWriterMock struct {
	// WriteFunc mocks the Write method.
	WriteFunc func(results []*observation.Result) error

	// calls tracks calls to the methods.
	calls struct {
		// Write holds details about calls to the Write method.
		Write []struct {
			// Results is the results argument value.
			Results []*observation.Result
		}
	}
}

// Write calls WriteFunc.
func (mock *ResultWriterMock) Write(results []*observation.Result) error {
	if mock.WriteFunc == nil {
		panic("moq: ResultWriterMock.WriteFunc is nil but ResultWriter.Write was just called")
	}
	callInfo := struct {
		Results []*observation.Result
	}{
		Results: results,
	}
	lockResultWriterMockWrite.Lock()
	mock.calls.Write = append(mock.calls.Write, callInfo)
	lockResultWriterMockWrite.Unlock()
	return mock.WriteFunc(results)
}

// WriteCalls gets all the calls that were made to Write.
// Check the length with:
//     len(mockedResultWriter.WriteCalls())
func (mock *ResultWriterMock) WriteCalls() []struct {
	Results []*observation.Result
} {
	var calls []struct {
		Results []*observation.Result
	}
	lockResultWriterMockWrite.RLock()
	calls = mock.calls.Write
	lockResultWriterMockWrite.RUnlock()
	return calls
}
