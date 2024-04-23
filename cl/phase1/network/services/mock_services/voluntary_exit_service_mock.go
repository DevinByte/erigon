// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ledgerwatch/erigon/cl/phase1/network/services (interfaces: VoluntaryExitService)
//
// Generated by this command:
//
//	mockgen -destination=./mock_services/voluntary_exit_service_mock.go -package=mock_services . VoluntaryExitService
//

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	cltypes "github.com/ledgerwatch/erigon/cl/cltypes"
	gomock "go.uber.org/mock/gomock"
)

// MockVoluntaryExitService is a mock of VoluntaryExitService interface.
type MockVoluntaryExitService struct {
	ctrl     *gomock.Controller
	recorder *MockVoluntaryExitServiceMockRecorder
}

// MockVoluntaryExitServiceMockRecorder is the mock recorder for MockVoluntaryExitService.
type MockVoluntaryExitServiceMockRecorder struct {
	mock *MockVoluntaryExitService
}

// NewMockVoluntaryExitService creates a new mock instance.
func NewMockVoluntaryExitService(ctrl *gomock.Controller) *MockVoluntaryExitService {
	mock := &MockVoluntaryExitService{ctrl: ctrl}
	mock.recorder = &MockVoluntaryExitServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVoluntaryExitService) EXPECT() *MockVoluntaryExitServiceMockRecorder {
	return m.recorder
}

// ProcessMessage mocks base method.
func (m *MockVoluntaryExitService) ProcessMessage(arg0 context.Context, arg1 *uint64, arg2 *cltypes.SignedVoluntaryExit) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessMessage", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessMessage indicates an expected call of ProcessMessage.
func (mr *MockVoluntaryExitServiceMockRecorder) ProcessMessage(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessMessage", reflect.TypeOf((*MockVoluntaryExitService)(nil).ProcessMessage), arg0, arg1, arg2)
}