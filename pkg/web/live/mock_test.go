package live

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository/mock"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware/context_keys"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
)

// MockConn is a mock implementation of net.Conn for testing
type MockConn struct {
	ReadData  chan []byte
	WriteData chan []byte
	Closed    bool
	ReadErr   error
	WriteErr  error
}

func NewMockConn() *MockConn {
	return &MockConn{
		ReadData:  make(chan []byte, 10), // Buffered to avoid blocking in tests
		WriteData: make(chan []byte, 10), // Buffered to avoid blocking in tests
		Closed:    false,
		ReadErr:   nil,
		WriteErr:  nil,
	}
}

// net.Conn interface implementation for MockConn
func (m *MockConn) Read(b []byte) (n int, err error) {
	if m.Closed {
		return 0, io.EOF
	}
	if m.ReadErr != nil {
		return 0, m.ReadErr
	}
	
	data := <-m.ReadData
	copy(b, data)
	return len(data), nil
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	if m.Closed {
		return 0, errors.New("connection closed")
	}
	if m.WriteErr != nil {
		return 0, m.WriteErr
	}
	
	// Copy data to avoid races
	dataCopy := make([]byte, len(b))
	copy(dataCopy, b)
	m.WriteData <- dataCopy
	return len(b), nil
}

func (m *MockConn) Close() error {
	if !m.Closed {
		m.Closed = true
		close(m.ReadData)
		close(m.WriteData)
	}
	return nil
}

// Implementing other required methods for net.Conn interface
func (m *MockConn) LocalAddr() net.Addr                { return &net.TCPAddr{IP: net.IPv4zero, Port: 0} }
func (m *MockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{IP: net.IPv4zero, Port: 0} }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

// MockResponseWriter is a mock implementation of http.ResponseWriter for testing
type MockResponseWriter struct {
	Headers    http.Header
	StatusCode int
	Body       []byte
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		Headers:    make(http.Header),
		StatusCode: 0,
		Body:       []byte{},
	}
}

func (m *MockResponseWriter) Header() http.Header {
	return m.Headers
}

func (m *MockResponseWriter) Write(b []byte) (int, error) {
	m.Body = append(m.Body, b...)
	return len(b), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.StatusCode = statusCode
}

// Helper function to setup mock dependencies for testing
func setupMockDependencies(ctrl *gomock.Controller) (*do.Injector, *models.User) {
	i := do.New()
	
	// Create mock repositories
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockMachineRepo := mock.NewMockMachineRepository(ctrl)
	
	// Register mocks in injector
	do.Provide(i, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})
	do.Provide(i, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})
	
	// Create a test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
	}
	
	// Setup mock repository behaviors
	mockUserRepo.EXPECT().
		GetUserByUsername(gomock.Eq(user.Username)).
		Return(user, nil).
		AnyTimes()
	
	mockUserRepo.EXPECT().
		GetUserByUsername(gomock.Not(user.Username)).
		Return(nil, errors.New("user not found")).
		AnyTimes()
	
	// Mock machine repository behaviors
	mockMachineRepo.EXPECT().
		GetMachineByNameAndUser(gomock.Eq("existing-machine"), gomock.Any()).
		Return(&models.Machine{
			ID:     uuid.New(),
			Name:   "existing-machine",
			UserID: user.ID,
		}, nil).
		AnyTimes()
	
	mockMachineRepo.EXPECT().
		GetMachineByNameAndUser(gomock.Not("existing-machine"), gomock.Any()).
		Return(nil, errors.New("sql: no rows in result set")).
		AnyTimes()
	
	mockMachineRepo.EXPECT().
		CreateMachine(gomock.Any()).
		DoAndReturn(func(machine *models.Machine) (*models.Machine, error) {
			machine.ID = uuid.New()
			return machine, nil
		}).
		AnyTimes()
	
	return i, user
}

// Helper to create a mock request with user context
func createMockRequestWithUser(user *models.User) *http.Request {
	req := &http.Request{
		Header: make(http.Header),
	}
	ctx := context.WithValue(context.Background(), context_keys.UserContextKey, user)
	return req.WithContext(ctx)
}

// Helper to write a dto message to a mock connection
func writeMessageToConn(conn *MockConn, message interface{}) error {
	data, err := json.Marshal(struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: "message",
		Data: message,
	})
	if err != nil {
		return err
	}
	
	conn.ReadData <- data
	return nil
}

// Helper to read WebSocket response from conn
func readResponseFromConn(conn *MockConn) (map[string]interface{}, error) {
	select {
	case data := <-conn.WriteData:
		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, err
		}
		return response, nil
	case <-time.After(time.Second):
		return nil, errors.New("timeout waiting for response")
	}
}

// Create a mocker that returns our mock connection
var mockUpgrade = func(conn net.Conn) func(*http.Request, http.ResponseWriter) (net.Conn, []byte, ws.OpCode, error) {
	return func(r *http.Request, w http.ResponseWriter) (net.Conn, []byte, ws.OpCode, error) {
		return conn, nil, ws.OpText, nil
	}
}

// Helper to create various DTOs for testing
func createChallengeResponseDto(challenge string) dto.ChallengeResponseDto {
	return dto.ChallengeResponseDto{
		Challenge: challenge,
	}
}

func createUserMachineDto(username, machineName string) dto.UserMachineDto {
	return dto.UserMachineDto{
		Username:    username,
		MachineName: machineName,
	}
}

func createPublicKeyDto(publicKey []byte) dto.PublicKeyDto {
	return dto.PublicKeyDto{
		PublicKey: publicKey,
	}
}

func createEncryptedMasterKeyDto(encryptedKey []byte) dto.EncryptedMasterKeyDto {
	return dto.EncryptedMasterKeyDto{
		EncryptedMasterKey: encryptedKey,
	}
}