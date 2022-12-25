package routes

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
)

func SetupRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var userDto UserDto
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		username := r.FormValue("username")
		if username == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		userDto.Username = username
		machineName := r.FormValue("machine_name")
		if machineName == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("key")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fileBytes, err := ioutil.ReadAll(file)
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
		machine := models.Machine{}
		machine.Name = machineName
		machine.UserID = user.ID
		err = machine.CreateMachine(i)
		if errors.Is(err, models.ErrMachineAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err != nil {
			log.Err(err).Msg("error creating machine")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		key := models.SshKey{
			UserID:   user.ID,
			Filename: header.Filename,
			Data:     fileBytes,
		}
	})
	return r
}
