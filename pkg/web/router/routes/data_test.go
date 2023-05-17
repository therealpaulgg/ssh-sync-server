package routes

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"mime/multipart"
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
	"github.com/therealpaulgg/ssh-sync-server/test/pgx"
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

func TestGetDataError(t *testing.T) {
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
	mockQueryServiceSsh := query.NewMockQueryService[models.SshKey](ctrl)
	mockQueryServiceSsh.EXPECT().Query(gomock.Any(), user.ID).Return(nil, errors.New("You are bad"))
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKey], error) {
		return mockQueryServiceSsh, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getData(injector))
	handler.ServeHTTP(rr, req)

	// Assert

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("getData returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

func TestAddData(t *testing.T) {
	// Arrange
	// request needs to have multipart form data (generate fake bytes and add to request)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// file
	fakeFileBytes := make([]byte, 1024) // Adjust the size as needed
	_, _ = rand.Read(fakeFileBytes)
	part, err := writer.CreateFormFile("file", "test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write(fakeFileBytes)
	if err != nil {
		t.Fatal(err)
	}
	_ = writer.WriteField("ssh_config", `[{"host":"test"}]`)
	writer.Close()

	req, err := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err != nil {
		t.Fatal(err)
	}
	user := testutils.GenerateUser()
	machine := testutils.GenerateMachine()
	req = testutils.AddUserContext(req, user)
	req = testutils.AddMachineContext(req, machine)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockQueryServiceUser := query.NewMockQueryServiceTx[models.User](ctrl)
	txMock := pgx.NewMockTx(ctrl)
	mockQueryServiceUser.EXPECT().StartTx(gomock.Any()).Return(txMock, nil)
	mockQueryServiceUser.EXPECT().Commit(txMock).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.User], error) {
		return mockQueryServiceUser, nil
	})

	mockQueryServiceConfig := query.NewMockQueryServiceTx[models.SshConfig](ctrl)
	mockQueryServiceConfig.EXPECT().QueryOne(txMock, gomock.Any(), user.ID, machine.ID, gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.SshConfig{}, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshConfig], error) {
		return mockQueryServiceConfig, nil
	})
	mockQueryServiceKeys := query.NewMockQueryServiceTx[models.SshKey](ctrl)
	mockQueryServiceKeys.EXPECT().QueryOne(txMock, gomock.Any(), user.ID, gomock.Any(), gomock.Any()).Return(&models.SshKey{}, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKey], error) {
		return mockQueryServiceKeys, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(addData(injector))
	handler.ServeHTTP(rr, req)

	// Assert

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("addData returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestAddDataBadRequest(t *testing.T) {
	// Arrange
	// POST random bytes
	body := &bytes.Buffer{}
	_, _ = rand.Read(body.Bytes())
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	user := testutils.GenerateUser()
	machine := testutils.GenerateMachine()
	req = testutils.AddUserContext(req, user)
	req = testutils.AddMachineContext(req, machine)

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(addData(do.New()))
	handler.ServeHTTP(rr, req)
	// Assert

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("addData returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestAddDataError(t *testing.T) {
	// Arrange
	// request needs to have multipart form data (generate fake bytes and add to request)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// file
	fakeFileBytes := make([]byte, 1024) // Adjust the size as needed
	_, _ = rand.Read(fakeFileBytes)
	part, err := writer.CreateFormFile("file", "test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write(fakeFileBytes)
	if err != nil {
		t.Fatal(err)
	}
	_ = writer.WriteField("ssh_config", `[{"host":"test"}]`)
	writer.Close()

	req, err := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err != nil {
		t.Fatal(err)
	}
	user := testutils.GenerateUser()
	machine := testutils.GenerateMachine()
	req = testutils.AddUserContext(req, user)
	req = testutils.AddMachineContext(req, machine)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockQueryServiceUser := query.NewMockQueryServiceTx[models.User](ctrl)
	txMock := pgx.NewMockTx(ctrl)
	mockQueryServiceUser.EXPECT().StartTx(gomock.Any()).Return(txMock, nil)
	mockQueryServiceUser.EXPECT().Rollback(txMock).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.User], error) {
		return mockQueryServiceUser, nil
	})

	mockQueryServiceConfig := query.NewMockQueryServiceTx[models.SshConfig](ctrl)
	mockQueryServiceConfig.EXPECT().QueryOne(txMock, gomock.Any(), user.ID, machine.ID, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshConfig], error) {
		return mockQueryServiceConfig, nil
	})
	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(addData(injector))
	handler.ServeHTTP(rr, req)

	// Assert

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("addData returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
