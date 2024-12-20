// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/database/repository/user.go

// Package repository is a generated GoMock package.
package repository

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	models "github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
)

// MockUserRepository is a mock of UserRepository interface.
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

// MockUserRepositoryMockRecorder is the mock recorder for MockUserRepository.
type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

// NewMockUserRepository creates a new mock instance.
func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

// AddAndUpdateConfig mocks base method.
func (m *MockUserRepository) AddAndUpdateConfig(user *models.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddAndUpdateConfig", user)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddAndUpdateConfig indicates an expected call of AddAndUpdateConfig.
func (mr *MockUserRepositoryMockRecorder) AddAndUpdateConfig(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddAndUpdateConfig", reflect.TypeOf((*MockUserRepository)(nil).AddAndUpdateConfig), user)
}

// AddAndUpdateConfigTx mocks base method.
func (m *MockUserRepository) AddAndUpdateConfigTx(user *models.User, tx pgx.Tx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddAndUpdateConfigTx", user, tx)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddAndUpdateConfigTx indicates an expected call of AddAndUpdateConfigTx.
func (mr *MockUserRepositoryMockRecorder) AddAndUpdateConfigTx(user, tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddAndUpdateConfigTx", reflect.TypeOf((*MockUserRepository)(nil).AddAndUpdateConfigTx), user, tx)
}

// AddAndUpdateKeys mocks base method.
func (m *MockUserRepository) AddAndUpdateKeys(user *models.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddAndUpdateKeys", user)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddAndUpdateKeys indicates an expected call of AddAndUpdateKeys.
func (mr *MockUserRepositoryMockRecorder) AddAndUpdateKeys(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddAndUpdateKeys", reflect.TypeOf((*MockUserRepository)(nil).AddAndUpdateKeys), user)
}

// AddAndUpdateKeysTx mocks base method.
func (m *MockUserRepository) AddAndUpdateKeysTx(user *models.User, tx pgx.Tx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddAndUpdateKeysTx", user, tx)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddAndUpdateKeysTx indicates an expected call of AddAndUpdateKeysTx.
func (mr *MockUserRepositoryMockRecorder) AddAndUpdateKeysTx(user, tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddAndUpdateKeysTx", reflect.TypeOf((*MockUserRepository)(nil).AddAndUpdateKeysTx), user, tx)
}

// CreateUser mocks base method.
func (m *MockUserRepository) CreateUser(user *models.User) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", user)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockUserRepositoryMockRecorder) CreateUser(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUserRepository)(nil).CreateUser), user)
}

// CreateUserTx mocks base method.
func (m *MockUserRepository) CreateUserTx(user *models.User, tx pgx.Tx) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUserTx", user, tx)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUserTx indicates an expected call of CreateUserTx.
func (mr *MockUserRepositoryMockRecorder) CreateUserTx(user, tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserTx", reflect.TypeOf((*MockUserRepository)(nil).CreateUserTx), user, tx)
}

// DeleteUser mocks base method.
func (m *MockUserRepository) DeleteUser(id uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockUserRepositoryMockRecorder) DeleteUser(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockUserRepository)(nil).DeleteUser), id)
}

// DeleteUserKeyTx mocks base method.
func (m *MockUserRepository) DeleteUserKeyTx(user *models.User, id uuid.UUID, tx pgx.Tx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUserKeyTx", user, id, tx)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUserKeyTx indicates an expected call of DeleteUserKeyTx.
func (mr *MockUserRepositoryMockRecorder) DeleteUserKeyTx(user, id, tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUserKeyTx", reflect.TypeOf((*MockUserRepository)(nil).DeleteUserKeyTx), user, id, tx)
}

// GetUser mocks base method.
func (m *MockUserRepository) GetUser(id uuid.UUID) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", id)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockUserRepositoryMockRecorder) GetUser(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockUserRepository)(nil).GetUser), id)
}

// GetUserByUsername mocks base method.
func (m *MockUserRepository) GetUserByUsername(username string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByUsername", username)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByUsername indicates an expected call of GetUserByUsername.
func (mr *MockUserRepositoryMockRecorder) GetUserByUsername(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByUsername", reflect.TypeOf((*MockUserRepository)(nil).GetUserByUsername), username)
}

// GetUserConfig mocks base method.
func (m *MockUserRepository) GetUserConfig(id uuid.UUID) ([]models.SshConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserConfig", id)
	ret0, _ := ret[0].([]models.SshConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserConfig indicates an expected call of GetUserConfig.
func (mr *MockUserRepositoryMockRecorder) GetUserConfig(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserConfig", reflect.TypeOf((*MockUserRepository)(nil).GetUserConfig), id)
}

// GetUserKey mocks base method.
func (m *MockUserRepository) GetUserKey(userId, keyId uuid.UUID) (*models.SshKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserKey", userId, keyId)
	ret0, _ := ret[0].(*models.SshKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserKey indicates an expected call of GetUserKey.
func (mr *MockUserRepositoryMockRecorder) GetUserKey(userId, keyId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserKey", reflect.TypeOf((*MockUserRepository)(nil).GetUserKey), userId, keyId)
}

// GetUserKeys mocks base method.
func (m *MockUserRepository) GetUserKeys(id uuid.UUID) ([]models.SshKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserKeys", id)
	ret0, _ := ret[0].([]models.SshKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserKeys indicates an expected call of GetUserKeys.
func (mr *MockUserRepositoryMockRecorder) GetUserKeys(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserKeys", reflect.TypeOf((*MockUserRepository)(nil).GetUserKeys), id)
}
