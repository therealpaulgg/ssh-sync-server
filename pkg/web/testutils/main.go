package testutils

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
)

func GenerateUser() *models.User {
	return &models.User{
		ID:       uuid.New(),
		Username: "test",
	}
}

func GenerateMachine() *models.Machine {
	return &models.Machine{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Name:      "test",
		PublicKey: []byte("test"),
	}
}

func AddUserContext(req *http.Request, user *models.User) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserContextKey, user)
	return req.Clone(ctx)
}

func AddMachineContext(req *http.Request, machine *models.Machine) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.MachineContextKey, machine)
	return req.Clone(ctx)
}
