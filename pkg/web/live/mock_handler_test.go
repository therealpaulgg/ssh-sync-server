package live

import (
	"net"
	"net/http"
	"testing"

	"github.com/samber/do"
	"github.com/undefinedlabs/go-mpatch"
)

// MockMachineChallengeResponseHandler is a mock implementation of MachineChallengeResponseHandler
func MockMachineChallengeResponseHandler(t *testing.T, mockConn *MockConn) (*mpatch.Patch, error) {
	return mpatch.PatchMethod(MachineChallengeResponse, func(i *do.Injector, r *http.Request, w http.ResponseWriter) error {
		// Always successfully upgrade and return nil
		go func() {
			conn := net.Conn(mockConn)
			MachineChallengeResponseHandler(i, r, w, &conn)
		}()
		return nil
	})
}

// MockNewMachineChallengeHandler is a mock implementation of NewMachineChallengeHandler
func MockNewMachineChallengeHandler(t *testing.T, mockConn *MockConn) (*mpatch.Patch, error) {
	return mpatch.PatchMethod(NewMachineChallenge, func(i *do.Injector, r *http.Request, w http.ResponseWriter) error {
		// Always successfully upgrade and return nil
		go func() {
			conn := net.Conn(mockConn)
			NewMachineChallengeHandler(i, r, w, &conn)
		}()
		return nil
	})
}