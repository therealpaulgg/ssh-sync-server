package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	testpgx "github.com/therealpaulgg/ssh-sync-server/test/pgx"
)

// Create a mock QueryService
type mockQueryService struct {
	mockQueryOne func(query string, args ...interface{}) (*models.SshConfig, error)
}

func (m *mockQueryService) Query(query string, args ...interface{}) ([]models.SshConfig, error) {
	return nil, nil
}

func (m *mockQueryService) QueryOne(query string, args ...interface{}) (*models.SshConfig, error) {
	return m.mockQueryOne(query, args...)
}

func (m *mockQueryService) Insert(query string, args ...interface{}) error {
	return nil
}

// Create a mock QueryServiceTx
type mockQueryServiceTx struct {
	mockQueryOne func(tx pgx.Tx, query string, args ...interface{}) (*models.SshConfig, error)
}

func (m *mockQueryServiceTx) Query(tx pgx.Tx, query string, args ...interface{}) ([]models.SshConfig, error) {
	return nil, nil
}

func (m *mockQueryServiceTx) QueryOne(tx pgx.Tx, query string, args ...interface{}) (*models.SshConfig, error) {
	return m.mockQueryOne(tx, query, args...)
}

func (m *mockQueryServiceTx) Insert(tx pgx.Tx, query string, args ...interface{}) error {
	return nil
}

func TestGetSshConfig(t *testing.T) {
	// Setup
	injector := do.New()
	repo := &SshConfigRepo{
		Injector: injector,
	}

	// Test cases
	tests := []struct {
		name          string
		setupMock     func()
		expectedError error
		expectedNil   bool
	}{
		{
			name: "successful query",
			setupMock: func() {
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
					mockService := &mockQueryService{
						mockQueryOne: func(query string, args ...interface{}) (*models.SshConfig, error) {
							return &models.SshConfig{
								UserID: uuid.New(),
								Host:   "test-host",
								Values: map[string][]string{"key": {"value"}},
							}, nil
						},
					}
					return mockService, nil
				})
			},
			expectedError: nil,
			expectedNil:   false,
		},
		{
			name: "no rows found",
			setupMock: func() {
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
					mockService := &mockQueryService{
						mockQueryOne: func(query string, args ...interface{}) (*models.SshConfig, error) {
							return nil, nil
						},
					}
					return mockService, nil
				})
			},
			expectedError: sql.ErrNoRows,
			expectedNil:   true,
		},
		{
			name: "database error",
			setupMock: func() {
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
					mockService := &mockQueryService{
						mockQueryOne: func(query string, args ...interface{}) (*models.SshConfig, error) {
							return nil, errors.New("database error")
						},
					}
					return mockService, nil
				})
			},
			expectedError: errors.New("database error"),
			expectedNil:   true,
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Run test
			result, err := repo.GetSshConfig(uuid.New())

			// Verify results
			if tc.expectedError != nil {
				if tc.expectedError == sql.ErrNoRows {
					assert.Equal(t, tc.expectedError, err)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if tc.expectedNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestUpsertSshConfig(t *testing.T) {
	// Setup
	injector := do.New()
	repo := &SshConfigRepo{
		Injector: injector,
	}

	// Test cases
	tests := []struct {
		name          string
		setupMock     func()
		expectedError error
		expectedNil   bool
	}{
		{
			name: "successful upsert",
			setupMock: func() {
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
					mockService := &mockQueryService{
						mockQueryOne: func(query string, args ...interface{}) (*models.SshConfig, error) {
							return &models.SshConfig{
								UserID: uuid.New(),
								Host:   "test-host",
								Values: map[string][]string{"key": {"value"}},
							}, nil
						},
					}
					return mockService, nil
				})
			},
			expectedError: nil,
			expectedNil:   false,
		},
		{
			name: "no rows returned",
			setupMock: func() {
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
					mockService := &mockQueryService{
						mockQueryOne: func(query string, args ...interface{}) (*models.SshConfig, error) {
							return nil, nil
						},
					}
					return mockService, nil
				})
			},
			expectedError: sql.ErrNoRows,
			expectedNil:   true,
		},
		{
			name: "database error",
			setupMock: func() {
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
					mockService := &mockQueryService{
						mockQueryOne: func(query string, args ...interface{}) (*models.SshConfig, error) {
							return nil, errors.New("database error")
						},
					}
					return mockService, nil
				})
			},
			expectedError: errors.New("database error"),
			expectedNil:   true,
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Run test
			config := &models.SshConfig{
				UserID:        uuid.New(),
				Host:          "test-host",
				Values:        map[string][]string{"key": {"value"}},
				IdentityFiles: []string{"id_rsa"},
			}
			result, err := repo.UpsertSshConfig(config)

			// Verify results
			if tc.expectedError != nil {
				if tc.expectedError == sql.ErrNoRows {
					assert.Equal(t, tc.expectedError, err)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if tc.expectedNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestUpsertSshConfigTx(t *testing.T) {
	// Setup
	injector := do.New()
	repo := &SshConfigRepo{
		Injector: injector,
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTx := testpgx.NewMockTx(ctrl)

	// Test cases
	tests := []struct {
		name          string
		setupMock     func()
		expectedError error
		expectedNil   bool
	}{
		{
			name: "successful upsert with transaction",
			setupMock: func() {
				do.Override(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshConfig], error) {
					mockService := &mockQueryServiceTx{
						mockQueryOne: func(tx pgx.Tx, query string, args ...interface{}) (*models.SshConfig, error) {
							return &models.SshConfig{
								UserID: uuid.New(),
								Host:   "test-host",
								Values: map[string][]string{"key": {"value"}},
							}, nil
						},
					}
					return mockService, nil
				})
			},
			expectedError: nil,
			expectedNil:   false,
		},
		{
			name: "no rows returned with transaction",
			setupMock: func() {
				do.Override(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshConfig], error) {
					mockService := &mockQueryServiceTx{
						mockQueryOne: func(tx pgx.Tx, query string, args ...interface{}) (*models.SshConfig, error) {
							return nil, nil
						},
					}
					return mockService, nil
				})
			},
			expectedError: sql.ErrNoRows,
			expectedNil:   true,
		},
		{
			name: "database error with transaction",
			setupMock: func() {
				do.Override(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshConfig], error) {
					mockService := &mockQueryServiceTx{
						mockQueryOne: func(tx pgx.Tx, query string, args ...interface{}) (*models.SshConfig, error) {
							return nil, errors.New("database error")
						},
					}
					return mockService, nil
				})
			},
			expectedError: errors.New("database error"),
			expectedNil:   true,
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Run test
			config := &models.SshConfig{
				UserID:        uuid.New(),
				Host:          "test-host",
				Values:        map[string][]string{"key": {"value"}},
				IdentityFiles: []string{"id_rsa"},
			}
			result, err := repo.UpsertSshConfigTx(config, mockTx)

			// Verify results
			if tc.expectedError != nil {
				if tc.expectedError == sql.ErrNoRows {
					assert.Equal(t, tc.expectedError, err)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if tc.expectedNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}
