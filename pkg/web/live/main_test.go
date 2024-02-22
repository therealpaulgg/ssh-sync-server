package live

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi"
	"github.com/gobwas/ws"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware/context_keys"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
	"github.com/therealpaulgg/ssh-sync/pkg/utils"
)

type HijackableResponseRecorder struct {
	*httptest.ResponseRecorder
	HijackFunc func() (net.Conn, *bufio.ReadWriter, error)
}

func (h *HijackableResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h.HijackFunc != nil {
		return h.HijackFunc()
	}
	// Return a mock connection or an error if Hijack is not expected to be called
	return nil, nil, fmt.Errorf("hijack not implemented")
}

func TestNewMachineChallenge_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	injector := do.New()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockConn := NewMockConn(ctrl) // Assuming you have a mock of net.Conn
	// Setup dependency injection for repositories
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	// Setup expectations for repository interactions
	mockUserRepo.EXPECT().GetUserByUsername(gomock.Any()).Return(&models.User{ID: uuid.New(), Username: "testuser"}, nil).AnyTimes()
	mockMachineRepo.EXPECT().GetMachineByNameAndUser(gomock.Any(), gomock.Any()).Return(nil, sql.ErrNoRows).AnyTimes()
	mockMachineRepo.EXPECT().CreateMachine(gomock.Any()).Return(&models.Machine{Name: "testmachine"}, nil).AnyTimes()
	mockConn.EXPECT().SetDeadline(gomock.Any()).AnyTimes()
	mockConn.EXPECT().Write(gomock.Any()).DoAndReturn(func(data []byte) (int, error) {
		// Simulate writing the data and return the length of the data as if it was written successfully
		return len(data), nil
	}).AnyTimes()
	mockConn.EXPECT().Read(gomock.Any()).DoAndReturn(func(data []byte) (int, error) {
		// Simulate reading the data and return the length of the data as if it was read successfully
		return len(data), nil
	}).AnyTimes()
	mockConn.EXPECT().Close().AnyTimes()

	// Create a request and recorder
	req := httptest.NewRequest("GET", "/newMachineChallenge", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13") // WebSocket version, 13 is the current version
	req = req.WithContext(context.WithValue(req.Context(), context_keys.UserContextKey, &models.User{Username: "testuser"}))

	hijackFunc := func() (net.Conn, *bufio.ReadWriter, error) {
		// Return your mocked net.Conn, and a bufio.ReadWriter if needed
		rw := bufio.NewReadWriter(bufio.NewReader(mockConn), bufio.NewWriter(mockConn))
		return mockConn, rw, nil
	}

	w := &HijackableResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		HijackFunc:       hijackFunc,
	}

	_, _, _, err := ws.UpgradeHTTP(req, w)
	if err != nil {
		t.Fatalf("Failed to upgrade HTTP: %v", err)
	}

	var conn net.Conn = mockConn
	var wg sync.WaitGroup
	wg.Add(1)
	// Call the NewMachineChallenge function, simulating a request to the endpoint
	go NewMachineChallengeHandler(injector, req, w, &conn, &wg)
	wg.Wait()

	// Since NewMachineChallengeHandler is asynchronous, you might need to add some synchronization
	// mechanism to wait for its completion before asserting the outcomes, such as using a WaitGroup or
	// checking for certain conditions in the database or the mock objects.

	// After ensuring the asynchronous part has completed, you can inspect the HTTP response
	response := w.Result()
	if response.StatusCode != http.StatusOK { // Replace http.StatusExpected with the actual expected status code
		t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
	}

	// Additionally, perform any necessary assertions on the database or other side effects
}

func TestNewMachineChallengeHandler_GoldenPath(t *testing.T) {
	// Setup mocks for repositories and any other dependencies...
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	injector := do.New()
	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	// Setup dependency injection for repositories
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})
	do.Provide(injector, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	// Setup expectations for repository interactions
	mockUserRepo.EXPECT().GetUserByUsername(gomock.Any()).Return(&models.User{ID: uuid.New(), Username: "testuser"}, nil).AnyTimes()
	mockMachineRepo.EXPECT().GetMachineByNameAndUser(gomock.Any(), gomock.Any()).Return(nil, sql.ErrNoRows).AnyTimes()
	mockMachineRepo.EXPECT().CreateMachine(gomock.Any()).Return(&models.Machine{Name: "testmachine"}, nil).AnyTimes()

	router := chi.NewRouter()
	router.Get("/challenge", func(w http.ResponseWriter, r *http.Request) {
		err := MachineChallengeResponse(injector, r, w)
		if err != nil {
			log.Err(err).Msg("error with challenge response creation")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	router.Get("/challenge/existing", func(w http.ResponseWriter, r *http.Request) {
		err := NewMachineChallenge(injector, r, w)
		if err != nil {
			log.Err(err).Msg("error creating challenge")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	// Start a test HTTP server that upgrades connections to WebSocket
	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to the server using gobwas/ws as a client
	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), "ws://"+server.Listener.Addr().String()+"/challenge/existing")
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	err = utils.WriteClientMessage(&conn, dto.UserMachineDto{Username: "testuser", MachineName: "testmachine"})
	if err != nil {
		t.Fatalf("Failed to send user machine data: %v", err)
	}

	// Read and assert the server's response
	serverResponse, err := utils.ReadServerMessage[dto.MessageDto](&conn)
	if err != nil {
		t.Fatalf("Failed to read server response: %v", err)
	}
	// Assert the serverResponse contains the expected data
	correctChallengeResponse := serverResponse.Data.Message

	// Make a new websocket connnection and respond to the challenge

	conn_response, _, _, err := ws.DefaultDialer.Dial(context.Background(), "ws://"+server.Listener.Addr().String()+"/challenge")
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	// Simulate client sending the correct challenge response using your utils
	challengeResponse := dto.ChallengeResponseDto{Challenge: correctChallengeResponse}
	if err := utils.WriteClientMessage(&conn_response, challengeResponse); err != nil {
		t.Fatalf("Failed to send challenge response: %v", err)
	}

	// Read and assert the server's response
	serverResponse, err = utils.ReadServerMessage[dto.MessageDto](&conn_response)
	if err != nil {
		t.Fatalf("Failed to read server response: %v", err)
	}
	// Assert the serverResponse contains the expected data

	// This is becoming a pain in the neck. The full integration test will require basically rewriting the existing functionaliy of ssh-sync, where one user sends data to the other.
	// I wonder if instead we can write unit tests to really just test each individual piece of the 'NewMachineChallengeHandler' function.
}
