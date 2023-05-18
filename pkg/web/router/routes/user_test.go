package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
)

func TestGetUser(t *testing.T) {
	// Arrange
	username := "test"
	req, err := http.NewRequest("GET", fmt.Sprintf("/%s", username), nil)
	if err != nil {
		t.Fatal(err)
	}
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserByUsername(username).Return(&models.User{
		Username: username,
	}, nil)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	// TODO: want to use this same pattern for data_test but the auth middleware is in the way
	UserRoutes(injector).ServeHTTP(rr, req)
	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var userDto dto.UserDto
	err = json.NewDecoder(rr.Body).Decode(&userDto)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, username, userDto.Username)
}

func TestUserNotFound(t *testing.T) {
	// Arrange
	username := "test"
	req, err := http.NewRequest("GET", fmt.Sprintf("/%s", username), nil)
	if err != nil {
		t.Fatal(err)
	}
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserByUsername(username).Return(nil, sql.ErrNoRows)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	UserRoutes(injector).ServeHTTP(rr, req)
	// Assert
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestUserInternalServerError(t *testing.T) {
	// Arrange
	username := "test"
	req, err := http.NewRequest("GET", fmt.Sprintf("/%s", username), nil)
	if err != nil {
		t.Fatal(err)
	}
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserByUsername(username).Return(nil, fmt.Errorf("error"))
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	UserRoutes(injector).ServeHTTP(rr, req)
	// Assert
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
