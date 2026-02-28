package live

import (
	"database/sql"
	"net"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	"github.com/therealpaulgg/ssh-sync-common/pkg/dto"
	"github.com/therealpaulgg/ssh-sync-common/pkg/wsutils"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/testutils"
)

func TestMachineChallengeResponseHandler_InvalidChallenge(t *testing.T) {
	ChallengeResponseDict = SafeChallengeResponseDict{dict: make(map[string]ChallengeSession)}
	injector := do.New()
	user := testutils.GenerateUser()
	req := httptest.NewRequest("GET", "/", nil)
	req = testutils.AddUserContext(req, user)

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		MachineChallengeResponseHandler(injector, req, httptest.NewRecorder(), &serverConn)
		close(done)
	}()

	require.NoError(t, wsutils.WriteClientMessage(&clientConn, dto.ChallengeResponseDto{Challenge: "missing"}))
	_, err := wsutils.ReadServerMessage[dto.ChallengeSuccessEncryptedKeyDto](&clientConn)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "Invalid challenge response."))
	<-done
}

func TestNewMachineChallengeHandler_UserNotFound(t *testing.T) {
	ChallengeResponseDict = SafeChallengeResponseDict{dict: make(map[string]ChallengeSession)}
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserByUsername("missing").Return(nil, sql.ErrNoRows)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	req := httptest.NewRequest("GET", "/", nil)
	done := make(chan struct{})
	go func() {
		NewMachineChallengeHandler(injector, req, httptest.NewRecorder(), &serverConn)
		close(done)
	}()

	require.NoError(t, wsutils.WriteClientMessage(&clientConn, dto.UserMachineDto{
		Username:    "missing",
		MachineName: "new",
	}))
	_, err := wsutils.ReadServerMessage[dto.MessageDto](&clientConn)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "User not found"))
	<-done
}

func TestNewMachineChallengeHandler_MachineExists(t *testing.T) {
	ChallengeResponseDict = SafeChallengeResponseDict{dict: make(map[string]ChallengeSession)}
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserByUsername("alice").Return(testutils.GenerateUser(), nil)
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetMachineByNameAndUser("laptop", gomock.Any()).Return(&models.Machine{ID: uuid.New()}, nil)
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	req := httptest.NewRequest("GET", "/", nil)
	done := make(chan struct{})
	go func() {
		NewMachineChallengeHandler(injector, req, httptest.NewRecorder(), &serverConn)
		close(done)
	}()

	require.NoError(t, wsutils.WriteClientMessage(&clientConn, dto.UserMachineDto{
		Username:    "alice",
		MachineName: "laptop",
	}))
	_, err := wsutils.ReadServerMessage[dto.MessageDto](&clientConn)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "Machine already exists"))
	<-done
}
