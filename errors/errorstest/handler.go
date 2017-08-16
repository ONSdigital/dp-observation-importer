// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package errorstest

import (
	"github.com/ONSdigital/go-ns/log"
	"sync"
)

var (
	lockHandlerMockHandle sync.RWMutex
)

// HandlerMock is a mock implementation of Handler.
//
//     func TestSomethingThatUsesHandler(t *testing.T) {
//
//         // make and configure a mocked Handler
//         mockedHandler := &HandlerMock{
//             HandleFunc: func(instanceID string, err error, data log.Data)  {
// 	               panic("TODO: mock out the Handle method")
//             },
//         }
//
//         // TODO: use mockedHandler in code that requires Handler
//         //       and then make assertions.
//
//     }
type HandlerMock struct {
	// HandleFunc mocks the Handle method.
	HandleFunc func(instanceID string, err error, data log.Data)

	// calls tracks calls to the methods.
	calls struct {
		// Handle holds details about calls to the Handle method.
		Handle []struct {
			// InstanceID is the instanceID argument value.
			InstanceID string
			// Err is the err argument value.
			Err error
			// Data is the data argument value.
			Data log.Data
		}
	}
}

// Handle calls HandleFunc.
func (mock *HandlerMock) Handle(instanceID string, err error, data log.Data) {
	if mock.HandleFunc == nil {
		panic("moq: HandlerMock.HandleFunc is nil but Handler.Handle was just called")
	}
	callInfo := struct {
		InstanceID string
		Err        error
		Data       log.Data
	}{
		InstanceID: instanceID,
		Err:        err,
		Data:       data,
	}
	lockHandlerMockHandle.Lock()
	mock.calls.Handle = append(mock.calls.Handle, callInfo)
	lockHandlerMockHandle.Unlock()
	mock.HandleFunc(instanceID, err, data)
}

// HandleCalls gets all the calls that were made to Handle.
// Check the length with:
//     len(mockedHandler.HandleCalls())
func (mock *HandlerMock) HandleCalls() []struct {
	InstanceID string
	Err        error
	Data       log.Data
} {
	var calls []struct {
		InstanceID string
		Err        error
		Data       log.Data
	}
	lockHandlerMockHandle.RLock()
	calls = mock.calls.Handle
	lockHandlerMockHandle.RUnlock()
	return calls
}