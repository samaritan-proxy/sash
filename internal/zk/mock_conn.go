// Code generated by MockGen. DO NOT EDIT.
// Source: ./conn.go

// Package zk is a generated GoMock package.
package zk

import (
	gomock "github.com/golang/mock/gomock"
	zk "github.com/mesosphere/go-zookeeper/zk"
	reflect "reflect"
)

// MockConn is a mock of Conn interface
type MockConn struct {
	ctrl     *gomock.Controller
	recorder *MockConnMockRecorder
}

// MockConnMockRecorder is the mock recorder for MockConn
type MockConnMockRecorder struct {
	mock *MockConn
}

// NewMockConn creates a new mock instance
func NewMockConn(ctrl *gomock.Controller) *MockConn {
	mock := &MockConn{ctrl: ctrl}
	mock.recorder = &MockConnMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConn) EXPECT() *MockConnMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockConn) Get(path string) ([]byte, *zk.Stat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", path)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(*zk.Stat)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Get indicates an expected call of Get
func (mr *MockConnMockRecorder) Get(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockConn)(nil).Get), path)
}

// GetW mocks base method
func (m *MockConn) GetW(path string) ([]byte, *zk.Stat, <-chan zk.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetW", path)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(*zk.Stat)
	ret2, _ := ret[2].(<-chan zk.Event)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// GetW indicates an expected call of GetW
func (mr *MockConnMockRecorder) GetW(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetW", reflect.TypeOf((*MockConn)(nil).GetW), path)
}

// Children mocks base method
func (m *MockConn) Children(path string) ([]string, *zk.Stat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Children", path)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(*zk.Stat)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Children indicates an expected call of Children
func (mr *MockConnMockRecorder) Children(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Children", reflect.TypeOf((*MockConn)(nil).Children), path)
}

// ChildrenW mocks base method
func (m *MockConn) ChildrenW(path string) ([]string, *zk.Stat, <-chan zk.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChildrenW", path)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(*zk.Stat)
	ret2, _ := ret[2].(<-chan zk.Event)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ChildrenW indicates an expected call of ChildrenW
func (mr *MockConnMockRecorder) ChildrenW(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChildrenW", reflect.TypeOf((*MockConn)(nil).ChildrenW), path)
}

// Exists mocks base method
func (m *MockConn) Exists(path string) (bool, *zk.Stat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", path)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*zk.Stat)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Exists indicates an expected call of Exists
func (mr *MockConnMockRecorder) Exists(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockConn)(nil).Exists), path)
}

// CreateRecursively mocks base method
func (m *MockConn) CreateRecursively(p string, data []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRecursively", p, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateRecursively indicates an expected call of CreateRecursively
func (mr *MockConnMockRecorder) CreateRecursively(p, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRecursively", reflect.TypeOf((*MockConn)(nil).CreateRecursively), p, data)
}

// DeleteWithChildren mocks base method
func (m *MockConn) DeleteWithChildren(pathcur string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteWithChildren", pathcur)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteWithChildren indicates an expected call of DeleteWithChildren
func (mr *MockConnMockRecorder) DeleteWithChildren(pathcur interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteWithChildren", reflect.TypeOf((*MockConn)(nil).DeleteWithChildren), pathcur)
}

// Update mocks base method
func (m *MockConn) Update() <-chan zk.Event {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update")
	ret0, _ := ret[0].(<-chan zk.Event)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockConnMockRecorder) Update() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockConn)(nil).Update))
}

// Close mocks base method
func (m *MockConn) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close
func (mr *MockConnMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockConn)(nil).Close))
}
