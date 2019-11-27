// Code generated by MockGen. DO NOT EDIT.
// Source: ./cache.go

// Package registry is a generated GoMock package.
package registry

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	model "github.com/samaritan-proxy/sash/model"
	reflect "reflect"
)

// MockCache is a mock of Cache interface
type MockCache struct {
	ctrl     *gomock.Controller
	recorder *MockCacheMockRecorder
}

// MockCacheMockRecorder is the mock recorder for MockCache
type MockCacheMockRecorder struct {
	mock *MockCache
}

// NewMockCache creates a new mock instance
func NewMockCache(ctrl *gomock.Controller) *MockCache {
	mock := &MockCache{ctrl: ctrl}
	mock.recorder = &MockCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCache) EXPECT() *MockCacheMockRecorder {
	return m.recorder
}

// Run mocks base method
func (m *MockCache) Run(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Run", ctx)
}

// Run indicates an expected call of Run
func (mr *MockCacheMockRecorder) Run(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockCache)(nil).Run), ctx)
}

// List mocks base method
func (m *MockCache) List() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockCacheMockRecorder) List() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockCache)(nil).List))
}

// Get mocks base method
func (m *MockCache) Get(name string) (*model.Service, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", name)
	ret0, _ := ret[0].(*model.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockCacheMockRecorder) Get(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCache)(nil).Get), name)
}

// Exists mocks base method
func (m *MockCache) Exists(name string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", name)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Exists indicates an expected call of Exists
func (mr *MockCacheMockRecorder) Exists(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockCache)(nil).Exists), name)
}

// RegisterServiceEventHandler mocks base method
func (m *MockCache) RegisterServiceEventHandler(handler ServiceEventHandler) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterServiceEventHandler", handler)
}

// RegisterServiceEventHandler indicates an expected call of RegisterServiceEventHandler
func (mr *MockCacheMockRecorder) RegisterServiceEventHandler(handler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterServiceEventHandler", reflect.TypeOf((*MockCache)(nil).RegisterServiceEventHandler), handler)
}

// RegisterInstanceEventHandler mocks base method
func (m *MockCache) RegisterInstanceEventHandler(handler InstanceEventHandler) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterInstanceEventHandler", handler)
}

// RegisterInstanceEventHandler indicates an expected call of RegisterInstanceEventHandler
func (mr *MockCacheMockRecorder) RegisterInstanceEventHandler(handler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterInstanceEventHandler", reflect.TypeOf((*MockCache)(nil).RegisterInstanceEventHandler), handler)
}
