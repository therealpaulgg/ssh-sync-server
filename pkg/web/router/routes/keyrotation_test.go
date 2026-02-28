package routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-common/pkg/dto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/testutils"
	pgxmock "github.com/therealpaulgg/ssh-sync-server/test/pgx"
)

func TestPostKeyRotation(t *testing.T) {
	// Arrange
	user := testutils.GenerateUser()
	machine1 := testutils.GenerateMachine()
	machine2 := testutils.GenerateMachine()

	reqBody := dto.MasterKeyRotationRequestDto{
		Keys: []dto.PerMachineMasterKeyDto{
			{MachineID: machine1.ID, EncryptedMasterKey: []byte("enc-key-1")},
			{MachineID: machine2.ID, EncryptedMasterKey: []byte("enc-key-2")},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetUserMachines(user.ID).Return([]models.Machine{
		{ID: machine1.ID, UserID: user.ID, Name: machine1.Name, PublicKey: machine1.PublicKey},
		{ID: machine2.ID, UserID: user.ID, Name: machine2.Name, PublicKey: machine2.PublicKey},
	}, nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	txMock := pgxmock.NewMockTx(ctrl)
	mockTxService := query.NewMockTransactionService(ctrl)
	mockTxService.EXPECT().StartTx(gomock.Any()).Return(txMock, nil)
	mockTxService.EXPECT().Commit(txMock).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTxService, nil
	})

	mockRotationRepo := repository.NewMockMasterKeyRotationRepository(ctrl)
	mockRotationRepo.EXPECT().UpsertRotationTx(txMock, machine1.ID, []byte("enc-key-1")).Return(nil)
	mockRotationRepo.EXPECT().UpsertRotationTx(txMock, machine2.ID, []byte("enc-key-2")).Return(nil)
	do.Provide(injector, func(i *do.Injector) (repository.MasterKeyRotationRepository, error) {
		return mockRotationRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(postKeyRotation(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestPostKeyRotation_BadRequest(t *testing.T) {
	// Arrange
	user := testutils.GenerateUser()
	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("not json")))
	req = testutils.AddUserContext(req, user)

	injector := do.New()

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(postKeyRotation(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestPostKeyRotation_ForbiddenMachineID(t *testing.T) {
	// Arrange
	user := testutils.GenerateUser()
	foreignMachineID := uuid.New() // belongs to a different user

	reqBody := dto.MasterKeyRotationRequestDto{
		Keys: []dto.PerMachineMasterKeyDto{
			{MachineID: foreignMachineID, EncryptedMasterKey: []byte("enc-key")},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// User owns no machines, so any machine ID will fail the ownership check.
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetUserMachines(user.ID).Return([]models.Machine{}, nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(postKeyRotation(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestPostKeyRotation_UpsertError(t *testing.T) {
	// Arrange
	user := testutils.GenerateUser()
	machine := testutils.GenerateMachine()

	reqBody := dto.MasterKeyRotationRequestDto{
		Keys: []dto.PerMachineMasterKeyDto{
			{MachineID: machine.ID, EncryptedMasterKey: []byte("enc-key")},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetUserMachines(user.ID).Return([]models.Machine{
		{ID: machine.ID, UserID: user.ID, Name: machine.Name, PublicKey: machine.PublicKey},
	}, nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	txMock := pgxmock.NewMockTx(ctrl)
	mockTxService := query.NewMockTransactionService(ctrl)
	mockTxService.EXPECT().StartTx(gomock.Any()).Return(txMock, nil)
	mockTxService.EXPECT().Rollback(txMock).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTxService, nil
	})

	mockRotationRepo := repository.NewMockMasterKeyRotationRepository(ctrl)
	mockRotationRepo.EXPECT().UpsertRotationTx(txMock, machine.ID, []byte("enc-key")).Return(errors.New("db error"))
	do.Provide(injector, func(i *do.Injector) (repository.MasterKeyRotationRepository, error) {
		return mockRotationRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(postKeyRotation(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetKeyRotation(t *testing.T) {
	// Arrange
	machine := testutils.GenerateMachine()
	encKey := []byte("encrypted-master-key")
	rotation := &models.MasterKeyRotation{
		ID:                 uuid.New(),
		MachineID:          machine.ID,
		EncryptedMasterKey: encKey,
		CreatedAt:          time.Now(),
	}

	req := httptest.NewRequest("GET", "/", nil)
	req = testutils.AddMachineContext(req, machine)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRotationRepo := repository.NewMockMasterKeyRotationRepository(ctrl)
	mockRotationRepo.EXPECT().GetRotationForMachine(machine.ID).Return(rotation, nil)
	do.Provide(injector, func(i *do.Injector) (repository.MasterKeyRotationRepository, error) {
		return mockRotationRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getKeyRotation(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	var rotDto dto.EncryptedMasterKeyDto
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&rotDto))
	assert.Equal(t, encKey, rotDto.EncryptedMasterKey)
}

func TestGetKeyRotation_NotFound(t *testing.T) {
	// Arrange
	machine := testutils.GenerateMachine()
	req := httptest.NewRequest("GET", "/", nil)
	req = testutils.AddMachineContext(req, machine)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRotationRepo := repository.NewMockMasterKeyRotationRepository(ctrl)
	mockRotationRepo.EXPECT().GetRotationForMachine(machine.ID).Return(nil, sql.ErrNoRows)
	do.Provide(injector, func(i *do.Injector) (repository.MasterKeyRotationRepository, error) {
		return mockRotationRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getKeyRotation(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetKeyRotation_Error(t *testing.T) {
	// Arrange
	machine := testutils.GenerateMachine()
	req := httptest.NewRequest("GET", "/", nil)
	req = testutils.AddMachineContext(req, machine)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRotationRepo := repository.NewMockMasterKeyRotationRepository(ctrl)
	mockRotationRepo.EXPECT().GetRotationForMachine(machine.ID).Return(nil, errors.New("db error"))
	do.Provide(injector, func(i *do.Injector) (repository.MasterKeyRotationRepository, error) {
		return mockRotationRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getKeyRotation(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestDeleteKeyRotation(t *testing.T) {
	// Arrange
	machine := testutils.GenerateMachine()
	req := httptest.NewRequest("DELETE", "/", nil)
	req = testutils.AddMachineContext(req, machine)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRotationRepo := repository.NewMockMasterKeyRotationRepository(ctrl)
	mockRotationRepo.EXPECT().DeleteRotationForMachine(machine.ID).Return(nil)
	do.Provide(injector, func(i *do.Injector) (repository.MasterKeyRotationRepository, error) {
		return mockRotationRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(deleteKeyRotation(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestDeleteKeyRotation_Error(t *testing.T) {
	// Arrange
	machine := testutils.GenerateMachine()
	req := httptest.NewRequest("DELETE", "/", nil)
	req = testutils.AddMachineContext(req, machine)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRotationRepo := repository.NewMockMasterKeyRotationRepository(ctrl)
	mockRotationRepo.EXPECT().DeleteRotationForMachine(machine.ID).Return(errors.New("db error"))
	do.Provide(injector, func(i *do.Injector) (repository.MasterKeyRotationRepository, error) {
		return mockRotationRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(deleteKeyRotation(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
