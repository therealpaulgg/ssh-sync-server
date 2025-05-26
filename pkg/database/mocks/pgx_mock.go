package mocks

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
)

// PgxTxInterface defines interface methods needed for mocking pgx.Tx
type PgxTxInterface interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Exec(ctx context.Context, sql string, args ...interface{}) (pgx.CommandTag, error)
}

// PgxConnInterface defines interface methods needed for mocking pgx.Conn
type PgxConnInterface interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// MockPgxTx is a mock implementation of the pgx.Tx interface for testing
type MockPgxTx struct {
	ctrl     *gomock.Controller
	recorder *MockPgxTxMockRecorder
}

type MockPgxTxMockRecorder struct {
	mock *MockPgxTx
}

func NewMockPgxTx(ctrl *gomock.Controller) *MockPgxTx {
	mock := &MockPgxTx{ctrl: ctrl}
	mock.recorder = &MockPgxTxMockRecorder{mock}
	return mock
}

func (m *MockPgxTx) EXPECT() *MockPgxTxMockRecorder {
	return m.recorder
}

func (m *MockPgxTx) Commit(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockPgxTxMockRecorder) Commit(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", nil, ctx)
}

func (m *MockPgxTx) Rollback(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rollback", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockPgxTxMockRecorder) Rollback(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", nil, ctx)
}

func (m *MockPgxTx) Exec(ctx context.Context, sql string, args ...interface{}) (pgx.CommandTag, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, sql}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	// We're creating a mock CommandTag
	var cmdTag pgx.CommandTag
	ret1, _ := ret[1].(error)
	return cmdTag, ret1
}

func (mr *MockPgxTxMockRecorder) Exec(ctx, sql interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{ctx, sql}
	varargs = append(varargs, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", nil, varargs...)
}

// MockPgxConn is a mock implementation of the relevant parts of pgx.Conn for testing
type MockPgxConn struct {
	ctrl     *gomock.Controller
	recorder *MockPgxConnMockRecorder
}

type MockPgxConnMockRecorder struct {
	mock *MockPgxConn
}

func NewMockPgxConn(ctrl *gomock.Controller) *MockPgxConn {
	mock := &MockPgxConn{ctrl: ctrl}
	mock.recorder = &MockPgxConnMockRecorder{mock}
	return mock
}

func (m *MockPgxConn) EXPECT() *MockPgxConnMockRecorder {
	return m.recorder
}

func (m *MockPgxConn) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginTx", ctx, txOptions)
	ret0, _ := ret[0].(pgx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPgxConnMockRecorder) BeginTx(ctx, txOptions interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTx", nil, ctx, txOptions)
}