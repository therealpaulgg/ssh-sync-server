// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/database/repository/machine.go

// Package repository is a generated GoMock package.
package repository

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	models "github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
)

// MockMachineRepository is a mock of MachineRepository interface.
type MockMachineRepository struct {
	ctrl     *gomock.Controller
	recorder *MockMachineRepositoryMockRecorder
}

// MockMachineRepositoryMockRecorder is the mock recorder for MockMachineRepository.
type MockMachineRepositoryMockRecorder struct {
	mock *MockMachineRepository
}

// NewMockMachineRepository creates a new mock instance.
func NewMockMachineRepository(ctrl *gomock.Controller) *MockMachineRepository {
	mock := &MockMachineRepository{ctrl: ctrl}
	mock.recorder = &MockMachineRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMachineRepository) EXPECT() *MockMachineRepositoryMockRecorder {
	return m.recorder
}

// CreateMachine mocks base method.
func (m *MockMachineRepository) CreateMachine(machine *models.Machine) (*models.Machine, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMachine", machine)
	ret0, _ := ret[0].(*models.Machine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMachine indicates an expected call of CreateMachine.
func (mr *MockMachineRepositoryMockRecorder) CreateMachine(machine interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMachine", reflect.TypeOf((*MockMachineRepository)(nil).CreateMachine), machine)
}

// CreateMachineTx mocks base method.
func (m *MockMachineRepository) CreateMachineTx(machine *models.Machine, tx pgx.Tx) (*models.Machine, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMachineTx", machine, tx)
	ret0, _ := ret[0].(*models.Machine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMachineTx indicates an expected call of CreateMachineTx.
func (mr *MockMachineRepositoryMockRecorder) CreateMachineTx(machine, tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMachineTx", reflect.TypeOf((*MockMachineRepository)(nil).CreateMachineTx), machine, tx)
}

// DeleteMachine mocks base method.
func (m *MockMachineRepository) DeleteMachine(id uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMachine", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMachine indicates an expected call of DeleteMachine.
func (mr *MockMachineRepositoryMockRecorder) DeleteMachine(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMachine", reflect.TypeOf((*MockMachineRepository)(nil).DeleteMachine), id)
}

// GetMachine mocks base method.
func (m *MockMachineRepository) GetMachine(id uuid.UUID) (*models.Machine, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMachine", id)
	ret0, _ := ret[0].(*models.Machine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMachine indicates an expected call of GetMachine.
func (mr *MockMachineRepositoryMockRecorder) GetMachine(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMachine", reflect.TypeOf((*MockMachineRepository)(nil).GetMachine), id)
}

// GetMachineByNameAndUser mocks base method.
func (m *MockMachineRepository) GetMachineByNameAndUser(machineName string, userID uuid.UUID) (*models.Machine, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMachineByNameAndUser", machineName, userID)
	ret0, _ := ret[0].(*models.Machine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMachineByNameAndUser indicates an expected call of GetMachineByNameAndUser.
func (mr *MockMachineRepositoryMockRecorder) GetMachineByNameAndUser(machineName, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMachineByNameAndUser", reflect.TypeOf((*MockMachineRepository)(nil).GetMachineByNameAndUser), machineName, userID)
}

// GetUserMachines mocks base method.
func (m *MockMachineRepository) GetUserMachines(id uuid.UUID) ([]models.Machine, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserMachines", id)
	ret0, _ := ret[0].([]models.Machine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserMachines indicates an expected call of GetUserMachines.
func (mr *MockMachineRepositoryMockRecorder) GetUserMachines(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserMachines", reflect.TypeOf((*MockMachineRepository)(nil).GetUserMachines), id)
}
