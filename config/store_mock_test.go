// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/samaritan-proxy/sash/config (interfaces: Store,SubscribableStore)

// Package config is a generated GoMock package.
package config

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockStore is a mock of Store interface
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// Del mocks base method
func (m *MockStore) Del(arg0, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Del", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Del indicates an expected call of Del
func (mr *MockStoreMockRecorder) Del(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del", reflect.TypeOf((*MockStore)(nil).Del), arg0, arg1, arg2)
}

// Exist mocks base method
func (m *MockStore) Exist(arg0, arg1, arg2 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exist", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Exist indicates an expected call of Exist
func (mr *MockStoreMockRecorder) Exist(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exist", reflect.TypeOf((*MockStore)(nil).Exist), arg0, arg1, arg2)
}

// Get mocks base method
func (m *MockStore) Get(arg0, arg1, arg2 string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockStoreMockRecorder) Get(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStore)(nil).Get), arg0, arg1, arg2)
}

// GetKeys mocks base method
func (m *MockStore) GetKeys(arg0, arg1 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKeys", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKeys indicates an expected call of GetKeys
func (mr *MockStoreMockRecorder) GetKeys(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKeys", reflect.TypeOf((*MockStore)(nil).GetKeys), arg0, arg1)
}

// Set mocks base method
func (m *MockStore) Set(arg0, arg1, arg2 string, arg3 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set
func (mr *MockStoreMockRecorder) Set(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockStore)(nil).Set), arg0, arg1, arg2, arg3)
}

// Start mocks base method
func (m *MockStore) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start
func (mr *MockStoreMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockStore)(nil).Start))
}

// Stop mocks base method
func (m *MockStore) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockStoreMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockStore)(nil).Stop))
}

// MockSubscribableStore is a mock of SubscribableStore interface
type MockSubscribableStore struct {
	ctrl     *gomock.Controller
	recorder *MockSubscribableStoreMockRecorder
}

// MockSubscribableStoreMockRecorder is the mock recorder for MockSubscribableStore
type MockSubscribableStoreMockRecorder struct {
	mock *MockSubscribableStore
}

// NewMockSubscribableStore creates a new mock instance
func NewMockSubscribableStore(ctrl *gomock.Controller) *MockSubscribableStore {
	mock := &MockSubscribableStore{ctrl: ctrl}
	mock.recorder = &MockSubscribableStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSubscribableStore) EXPECT() *MockSubscribableStoreMockRecorder {
	return m.recorder
}

// Del mocks base method
func (m *MockSubscribableStore) Del(arg0, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Del", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Del indicates an expected call of Del
func (mr *MockSubscribableStoreMockRecorder) Del(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del", reflect.TypeOf((*MockSubscribableStore)(nil).Del), arg0, arg1, arg2)
}

// Event mocks base method
func (m *MockSubscribableStore) Event() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Event")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Event indicates an expected call of Event
func (mr *MockSubscribableStoreMockRecorder) Event() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Event", reflect.TypeOf((*MockSubscribableStore)(nil).Event))
}

// Exist mocks base method
func (m *MockSubscribableStore) Exist(arg0, arg1, arg2 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exist", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Exist indicates an expected call of Exist
func (mr *MockSubscribableStoreMockRecorder) Exist(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exist", reflect.TypeOf((*MockSubscribableStore)(nil).Exist), arg0, arg1, arg2)
}

// Get mocks base method
func (m *MockSubscribableStore) Get(arg0, arg1, arg2 string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockSubscribableStoreMockRecorder) Get(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSubscribableStore)(nil).Get), arg0, arg1, arg2)
}

// GetKeys mocks base method
func (m *MockSubscribableStore) GetKeys(arg0, arg1 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKeys", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKeys indicates an expected call of GetKeys
func (mr *MockSubscribableStoreMockRecorder) GetKeys(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKeys", reflect.TypeOf((*MockSubscribableStore)(nil).GetKeys), arg0, arg1)
}

// Set mocks base method
func (m *MockSubscribableStore) Set(arg0, arg1, arg2 string, arg3 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set
func (mr *MockSubscribableStoreMockRecorder) Set(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockSubscribableStore)(nil).Set), arg0, arg1, arg2, arg3)
}

// Start mocks base method
func (m *MockSubscribableStore) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start
func (mr *MockSubscribableStoreMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockSubscribableStore)(nil).Start))
}

// Stop mocks base method
func (m *MockSubscribableStore) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockSubscribableStoreMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockSubscribableStore)(nil).Stop))
}

// Subscribe mocks base method
func (m *MockSubscribableStore) Subscribe(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Subscribe indicates an expected call of Subscribe
func (mr *MockSubscribableStoreMockRecorder) Subscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockSubscribableStore)(nil).Subscribe), arg0)
}

// UnSubscribe mocks base method
func (m *MockSubscribableStore) UnSubscribe(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnSubscribe", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UnSubscribe indicates an expected call of UnSubscribe
func (mr *MockSubscribableStoreMockRecorder) UnSubscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnSubscribe", reflect.TypeOf((*MockSubscribableStore)(nil).UnSubscribe), arg0)
}
