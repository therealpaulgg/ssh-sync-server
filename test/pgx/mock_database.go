package pgx

import (
	"github.com/jackc/pgx/v5"
)

// MockDatabase is a mock implementation of database for testing
type MockDatabase struct {
	MockQuery     func(sql string, args ...interface{}) [][]interface{}
	MockQueryRow  func(sql string, args ...interface{}) []interface{}
	MockTxQueryRow func(tx pgx.Tx, sql string, args ...interface{}) []interface{}
}

// NewMockDatabase creates a new MockDatabase with default implementations
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		MockQuery: func(sql string, args ...interface{}) [][]interface{} {
			return [][]interface{}{}
		},
		MockQueryRow: func(sql string, args ...interface{}) []interface{} {
			return nil
		},
		MockTxQueryRow: func(tx pgx.Tx, sql string, args ...interface{}) []interface{} {
			return nil
		},
	}
}