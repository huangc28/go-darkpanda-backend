// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/huangchihan/Documents/go-darkpanda-backend/internal/app/contracts/service.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	contracts "github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	models "github.com/huangc28/go-darkpanda-backend/internal/app/models"
	reflect "reflect"
)

// MockServiceDAOer is a mock of ServiceDAOer interface
type MockServiceDAOer struct {
	ctrl     *gomock.Controller
	recorder *MockServiceDAOerMockRecorder
}

// MockServiceDAOerMockRecorder is the mock recorder for MockServiceDAOer
type MockServiceDAOerMockRecorder struct {
	mock *MockServiceDAOer
}

// NewMockServiceDAOer creates a new mock instance
func NewMockServiceDAOer(ctrl *gomock.Controller) *MockServiceDAOer {
	mock := &MockServiceDAOer{ctrl: ctrl}
	mock.recorder = &MockServiceDAOerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockServiceDAOer) EXPECT() *MockServiceDAOerMockRecorder {
	return m.recorder
}

// GetUserHistoricalServicesByUuid mocks base method
func (m *MockServiceDAOer) GetUserHistoricalServicesByUuid(uuid string, perPage, offset int) ([]models.Service, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserHistoricalServicesByUuid", uuid, perPage, offset)
	ret0, _ := ret[0].([]models.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserHistoricalServicesByUuid indicates an expected call of GetUserHistoricalServicesByUuid
func (mr *MockServiceDAOerMockRecorder) GetUserHistoricalServicesByUuid(uuid, perPage, offset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserHistoricalServicesByUuid", reflect.TypeOf((*MockServiceDAOer)(nil).GetUserHistoricalServicesByUuid), uuid, perPage, offset)
}

// GetServiceByInquiryUUID mocks base method
func (m *MockServiceDAOer) GetServiceByInquiryUUID(uuid string) (*models.Service, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServiceByInquiryUUID", uuid)
	ret0, _ := ret[0].(*models.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServiceByInquiryUUID indicates an expected call of GetServiceByInquiryUUID
func (mr *MockServiceDAOerMockRecorder) GetServiceByInquiryUUID(uuid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServiceByInquiryUUID", reflect.TypeOf((*MockServiceDAOer)(nil).GetServiceByInquiryUUID), uuid)
}

// UpdateServiceByID mocks base method
func (m *MockServiceDAOer) UpdateServiceByID(params contracts.UpdateServiceByIDParams) (*models.Service, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateServiceByID", params)
	ret0, _ := ret[0].(*models.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateServiceByID indicates an expected call of UpdateServiceByID
func (mr *MockServiceDAOerMockRecorder) UpdateServiceByID(params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateServiceByID", reflect.TypeOf((*MockServiceDAOer)(nil).UpdateServiceByID), params)
}
