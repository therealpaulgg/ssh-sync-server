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

func AddUserContext(req *http.Request, user *models.User) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserContextKey, user)
	return req.Clone(ctx)
}
