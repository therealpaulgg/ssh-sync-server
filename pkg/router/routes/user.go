package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
)

type UserDto struct {
	Username string `json:"username"`
}

func UserRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Get("/{username}", func(w http.ResponseWriter, r *http.Request) {
		user := models.User{}
		user.Username = chi.URLParam(r, "username")
		err := user.GetUserByUsername(i)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, user)
	})
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var userDto UserDto
		err := json.NewDecoder(r.Body).Decode(&userDto)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user := models.User{}
		user.Username = userDto.Username
		err = user.CreateUser(i)
		if errors.Is(err, models.ErrUserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err != nil {
			log.Err(err).Msg("error creating user")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, user)
	})
	return r
}
