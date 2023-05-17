package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
)

func getData(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok {
			log.Err(errors.New("could not get user from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err := user.GetUserKeys(i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := user.GetUserConfig(i); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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

func DataRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.ConfigureAuth(i))
	r.Get("/", getData(i))
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok {
			log.Err(errors.New("could not get user from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		machine, ok := r.Context().Value(middleware.MachineContextKey).(*models.Machine)
		if !ok {
			log.Err(errors.New("could not get machine from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
		if err = user.AddAndUpdateConfigTx(i, tx); err != nil {
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
		if err = user.AddAndUpdateKeysTx(i, tx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return r
}
