package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	testpgx "github.com/therealpaulgg/ssh-sync-server/test/pgx"
)

func TestGetSshConfig(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	injector := do.New()
	repo := &SshConfigRepo{
		Injector: injector,
	}

	// Test cases
	tests := []struct {
		name          string
		setupMock     func(*gomock.Controller)
		expectedError error
		expectedNil   bool
	}{
		{
			name: "successful query",
			setupMock: func(ctrl *gomock.Controller) {
				mockService := query.NewMockQueryService[models.SshConfig](ctrl)
				mockService.EXPECT().
					QueryOne(gomock.Any(), gomock.Any()).
					Return(&models.SshConfig{
						UserID: uuid.New(),
						Host:   "test-host",
						Values: map[string][]string{"key": {"value"}},
					}, nil)
				
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
					return mockService, nil
				})
			},
			expectedError: nil,
			expectedNil:   false,
		},
		{
			name: "no rows found",
			setupMock: func(ctrl *gomock.Controller) {
				mockService := query.NewMockQueryService[models.SshConfig](ctrl)
				mockService.EXPECT().
					QueryOne(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
					return mockService, nil
				})
			},
			expectedError: sql.ErrNoRows,
			expectedNil:   true,
		},
		{
			name: "database error",
			setupMock: func(ctrl *gomock.Controller) {
				mockService := query.NewMockQueryService[models.SshConfig](ctrl)
				mockService.EXPECT().
					QueryOne(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
				
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
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
			tc.setupMock(ctrl)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	injector := do.New()
	repo := &SshConfigRepo{
		Injector: injector,
	}

	// Test cases
	tests := []struct {
		name          string
		setupMock     func(*gomock.Controller)
		expectedError error
		expectedNil   bool
	}{
		{
			name: "successful upsert",
			setupMock: func(ctrl *gomock.Controller) {
				mockService := query.NewMockQueryService[models.SshConfig](ctrl)
				mockService.EXPECT().
					QueryOne(gomock.Any(), gomock.Any()).
					Return(&models.SshConfig{
						UserID: uuid.New(),
						Host:   "test-host",
						Values: map[string][]string{"key": {"value"}},
					}, nil)
				
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
					return mockService, nil
				})
			},
			expectedError: nil,
			expectedNil:   false,
		},
		{
			name: "no rows returned",
			setupMock: func(ctrl *gomock.Controller) {
				mockService := query.NewMockQueryService[models.SshConfig](ctrl)
				mockService.EXPECT().
					QueryOne(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
					return mockService, nil
				})
			},
			expectedError: sql.ErrNoRows,
			expectedNil:   true,
		},
		{
			name: "database error",
			setupMock: func(ctrl *gomock.Controller) {
				mockService := query.NewMockQueryService[models.SshConfig](ctrl)
				mockService.EXPECT().
					QueryOne(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
				
				do.Override(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
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
			tc.setupMock(ctrl)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	injector := do.New()
	repo := &SshConfigRepo{
		Injector: injector,
	}
	mockTx := testpgx.NewMockTx(ctrl)

	// Test cases
	tests := []struct {
		name          string
		setupMock     func(*gomock.Controller)
		expectedError error
		expectedNil   bool
	}{
		{
			name: "successful upsert with transaction",
			setupMock: func(ctrl *gomock.Controller) {
				mockService := query.NewMockQueryServiceTx[models.SshConfig](ctrl)
				mockService.EXPECT().
					QueryOne(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&models.SshConfig{
						UserID: uuid.New(),
						Host:   "test-host",
						Values: map[string][]string{"key": {"value"}},
					}, nil)
				
				do.Override(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshConfig], error) {
					return mockService, nil
				})
			},
			expectedError: nil,
			expectedNil:   false,
		},
		{
			name: "no rows returned with transaction",
			setupMock: func(ctrl *gomock.Controller) {
				mockService := query.NewMockQueryServiceTx[models.SshConfig](ctrl)
				mockService.EXPECT().
					QueryOne(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)
				
				do.Override(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshConfig], error) {
					return mockService, nil
				})
			},
			expectedError: sql.ErrNoRows,
			expectedNil:   true,
		},
		{
			name: "database error with transaction",
			setupMock: func(ctrl *gomock.Controller) {
				mockService := query.NewMockQueryServiceTx[models.SshConfig](ctrl)
				mockService.EXPECT().
					QueryOne(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
				
				do.Override(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshConfig], error) {
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
			tc.setupMock(ctrl)

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
