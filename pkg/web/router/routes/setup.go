package routes

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-common/pkg/dto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/crypto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/live"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
)

func initialSetup(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("initialSetup: request received")
		var userDto dto.UserDto
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Msg("initialSetup: parsed multipart form")
		username := r.FormValue("username")
		if username == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Str("username", username).Msg("initialSetup: parsed username")
		userDto.Username = username
		machineName := r.FormValue("machine_name")
		if machineName == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Str("machine_name", machineName).Msg("initialSetup: parsed machine name")
		file, _, err := r.FormFile("key")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := crypto.ValidatePublicKey(fileBytes); err != nil {
			log.Debug().Err(err).Msg("invalid public key")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Msg("initialSetup: public key validated")
		var encapsulationKeyBytes []byte
		ekFile, _, ekErr := r.FormFile("encapsulation_key")
		if ekErr == nil {
			defer ekFile.Close()
			encapsulationKeyBytes, err = io.ReadAll(ekFile)
			if err != nil {
				log.Err(err).Msg("error reading encapsulation key file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		txQueryService := do.MustInvoke[query.TransactionService](i)
		tx, err := txQueryService.StartTx(pgx.TxOptions{})
		if err != nil {
			log.Err(err).Msg("error starting transaction")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Msg("initialSetup: transaction started")
		defer query.RollbackFunc(txQueryService, tx, w, &err)
		userRepo := do.MustInvoke[repository.UserRepository](i)
		user := &models.User{}
		user.Username = userDto.Username
		log.Debug().Str("username", user.Username).Msg("initialSetup: creating user")
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
		log.Debug().Str("user_id", user.ID.String()).Msg("initialSetup: user created")
		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		machine := &models.Machine{}
		machine.Name = machineName
		machine.UserID = user.ID
		machine.PublicKey = fileBytes
		machine.EncapsulationKey = encapsulationKeyBytes
		log.Debug().Str("machine_name", machine.Name).Msg("initialSetup: creating machine")
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
		log.Debug().Str("machine_name", machine.Name).Msg("initialSetup: machine created")
	}
}

func challengeResponse(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("challengeResponse: request received")
		err := live.MachineChallengeResponse(i, r, w)
		if err != nil {
			log.Err(err).Msg("error with challenge response creation")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Msg("challengeResponse: response created")
	}
}

func getExisting(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("getExisting: request received")
		err := live.NewMachineChallenge(i, r, w)
		if err != nil {
			log.Err(err).Msg("error creating challenge")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Msg("getExisting: challenge created")
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
