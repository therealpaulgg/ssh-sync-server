package routes

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/live"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
)

func initialSetup(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		txQueryService := do.MustInvoke[query.TransactionService](i)
		tx, err := txQueryService.StartTx(pgx.TxOptions{})
		if err != nil {
			log.Err(err).Msg("error starting transaction")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer query.RollbackFunc(txQueryService, tx, w)
		userRepo := do.MustInvoke[repository.UserRepository](i)
		user := &models.User{}
		user.Username = userDto.Username
		user, err = userRepo.CreateUserTx(user, tx)
		if err != nil {
			if errors.Is(err, repository.ErrUserAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			log.Err(err).Msg("error creating user")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		machine := &models.Machine{}
		machine.Name = machineName
		machine.UserID = user.ID
		machine.PublicKey = fileBytes
		_, err = machineRepo.CreateMachineTx(machine, tx)
		if err != nil {
			if errors.Is(err, repository.ErrMachineAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			log.Err(err).Msg("error creating machine")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func challengeResponse(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := live.MachineChallengeResponse(i, r, w)
		if err != nil {
			log.Err(err).Msg("error with challenge response creation")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func getExisting(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := live.NewMachineChallenge(i, r, w)
		if err != nil {
			log.Err(err).Msg("error creating challenge")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func SetupRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Post("/", initialSetup(i))
	ch := chi.NewRouter()
	ch.Use(middleware.ConfigureAuth(i))
	ch.Get("/", challengeResponse(i))
	r.Mount("/challenge", ch)
	r.Get("/existing", getExisting(i))
	return r
}
