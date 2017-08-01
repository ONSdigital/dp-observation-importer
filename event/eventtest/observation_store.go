// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package eventtest

import (
	"github.com/ONSdigital/dp-observation-importer/observation"
	"sync"
)

var (
	lockObservationStoreMockSaveAll sync.RWMutex
)

// ObservationStoreMock is a mock implementation of ObservationStore.
//
//     func TestSomethingThatUsesObservationStore(t *testing.T) {
//
//         // make and configure a mocked ObservationStore
//         mockedObservationStore := &ObservationStoreMock{
//             SaveAllFunc: func(observations []*observation.Observation) ([]*observation.Result, error) {
// 	               panic("TODO: mock out the SaveAll method")
//             },
//         }
//
//         // TODO: use mockedObservationStore in code that requires ObservationStore
//         //       and then make assertions.
//
//     }
type ObservationStoreMock struct {
	// SaveAllFunc mocks the SaveAll method.
	SaveAllFunc func(observations []*observation.Observation) ([]*observation.Result, error)

	// calls tracks calls to the methods.
	calls struct {
		// SaveAll holds details about calls to the SaveAll method.
		SaveAll []struct {
			// Observations is the observations argument value.
			Observations []*observation.Observation
		}
	}
}

// SaveAll calls SaveAllFunc.
func (mock *ObservationStoreMock) SaveAll(observations []*observation.Observation) ([]*observation.Result, error) {
	if mock.SaveAllFunc == nil {
		panic("moq: ObservationStoreMock.SaveAllFunc is nil but ObservationStore.SaveAll was just called")
	}
	callInfo := struct {
		Observations []*observation.Observation
	}{
		Observations: observations,
	}
	lockObservationStoreMockSaveAll.Lock()
	mock.calls.SaveAll = append(mock.calls.SaveAll, callInfo)
	lockObservationStoreMockSaveAll.Unlock()
	return mock.SaveAllFunc(observations)
}

// SaveAllCalls gets all the calls that were made to SaveAll.
// Check the length with:
//     len(mockedObservationStore.SaveAllCalls())
func (mock *ObservationStoreMock) SaveAllCalls() []struct {
	Observations []*observation.Observation
} {
	var calls []struct {
		Observations []*observation.Observation
	}
	lockObservationStoreMockSaveAll.RLock()
	calls = mock.calls.SaveAll
	lockObservationStoreMockSaveAll.RUnlock()
	return calls
}
