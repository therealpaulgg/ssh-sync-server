package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/testutils"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
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
