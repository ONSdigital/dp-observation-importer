// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package mocks

import (
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
//             HandleFunc: func(instanceID string, err error)  {
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
	HandleFunc func(instanceID string, err error)

	// calls tracks calls to the methods.
	calls struct {
		// Handle holds details about calls to the Handle method.
		Handle []struct {
			// InstanceID is the instanceID argument value.
			InstanceID string
			// Err is the err argument value.
			Err error
		}
	}
}

// Handle calls HandleFunc.
func (mock *HandlerMock) Handle(instanceID string, err error) {
	if mock.HandleFunc == nil {
		panic("moq: HandlerMock.HandleFunc is nil but Handler.Handle was just called")
	}
	callInfo := struct {
		InstanceID string
		Err        error
	}{
		InstanceID: instanceID,
		Err:        err,
	}
	lockHandlerMockHandle.Lock()
	mock.calls.Handle = append(mock.calls.Handle, callInfo)
	lockHandlerMockHandle.Unlock()
	mock.HandleFunc(instanceID, err)
}

// HandleCalls gets all the calls that were made to Handle.
// Check the length with:
//     len(mockedHandler.HandleCalls())
func (mock *HandlerMock) HandleCalls() []struct {
	InstanceID string
	Err        error
} {
	var calls []struct {
		InstanceID string
		Err        error
	}
	lockHandlerMockHandle.RLock()
	calls = mock.calls.Handle
	lockHandlerMockHandle.RUnlock()
	return calls
}

var (
	lockMessageProducerMockCloser sync.RWMutex
	lockMessageProducerMockOutput sync.RWMutex
)

// MessageProducerMock is a mock implementation of MessageProducer.
//
//     func TestSomethingThatUsesMessageProducer(t *testing.T) {
//
//         // make and configure a mocked MessageProducer
//         mockedMessageProducer := &MessageProducerMock{
//             CloserFunc: func() chan bool {
// 	               panic("TODO: mock out the Closer method")
//             },
//             OutputFunc: func() chan []byte {
// 	               panic("TODO: mock out the Output method")
//             },
//         }
//
//         // TODO: use mockedMessageProducer in code that requires MessageProducer
//         //       and then make assertions.
//
//     }
type MessageProducerMock struct {
	// CloserFunc mocks the Closer method.
	CloserFunc func() chan bool

	// OutputFunc mocks the Output method.
	OutputFunc func() chan []byte

	// calls tracks calls to the methods.
	calls struct {
		// Closer holds details about calls to the Closer method.
		Closer []struct {
		}
		// Output holds details about calls to the Output method.
		Output []struct {
		}
	}
}

// Closer calls CloserFunc.
func (mock *MessageProducerMock) Closer() chan bool {
	if mock.CloserFunc == nil {
		panic("moq: MessageProducerMock.CloserFunc is nil but MessageProducer.Closer was just called")
	}
	callInfo := struct {
	}{}
	lockMessageProducerMockCloser.Lock()
	mock.calls.Closer = append(mock.calls.Closer, callInfo)
	lockMessageProducerMockCloser.Unlock()
	return mock.CloserFunc()
}

// CloserCalls gets all the calls that were made to Closer.
// Check the length with:
//     len(mockedMessageProducer.CloserCalls())
func (mock *MessageProducerMock) CloserCalls() []struct {
} {
	var calls []struct {
	}
	lockMessageProducerMockCloser.RLock()
	calls = mock.calls.Closer
	lockMessageProducerMockCloser.RUnlock()
	return calls
}

// Output calls OutputFunc.
func (mock *MessageProducerMock) Output() chan []byte {
	if mock.OutputFunc == nil {
		panic("moq: MessageProducerMock.OutputFunc is nil but MessageProducer.Output was just called")
	}
	callInfo := struct {
	}{}
	lockMessageProducerMockOutput.Lock()
	mock.calls.Output = append(mock.calls.Output, callInfo)
	lockMessageProducerMockOutput.Unlock()
	return mock.OutputFunc()
}

// OutputCalls gets all the calls that were made to Output.
// Check the length with:
//     len(mockedMessageProducer.OutputCalls())
func (mock *MessageProducerMock) OutputCalls() []struct {
} {
	var calls []struct {
	}
	lockMessageProducerMockOutput.RLock()
	calls = mock.calls.Output
	lockMessageProducerMockOutput.RUnlock()
	return calls
}
