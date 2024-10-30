package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware/context_keys"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
)

func getData(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			log.Err(errors.New("could not get user from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		userRepo := do.MustInvoke[repository.UserRepository](i)
		keys, err := userRepo.GetUserKeys(user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		user.Keys = keys
		config, err := userRepo.GetUserConfig(user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		user.Config = config
		dto := dto.DataDto{
			ID:       user.ID,
			Username: user.Username,
			Keys: lo.Map(user.Keys, func(key models.SshKey, index int) dto.KeyDto {
				return dto.KeyDto{
					ID:       key.ID,
					UserID:   key.UserID,
					Filename: key.Filename,
					Data:     key.Data,
				}
			}),
			SshConfig: lo.Map(user.Config, func(conf models.SshConfig, index int) dto.SshConfigDto {
				return dto.SshConfigDto{
					Host:          conf.Host,
					Values:        conf.Values,
					IdentityFiles: conf.IdentityFiles,
				}
			}),
		}
		json.NewEncoder(w).Encode(dto)
	}
}

func addData(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			log.Err(errors.New("could not get user from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		machine, ok := r.Context().Value(context_keys.MachineContextKey).(*models.Machine)
		if !ok {
			log.Err(errors.New("could not get machine from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		userRepo := do.MustInvoke[repository.UserRepository](i)
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			log.Err(err).Msg("could not parse multipart form")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m := r.MultipartForm
		sshConfigDataRaw := r.FormValue("ssh_config")
		if sshConfigDataRaw == "" {
			log.Debug().Msg("ssh config is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var sshConfig []dto.SshConfigDto
		if err := json.NewDecoder(bytes.NewBufferString(sshConfigDataRaw)).Decode(&sshConfig); err != nil {
			log.Debug().Err(err).Msg("could not decode ssh config")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sshConfigData := lo.Map(sshConfig, func(conf dto.SshConfigDto, i int) models.SshConfig {
			return models.SshConfig{
				UserID:        user.ID,
				MachineID:     machine.ID,
				Host:          conf.Host,
				Values:        conf.Values,
				IdentityFiles: conf.IdentityFiles,
			}
		})
		user.Config = sshConfigData
		txQueryService := do.MustInvoke[query.TransactionService](i)
		tx, err := txQueryService.StartTx(pgx.TxOptions{})
		if err != nil {
			log.Err(err).Msg("error starting transaction")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer query.RollbackFunc(txQueryService, tx, w, &err)
		if err = userRepo.AddAndUpdateConfigTx(user, tx); err != nil {
			log.Err(err).Msg("could not add config")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var files []*multipart.FileHeader
		for _, filelist := range m.File {
			files = append(files, filelist...)
		}
		for i := range files {
			file, err := files[i].Open()
			if err != nil {
				log.Err(err).Msg("could not open file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer file.Close()
			user.Keys = append(user.Keys, models.SshKey{
				UserID:   user.ID,
				Filename: files[i].Filename,
				Data:     make([]byte, files[i].Size),
			})
			if _, err = file.Read(user.Keys[i].Data); err != nil {
				log.Err(err).Msg("could not open file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		if err = userRepo.AddAndUpdateKeysTx(user, tx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func deleteData(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			log.Err(errors.New("could not get user from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		keyIdStr := chi.URLParam(r, "id")
		keyId, err := uuid.Parse(keyIdStr)
		if err != nil {
			log.Err(err).Msg("could not parse key id")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		userRepo := do.MustInvoke[repository.UserRepository](i)
		key, err := userRepo.GetUserKey(user.ID, keyId)
		if err != nil {
			log.Err(err).Msg("could not get key")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		txQueryService := do.MustInvoke[query.TransactionService](i)
		tx, err := txQueryService.StartTx(pgx.TxOptions{})
		if err != nil {
			log.Err(err).Msg("error starting transaction")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer query.RollbackFunc(txQueryService, tx, w, &err)
		if err = userRepo.DeleteUserKeyTx(user, key.ID, tx); err != nil {
			log.Err(err).Msg("could not delete key")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func DataRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.ConfigureAuth(i))
	r.Get("/", getData(i))
	r.Post("/", addData(i))
	r.Delete("/key/{id}", deleteData(i))
	return r
}
