package routes

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/live"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
)

func SetupRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO need to introduce the concept of a transaction in case one of the DB operations fail
		// We don't want a user that has no machines - they would be unable to login
		var userDto dto.UserDto
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
		file, _, err := r.FormFile("key")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			log.Err(err).Msg("error reading file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		masterKeyStr := r.FormValue("master_key")
		if masterKeyStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// validate that it is in fact a public key
		key, err := jwk.ParseKey(fileBytes, jwk.WithPEM(true))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		keyType := key.KeyType()
		if keyType != jwa.EC {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user := models.User{}
		user.Username = userDto.Username
		if err := user.CreateUser(i); errors.Is(err, models.ErrUserAlreadyExists) {
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
		machine.PublicKey = fileBytes
		if err := machine.CreateMachine(i); errors.Is(err, models.ErrMachineAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err != nil {
			log.Err(err).Msg("error creating machine")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// next, create master key
		masterKey := models.MasterKey{}
		masterKey.UserID = user.ID
		masterKey.MachineID = machine.ID
		masterKey.Data = []byte(masterKeyStr)
		masterKey.CreateMasterKey(i)
		if errors.Is(err, models.ErrKeyAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err != nil {
			log.Err(err).Msg("error creating key")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	ch := chi.NewRouter()
	ch.Use(middleware.ConfigureAuth(i))
	ch.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := live.MachineChallengeResponse(i, r, w)
		if err != nil {
			log.Err(err).Msg("error with challenge response creation")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	r.Mount("/challenge", ch)
	r.Get("/existing", func(w http.ResponseWriter, r *http.Request) {
		err := live.NewMachineChallenge(i, r, w)
		if err != nil {
			log.Err(err).Msg("error creating challenge")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return r
}
