package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/testutils"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
)

func TestGetData(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sshId := uuid.New()
	bytes := []byte("test")
	data := []models.SshKey{{
		ID:       sshId,
		UserID:   user.ID,
		Filename: "test",
		Data:     bytes,
	}}
	mockQueryServiceSsh := query.NewMockQueryService[models.SshKey](ctrl)
	mockQueryServiceSsh.EXPECT().Query(gomock.Any(), user.ID).Return(data, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKey], error) {
		return mockQueryServiceSsh, nil
	})

	mockQueryServiceConfig := query.NewMockQueryService[models.SshConfig](ctrl)
	mockQueryServiceConfig.EXPECT().Query(gomock.Any(), user.ID).Return(nil, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
		return mockQueryServiceConfig, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getData(injector))
	handler.ServeHTTP(rr, req)

	// Assert

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("getData returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var dataDto dto.DataDto
	err = json.NewDecoder(rr.Body).Decode(&dataDto)
	if err != nil {
		t.Errorf("getData returned unexpected body: got %v want %v, could not decode",
			rr.Body.String(), err)
	}

	assert.Equal(t, user.ID, dataDto.ID)
	assert.Equal(t, "test", dataDto.Keys[0].Filename)
	assert.Equal(t, bytes, dataDto.Keys[0].Data)
	assert.Equal(t, 0, len(dataDto.SshConfig))
}
