package routes

import (
	"bytes"
	"crypto/rand"
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
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
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
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserKeys(user.ID).Return(data, nil)
	mockUserRepo.EXPECT().GetUserConfig(user.ID).Return(nil, nil)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
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

func TestGetDataErrorOnGetUserKeys(t *testing.T) {
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
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserKeys(user.ID).Return(nil, errors.New("You are bad"))
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
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

func TestGetDataErrorOnGetUserConfig(t *testing.T) {
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
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserKeys(user.ID).Return([]models.SshKey{}, nil)
	mockUserRepo.EXPECT().GetUserConfig(user.ID).Return(nil, errors.New("config error"))
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
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

func TestGetDataNoUserContext(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	// No user context added

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getData(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("getData with no user context returned wrong status code: got %v want %v",
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
	txMock := pgx.NewMockTx(ctrl)
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().AddAndUpdateConfigTx(gomock.Any(), txMock).Return(nil)
	mockUserRepo.EXPECT().AddAndUpdateKeysTx(gomock.Any(), txMock).Return(nil)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})
	mockTransactionService := query.NewMockTransactionService(ctrl)
	mockTransactionService.EXPECT().StartTx(gomock.Any()).Return(txMock, nil)
	mockTransactionService.EXPECT().Commit(txMock).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTransactionService, nil
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

func TestAddDataNoUserContext(t *testing.T) {
	// Arrange
	body := &bytes.Buffer{}
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	// No user context

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(addData(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("addData with no user context returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

func TestAddDataInvalidSshConfig(t *testing.T) {
	// Arrange
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Create invalid SSH config
	_ = writer.WriteField("ssh_config", `{"invalid": "json"`) // Invalid JSON
	writer.Close()

	req, err := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		t.Fatal(err)
	}
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(addData(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("addData with invalid SSH config returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestAddDataStartTxError(t *testing.T) {
	// Arrange
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("ssh_config", `[{"host":"test"}]`)
	writer.Close()

	req, err := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		t.Fatal(err)
	}
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})
	mockTransactionService := query.NewMockTransactionService(ctrl)
	mockTransactionService.EXPECT().StartTx(gomock.Any()).Return(nil, errors.New("tx error"))
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTransactionService, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(addData(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("addData with transaction error returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

func TestAddDataEmptySshConfig(t *testing.T) {
	// Arrange
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Empty SSH config
	writer.Close()

	req, err := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		t.Fatal(err)
	}
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(addData(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("addData with empty SSH config returned wrong status code: got %v want %v",
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
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	txMock := pgx.NewMockTx(ctrl)
	mockUserRepo.EXPECT().AddAndUpdateConfigTx(gomock.Any(), txMock).Return(nil)
	mockUserRepo.EXPECT().AddAndUpdateKeysTx(gomock.Any(), txMock).Return(errors.New("error"))
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})
	mockTransactionService := query.NewMockTransactionService(ctrl)
	mockTransactionService.EXPECT().StartTx(gomock.Any()).Return(txMock, nil)
	mockTransactionService.EXPECT().Rollback(txMock).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTransactionService, nil
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

func TestDeleteKey(t *testing.T) {
	// Arrange
	keyId := uuid.New()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/%s", keyId.String()), nil)
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)
	key := &models.SshKey{
		ID:     keyId,
		UserID: user.ID,
	}

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	txMock := pgx.NewMockTx(ctrl)
	mockUserRepo.EXPECT().GetUserKey(user.ID, keyId).Return(key, nil)
	mockUserRepo.EXPECT().DeleteUserKeyTx(gomock.Any(), keyId, txMock).Return(nil)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})
	mockTransactionService := query.NewMockTransactionService(ctrl)
	mockTransactionService.EXPECT().StartTx(gomock.Any()).Return(txMock, nil)
	mockTransactionService.EXPECT().Commit(txMock).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTransactionService, nil
	})
	// Act
	rr := httptest.NewRecorder()
	handler := chi.NewRouter()
	handler.Delete("/{id}", deleteData(injector))
	handler.ServeHTTP(rr, req)
	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("deleteData returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestDeleteKeyNoUserContext(t *testing.T) {
	// Arrange
	keyId := uuid.New()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/%s", keyId.String()), nil)
	// No user context

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Act
	rr := httptest.NewRecorder()
	handler := chi.NewRouter()
	handler.Delete("/{id}", deleteData(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("deleteData with no user context returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

func TestDeleteKeyInvalidUUID(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("DELETE", "/invalid-uuid", nil)
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Act
	rr := httptest.NewRecorder()
	handler := chi.NewRouter()
	handler.Delete("/{id}", deleteData(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("deleteData with invalid UUID returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestDeleteKeyKeyNotFound(t *testing.T) {
	// Arrange
	keyId := uuid.New()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/%s", keyId.String()), nil)
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserKey(user.ID, keyId).Return(nil, errors.New("key not found"))
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := chi.NewRouter()
	handler.Delete("/{id}", deleteData(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("deleteData with key not found returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestDeleteKeyTxStartError(t *testing.T) {
	// Arrange
	keyId := uuid.New()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/%s", keyId.String()), nil)
	user := testutils.GenerateUser()
	req = testutils.AddUserContext(req, user)
	key := &models.SshKey{
		ID:     keyId,
		UserID: user.ID,
	}

	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserKey(user.ID, keyId).Return(key, nil)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})
	mockTransactionService := query.NewMockTransactionService(ctrl)
	mockTransactionService.EXPECT().StartTx(gomock.Any()).Return(nil, errors.New("tx start error"))
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTransactionService, nil
	})

	// Act
	rr := httptest.NewRecorder()
	handler := chi.NewRouter()
	handler.Delete("/{id}", deleteData(injector))
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("deleteData with tx start error returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}
