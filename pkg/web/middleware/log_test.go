package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Forwarded-For", "192.0.2.1")

	rec := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {}

	writer := &bytes.Buffer{}
	logger := zerolog.New(writer)
	log.Logger = logger

	Log(http.HandlerFunc(handler)).ServeHTTP(rec, req)

	logLine := writer.String()
	assert.Contains(t, logLine, `"method":"GET"`)
	assert.Contains(t, logLine, `"path":"/test"`)
	assert.Contains(t, logLine, `"remote_addr":"192.0.2.1"`)
	assert.True(t, strings.Contains(logLine, `"duration"`))
}

func TestLogRealIp(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Real-Ip", "192.0.2.1")

	rec := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {}

	writer := &bytes.Buffer{}
	logger := zerolog.New(writer)
	log.Logger = logger

	Log(http.HandlerFunc(handler)).ServeHTTP(rec, req)

	logLine := writer.String()
	assert.Contains(t, logLine, `"method":"GET"`)
	assert.Contains(t, logLine, `"path":"/test"`)
	assert.Contains(t, logLine, `"remote_addr":"192.0.2.1"`)
	assert.True(t, strings.Contains(logLine, `"duration"`))
}

func TestLogRemoteAddr(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "192.0.2.1"

	rec := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {}

	writer := &bytes.Buffer{}
	logger := zerolog.New(writer)
	log.Logger = logger

	Log(http.HandlerFunc(handler)).ServeHTTP(rec, req)

	logLine := writer.String()
	assert.Contains(t, logLine, `"method":"GET"`)
	assert.Contains(t, logLine, `"path":"/test"`)
	assert.Contains(t, logLine, `"remote_addr":"192.0.2.1"`)
	assert.True(t, strings.Contains(logLine, `"duration"`))
}

func TestLogRemoteAddrPort(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "192.0.2.1:6969"

	rec := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {}

	writer := &bytes.Buffer{}
	logger := zerolog.New(writer)
	log.Logger = logger

	Log(http.HandlerFunc(handler)).ServeHTTP(rec, req)

	logLine := writer.String()
	assert.Contains(t, logLine, `"method":"GET"`)
	assert.Contains(t, logLine, `"path":"/test"`)
	assert.Contains(t, logLine, `"remote_addr":"192.0.2.1"`)
	assert.True(t, strings.Contains(logLine, `"duration"`))
}
