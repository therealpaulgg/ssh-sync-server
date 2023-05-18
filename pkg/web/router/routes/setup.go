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

func SetupRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
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
		txQueryService := do.MustInvoke[query.QueryServiceTx[models.User]](i)
		tx, err := txQueryService.StartTx(pgx.TxOptions{})
		if err != nil {
			log.Err(err).Msg("error starting transaction")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			rb := func(tx pgx.Tx) {
				err := txQueryService.Rollback(tx)
				if err != nil {
					log.Err(err).Msg("error rolling back transaction")
				}
			}
			if err != nil {
				rb(tx)
			} else {
				internalErr := txQueryService.Commit(tx)
				if internalErr != nil {
					log.Err(err).Msg("error committing transaction")
					rb(tx)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()
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
		machine, err = machineRepo.CreateMachineTx(machine, tx)
		if err != nil {
			if errors.Is(err, repository.ErrMachineAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			log.Err(err).Msg("error creating machine")
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
