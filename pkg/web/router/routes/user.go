package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-common/pkg/dto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
)

func getUser(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("getUser: request received")
		username := chi.URLParam(r, "username")
		log.Debug().Str("username", username).Msg("getUser: username parameter retrieved")
		userRepo := do.MustInvoke[repository.UserRepository](i)
		user, err := userRepo.GetUserByUsername(username)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("username", user.Username).Msg("getUser: user retrieved")
		userDto := dto.UserDto{
			Username: user.Username,
		}
		json.NewEncoder(w).Encode(userDto)
		log.Debug().Msg("getUser: response sent")
	}
}

func UserRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Get("/{username}", getUser(i))
	return r
}
