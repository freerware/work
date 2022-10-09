// Code generated by MockGen. DO NOT EDIT.
// Source: v4/unit_data_mapper.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	v4 "github.com/freerware/work/v4"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// UnitDataMapper is a mock of UnitDataMapper interface
type UnitDataMapper struct {
	ctrl     *gomock.Controller
	recorder *UnitDataMapperMockRecorder
}

// UnitDataMapperMockRecorder is the mock recorder for UnitDataMapper
type UnitDataMapperMockRecorder struct {
	mock *UnitDataMapper
}

// NewUnitDataMapper creates a new mock instance
func NewUnitDataMapper(ctrl *gomock.Controller) *UnitDataMapper {
	mock := &UnitDataMapper{ctrl: ctrl}
	mock.recorder = &UnitDataMapperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *UnitDataMapper) EXPECT() *UnitDataMapperMockRecorder {
	return m.recorder
}

// Insert mocks base method
func (m *UnitDataMapper) Insert(arg0 context.Context, arg1 v4.MapperContext, arg2 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Insert", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert
func (mr *UnitDataMapperMockRecorder) Insert(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*UnitDataMapper)(nil).Insert), varargs...)
}

// Update mocks base method
func (m *UnitDataMapper) Update(arg0 context.Context, arg1 v4.MapperContext, arg2 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *UnitDataMapperMockRecorder) Update(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*UnitDataMapper)(nil).Update), varargs...)
}

// Delete mocks base method
func (m *UnitDataMapper) Delete(arg0 context.Context, arg1 v4.MapperContext, arg2 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Delete", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *UnitDataMapperMockRecorder) Delete(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*UnitDataMapper)(nil).Delete), varargs...)
}
