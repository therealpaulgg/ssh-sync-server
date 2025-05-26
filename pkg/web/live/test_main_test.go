package live

import (
	"os"
	"testing"
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	
	// Exit with the same code as the tests
	os.Exit(code)
}