package routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-common/pkg/dto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/testutils"
)

func TestGetMachine(t *testing.T) {
	// Arrange
	machineId := uuid.New()
	req, err := http.NewRequest("GET", fmt.Sprintf("/%s", machineId), nil)
	if err != nil {
		t.Fatal(err)
	}
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)
	userMachines := []models.Machine{{
		ID:        machineId,
		UserID:    user.ID,
		Name:      "test",
		PublicKey: []byte("test"),
	}}

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetUserMachines(user.ID).Return(userMachines, nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})
	// Act
	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Get("/{machineId}", getMachineById(injector))
	router.ServeHTTP(rr, req)
	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	var machineDto dto.MachineDto
	err = json.NewDecoder(rr.Body).Decode(&machineDto)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, userMachines[0].Name, machineDto.Name)
}

func TestGetMachines(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)
	userMachines := []models.Machine{{
		ID:        uuid.New(),
		UserID:    user.ID,
		Name:      "test",
		PublicKey: []byte("test"),
	}}

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetUserMachines(user.ID).Return(userMachines, nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})
	// Act
	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Get("/", getMachines(injector))
	router.ServeHTTP(rr, req)
	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	var machineDtos []dto.MachineDto
	err = json.NewDecoder(rr.Body).Decode(&machineDtos)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(userMachines), len(machineDtos))
	assert.Equal(t, userMachines[0].Name, machineDtos[0].Name)
}

func TestDeleteMachine(t *testing.T) {
	// Arrange
	machineName := "test"
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(DeleteRequest{
		MachineName: machineName,
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("DELETE", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)
	if err != nil {
		t.Fatal(err)
	}
	machineId := uuid.New()
	userMachine := &models.Machine{
		ID:        machineId,
		UserID:    user.ID,
		Name:      machineName,
		PublicKey: []byte("test"),
	}

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetMachineByNameAndUser(machineName, user.ID).Return(userMachine, nil)
	mockMachineRepo.EXPECT().DeleteMachine(machineId).Return(nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})
	// Act
	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Delete("/", deleteMachine(injector))
	router.ServeHTTP(rr, req)
	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestUpdateMachineKey(t *testing.T) {
	// Arrange
	pub, _, err := testutils.GenerateMLDSATestKeys()
	if err != nil {
		t.Fatal(err)
	}
	pubPEM, err := testutils.EncodeMLDSAToPem(pub)
	if err != nil {
		t.Fatal(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("key", "key")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write(pubPEM)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("PUT", "/key", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	user := testutils.GenerateUser()
	machine := &models.Machine{
		ID:        uuid.New(),
		UserID:    user.ID,
		Name:      "test",
		PublicKey: []byte("old-key"),
	}
	req = testutils.AddUserContext(req, user)
	req = testutils.AddMachineContext(req, machine)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().UpdateMachinePublicKey(machine.ID, pubPEM).Return(nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Put("/key", updateMachineKey(injector))
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdateMachineKey_InvalidKey(t *testing.T) {
	// Arrange
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("key", "key")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write([]byte("not a valid key"))
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("PUT", "/key", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	user := testutils.GenerateUser()
	machine := &models.Machine{
		ID:        uuid.New(),
		UserID:    user.ID,
		Name:      "test",
		PublicKey: []byte("old-key"),
	}
	req = testutils.AddUserContext(req, user)
	req = testutils.AddMachineContext(req, machine)

	injector := do.New()

	// Act
	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Put("/key", updateMachineKey(injector))
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetMachine_InvalidId(t *testing.T) {
	req := httptest.NewRequest("GET", "/not-a-uuid", nil)
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetUserMachines(user.ID).Return([]models.Machine{}, nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Get("/{machineId}", getMachineById(injector))
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeleteMachine_NotFound(t *testing.T) {
	body := &bytes.Buffer{}
	_ = json.NewEncoder(body).Encode(DeleteRequest{MachineName: "missing"})
	req := httptest.NewRequest("DELETE", "/", body)
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetMachineByNameAndUser("missing", user.ID).Return(nil, sql.ErrNoRows)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Delete("/", deleteMachine(injector))
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateMachineKey_UpdateError(t *testing.T) {
	pub, _, err := testutils.GenerateMLDSATestKeys()
	if err != nil {
		t.Fatal(err)
	}
	pubPEM, err := testutils.EncodeMLDSAToPem(pub)
	if err != nil {
		t.Fatal(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("key", "key")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write(pubPEM); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("PUT", "/key", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	user := testutils.GenerateUser()
	machine := &models.Machine{
		ID:        uuid.New(),
		UserID:    user.ID,
		Name:      "test",
		PublicKey: []byte("old-key"),
	}
	req = testutils.AddUserContext(req, user)
	req = testutils.AddMachineContext(req, machine)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().UpdateMachinePublicKey(machine.ID, pubPEM).Return(errors.New("update failed"))
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Put("/key", updateMachineKey(injector))
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// TODO non-happy-paths
