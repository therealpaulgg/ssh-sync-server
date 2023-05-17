package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
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
	mockQueryServiceUser := query.NewMockQueryService[models.User](ctrl)
	mockQueryServiceUser.EXPECT().QueryOne(gomock.Any(), username).Return(&models.User{
		Username: username,
	}, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.User], error) {
		return mockQueryServiceUser, nil
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
