// Code generated by MockGen. DO NOT EDIT.
// Source: order_cache.go

// Package dimensiontest is a generated GoMock package.
package dimensiontest

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockOrderStore is a mock of OrderStore interface
type MockOrderStore struct {
	ctrl     *gomock.Controller
	recorder *MockOrderStoreMockRecorder
}

// MockOrderStoreMockRecorder is the mock recorder for MockOrderStore
type MockOrderStoreMockRecorder struct {
	mock *MockOrderStore
}

// NewMockOrderStore creates a new mock instance
func NewMockOrderStore(ctrl *gomock.Controller) *MockOrderStore {
	mock := &MockOrderStore{ctrl: ctrl}
	mock.recorder = &MockOrderStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockOrderStore) EXPECT() *MockOrderStoreMockRecorder {
	return m.recorder
}

// GetOrder mocks base method
func (m *MockOrderStore) GetOrder(ctx context.Context, instanceID string) ([]string, error) {
	ret := m.ctrl.Call(m, "GetOrder", ctx, instanceID)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrder indicates an expected call of GetOrder
func (mr *MockOrderStoreMockRecorder) GetOrder(ctx, instanceID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrder", reflect.TypeOf((*MockOrderStore)(nil).GetOrder), ctx, instanceID)
}
