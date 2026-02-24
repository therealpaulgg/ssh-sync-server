package routes

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/testutils"
	"github.com/therealpaulgg/ssh-sync-server/test/pgx"
)

func TestInitialSetup(t *testing.T) {
	// Arrange
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("username", "test")
	_ = writer.WriteField("machine_name", "mymachine")
	priv, pub, err := testutils.GenerateTestKeys()
	if err != nil {
		t.Fatal(err)
	}
	pubBytes, _, err := testutils.EncodeToPem(priv, pub)
	if err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("key", "key")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write(pubBytes)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	injector := do.New()
	ctrl := gomock.NewController(t)
	mockTx := pgx.NewMockTx(ctrl)
	mockTransactionService := query.NewMockTransactionService(ctrl)
	mockTransactionService.EXPECT().StartTx(gomock.Any()).Return(mockTx, nil)
	mockTransactionService.EXPECT().Commit(mockTx).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTransactionService, nil
	})
	mockUserRepository := repository.NewMockUserRepository(ctrl)
	user := testutils.GenerateUser()
	machine := testutils.GenerateMachine()
	mockUserRepository.EXPECT().CreateUserTx(gomock.Any(), mockTx).Return(user, nil)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepository, nil
	})
	mockMachineRepository := repository.NewMockMachineRepository(ctrl)
	mockMachineRepository.EXPECT().CreateMachineTx(gomock.Any(), mockTx).Return(machine, nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepository, nil
	})
	// Act
	rr := httptest.NewRecorder()
	handler := initialSetup(injector)
	handler.ServeHTTP(rr, req)
	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("initialSetup returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestInitialSetup_MLDSA65(t *testing.T) {
	// Arrange
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("username", "test")
	_ = writer.WriteField("machine_name", "mymachine")
	pub, _, err := testutils.GenerateMLDSA65TestKeys()
	if err != nil {
		t.Fatal(err)
	}
	pubPEM, err := testutils.EncodeMLDSA65ToPem(pub)
	if err != nil {
		t.Fatal(err)
	}
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
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	injector := do.New()
	ctrl := gomock.NewController(t)
	mockTx := pgx.NewMockTx(ctrl)
	mockTransactionService := query.NewMockTransactionService(ctrl)
	mockTransactionService.EXPECT().StartTx(gomock.Any()).Return(mockTx, nil)
	mockTransactionService.EXPECT().Commit(mockTx).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTransactionService, nil
	})
	mockUserRepository := repository.NewMockUserRepository(ctrl)
	user := testutils.GenerateUser()
	machine := testutils.GenerateMachine()
	mockUserRepository.EXPECT().CreateUserTx(gomock.Any(), mockTx).Return(user, nil)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepository, nil
	})
	mockMachineRepository := repository.NewMockMachineRepository(ctrl)
	mockMachineRepository.EXPECT().CreateMachineTx(gomock.Any(), mockTx).Return(machine, nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepository, nil
	})
	// Act
	rr := httptest.NewRecorder()
	handler := initialSetup(injector)
	handler.ServeHTTP(rr, req)
	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("initialSetup returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// TODO non-happy-paths
