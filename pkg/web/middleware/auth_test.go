package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/testutils"
)

func GenerateTestToken(username, machine string, key jwk.Key) (string, error) {
	builder := jwt.NewBuilder()
	builder.Issuer("github.com/therealpaulgg/ssh-sync")
	builder.IssuedAt(time.Now())
	builder.Expiration(time.Now().Add(time.Minute))
	builder.Claim("username", username)
	builder.Claim("machine", machine)
	tok, err := builder.Build()
	if err != nil {
		return "", err
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.ES512, key))
	if err != nil {
		return "", err
	}
	return string(signed), nil
}

func TestConfigureAuth(t *testing.T) {
	// Arrange
	i := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	priv, pub, err := testutils.GenerateTestKeys()
	if err != nil {
		t.Fatal(err)
	}
	pubBytes, privBytes, err := testutils.EncodeToPem(priv, pub)
	if err != nil {
		t.Fatal(err)
	}
	key, err := jwk.ParseKey(privBytes, jwk.WithPEM(true))
	if err != nil {
		t.Fatal(err)
	}
	// Mock user and machine data
	user := &models.User{ID: uuid.New(), Username: "testuser"}
	machine := &models.Machine{ID: uuid.New(), Name: "testmachine", UserID: user.ID, PublicKey: pubBytes}
	token, err := GenerateTestToken(user.Username, machine.Name, key)
	if err != nil {
		t.Fatal(err)
	}

	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserByUsername(user.Username).Return(user, nil).Times(1)
	do.Provide(i, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetMachineByNameAndUser(machine.Name, user.ID).Return(machine, nil).Times(1)
	do.Provide(i, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	// Mock http request with Authorization header
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	rr := httptest.NewRecorder()
	f := ConfigureAuth(i)
	f(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dummy handler
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestConfigureAuthNoUser(t *testing.T) {
	// Arrange
	i := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	priv, pub, err := testutils.GenerateTestKeys()
	if err != nil {
		t.Fatal(err)
	}
	pubBytes, privBytes, err := testutils.EncodeToPem(priv, pub)
	if err != nil {
		t.Fatal(err)
	}
	key, err := jwk.ParseKey(privBytes, jwk.WithPEM(true))
	if err != nil {
		t.Fatal(err)
	}
	// Mock user and machine data
	user := &models.User{ID: uuid.New(), Username: "testuser"}
	machine := &models.Machine{ID: uuid.New(), Name: "testmachine", UserID: user.ID, PublicKey: pubBytes}
	token, err := GenerateTestToken(user.Username, machine.Name, key)
	if err != nil {
		t.Fatal(err)
	}

	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserByUsername(user.Username).Return(nil, sql.ErrNoRows).Times(1)
	do.Provide(i, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	// Mock http request with Authorization header
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	rr := httptest.NewRecorder()
	f := ConfigureAuth(i)
	f(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dummy handler
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestConfigureAuthNoMachine(t *testing.T) {
	// Arrange
	i := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	priv, pub, err := testutils.GenerateTestKeys()
	if err != nil {
		t.Fatal(err)
	}
	pubBytes, privBytes, err := testutils.EncodeToPem(priv, pub)
	if err != nil {
		t.Fatal(err)
	}
	key, err := jwk.ParseKey(privBytes, jwk.WithPEM(true))
	if err != nil {
		t.Fatal(err)
	}
	// Mock user and machine data
	user := &models.User{ID: uuid.New(), Username: "testuser"}
	machine := &models.Machine{ID: uuid.New(), Name: "testmachine", UserID: user.ID, PublicKey: pubBytes}
	token, err := GenerateTestToken(user.Username, machine.Name, key)
	if err != nil {
		t.Fatal(err)
	}

	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserByUsername(user.Username).Return(user, nil).Times(1)
	do.Provide(i, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetMachineByNameAndUser(machine.Name, user.ID).Return(nil, sql.ErrNoRows).Times(1)
	do.Provide(i, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	// Mock http request with Authorization header
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	rr := httptest.NewRecorder()
	f := ConfigureAuth(i)
	f(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dummy handler
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestConfigureAuthUnsignedToken(t *testing.T) {
	// Arrange
	i := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	priv, pub, err := testutils.GenerateTestKeys()
	if err != nil {
		t.Fatal(err)
	}
	pubBytes, _, err := testutils.EncodeToPem(priv, pub)
	if err != nil {
		t.Fatal(err)
	}
	// Mock user and machine data
	user := &models.User{ID: uuid.New(), Username: "testuser"}
	machine := &models.Machine{ID: uuid.New(), Name: "testmachine", UserID: user.ID, PublicKey: pubBytes}
	// Create HMAC signed token
	builder := jwt.NewBuilder()
	builder.Issuer("github.com/therealpaulgg/ssh-sync")
	builder.IssuedAt(time.Now())
	builder.Expiration(time.Now().Add(time.Minute))
	builder.Claim("username", user.Username)
	builder.Claim("machine", machine.Name)
	tok, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	token, err := jwt.Sign(tok, jwt.WithInsecureNoSignature())
	if err != nil {
		t.Fatal(err)
	}

	mockUserRepo := repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetUserByUsername(user.Username).Return(user, nil).Times(1)
	do.Provide(i, func(i *do.Injector) (repository.UserRepository, error) {
		return mockUserRepo, nil
	})

	mockMachineRepo := repository.NewMockMachineRepository(ctrl)
	mockMachineRepo.EXPECT().GetMachineByNameAndUser(machine.Name, user.ID).Return(machine, nil).Times(1)
	do.Provide(i, func(i *do.Injector) (repository.MachineRepository, error) {
		return mockMachineRepo, nil
	})

	// Mock http request with Authorization header
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+string(token))

	// Act
	rr := httptest.NewRecorder()
	f := ConfigureAuth(i)
	f(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dummy handler
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestConfigureAuthNoAuthHeader(t *testing.T) {
	// Arrange
	i := do.New()

	// Mock http request without Authorization header
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Act
	rr := httptest.NewRecorder()
	f := ConfigureAuth(i)
	f(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dummy handler
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestConfigureAuthBearerKeywordOnly(t *testing.T) {
	// Arrange
	i := do.New()

	// Mock http request without Authorization header
	req, err := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer")
	if err != nil {
		t.Fatal(err)
	}

	// Act
	rr := httptest.NewRecorder()
	f := ConfigureAuth(i)
	f(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dummy handler
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestConfigureAuthBearerKeywordWithSpace(t *testing.T) {
	// Arrange
	i := do.New()

	// Mock http request without Authorization header
	req, err := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	if err != nil {
		t.Fatal(err)
	}

	// Act
	rr := httptest.NewRecorder()
	f := ConfigureAuth(i)
	f(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dummy handler
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestConfigureAuthFakeToken(t *testing.T) {
	// Arrange
	i := do.New()

	// Mock http request without Authorization header
	req, err := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer thisisnotreal!!!")
	if err != nil {
		t.Fatal(err)
	}

	// Act
	rr := httptest.NewRecorder()
	f := ConfigureAuth(i)
	f(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dummy handler
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
