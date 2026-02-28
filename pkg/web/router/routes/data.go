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
	"github.com/therealpaulgg/ssh-sync-common/pkg/dto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware/context_keys"
)

func getData(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			log.Err(errors.New("could not get user from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("username", user.Username).Msg("getData: request received")
		userRepo := do.MustInvoke[repository.UserRepository](i)
		keys, err := userRepo.GetUserKeys(user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Int("keys_count", len(keys)).Msg("getData: fetched user keys")
		user.Keys = keys
		config, err := userRepo.GetUserConfig(user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Int("config_count", len(config)).Msg("getData: fetched user config")
		user.Config = config
		knownHosts, err := userRepo.GetUserKnownHosts(user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		user.KnownHosts = knownHosts
		dto := dto.DataDto{
			ID:       user.ID,
			Username: user.Username,
			Keys: lo.Map(user.Keys, func(key models.SshKey, index int) dto.KeyDto {
				return dto.KeyDto{
					ID:        key.ID,
					UserID:    key.UserID,
					Filename:  key.Filename,
					Data:      key.Data,
					UpdatedAt: key.UpdatedAt,
				}
			}),
			SshConfig: lo.Map(user.Config, func(conf models.SshConfig, index int) dto.SshConfigDto {
				return dto.SshConfigDto{
					ID:            conf.ID,
					Host:          conf.Host,
					Values:        conf.Values,
					IdentityFiles: conf.IdentityFiles,
				}
			}),
			KnownHosts: lo.Map(user.KnownHosts, func(kh models.KnownHost, index int) dto.KnownHostDto {
				return dto.KnownHostDto{
					HostPattern: kh.HostPattern,
					KeyType:     kh.KeyType,
					KeyData:     kh.KeyData,
					Marker:      kh.Marker,
				}
			}),
		}
		log.Debug().Msg("getData: responding with user data")
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
		log.Debug().Str("username", user.Username).Msg("addData: request received")
		userRepo := do.MustInvoke[repository.UserRepository](i)
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			log.Err(err).Msg("could not parse multipart form")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Msg("addData: parsed multipart form")
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
		log.Debug().Int("ssh_config_count", len(sshConfig)).Msg("addData: decoded ssh config")
		sshConfigData := lo.Map(sshConfig, func(conf dto.SshConfigDto, i int) models.SshConfig {
			return models.SshConfig{
				UserID:        user.ID,
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
		log.Debug().Msg("addData: transaction started")
		defer query.RollbackFunc(txQueryService, tx, w, &err)
		if err = userRepo.AddAndUpdateConfigTx(user, tx); err != nil {
			log.Err(err).Msg("could not add config")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Int("ssh_config_count", len(user.Config)).Msg("addData: stored ssh config")
		knownHostsRaw := r.FormValue("known_hosts")
		if knownHostsRaw != "" {
			var knownHostDtos []dto.KnownHostDto
			if err = json.NewDecoder(bytes.NewBufferString(knownHostsRaw)).Decode(&knownHostDtos); err != nil {
				log.Debug().Err(err).Msg("could not decode known_hosts")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			user.KnownHosts = lo.Map(knownHostDtos, func(kh dto.KnownHostDto, _ int) models.KnownHost {
				return models.KnownHost{
					UserID:      user.ID,
					HostPattern: kh.HostPattern,
					KeyType:     kh.KeyType,
					KeyData:     kh.KeyData,
					Marker:      kh.Marker,
				}
			})
			if err = userRepo.AddAndUpdateKnownHostsTx(user, tx); err != nil {
				log.Err(err).Msg("could not add known_hosts")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		log.Debug().Int("known_hosts_count", len(user.KnownHosts)).Msg("addData: stored known hosts")
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
			log.Debug().Str("filename", files[i].Filename).Msg("addData: processing key file")
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
			log.Debug().Str("filename", files[i].Filename).Msg("addData: read key file")
		}
		if err = userRepo.AddAndUpdateKeysTx(user, tx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Int("keys_count", len(user.Keys)).Msg("addData: stored keys")
		responseKeys := lo.Map(user.Keys, func(key models.SshKey, _ int) dto.KeyDto {
			return dto.KeyDto{
				Filename:  key.Filename,
				UpdatedAt: key.UpdatedAt,
			}
		})
		json.NewEncoder(w).Encode(responseKeys)
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
		log.Debug().Str("username", user.Username).Msg("deleteData: request received")
		keyIdStr := chi.URLParam(r, "id")
		keyId, err := uuid.Parse(keyIdStr)
		if err != nil {
			log.Err(err).Msg("could not parse key id")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Str("key_id", keyId.String()).Msg("deleteData: parsed key id")
		userRepo := do.MustInvoke[repository.UserRepository](i)
		key, err := userRepo.GetUserKey(user.ID, keyId)
		if err != nil {
			log.Err(err).Msg("could not get key")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Debug().Str("key_filename", key.Filename).Msg("deleteData: fetched key")
		txQueryService := do.MustInvoke[query.TransactionService](i)
		tx, err := txQueryService.StartTx(pgx.TxOptions{})
		if err != nil {
			log.Err(err).Msg("error starting transaction")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Msg("deleteData: transaction started")
		defer query.RollbackFunc(txQueryService, tx, w, &err)
		if err = userRepo.DeleteUserKeyTx(user, key.ID, tx); err != nil {
			log.Err(err).Msg("could not delete key")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("key_id", key.ID.String()).Msg("deleteData: key deleted")
	}
}

func upsertConfigEntry(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			log.Err(errors.New("could not get user from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("username", user.Username).Msg("upsertConfigEntry: request received")
		var configDto dto.SshConfigDto
		if err := json.NewDecoder(r.Body).Decode(&configDto); err != nil {
			log.Err(err).Msg("upsertConfigEntry: could not decode body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if configDto.Host == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		config := models.SshConfig{
			UserID:        user.ID,
			Host:          configDto.Host,
			Values:        configDto.Values,
			IdentityFiles: configDto.IdentityFiles,
		}
		configRepo := do.MustInvoke[repository.SshConfigRepository](i)
		txQueryService := do.MustInvoke[query.TransactionService](i)
		tx, err := txQueryService.StartTx(pgx.TxOptions{})
		if err != nil {
			log.Err(err).Msg("upsertConfigEntry: error starting transaction")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer query.RollbackFunc(txQueryService, tx, w, &err)
		if _, err = configRepo.UpsertSshConfigTx(&config, tx); err != nil {
			log.Err(err).Msg("upsertConfigEntry: could not upsert config")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("host", config.Host).Msg("upsertConfigEntry: config entry upserted")
	}
}

func deleteConfigEntry(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			log.Err(errors.New("could not get user from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("username", user.Username).Msg("deleteConfigEntry: request received")
		configIdStr := chi.URLParam(r, "id")
		configId, err := uuid.Parse(configIdStr)
		if err != nil {
			log.Err(err).Msg("deleteConfigEntry: could not parse config id")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Str("config_id", configId.String()).Msg("deleteConfigEntry: parsed config id")
		userRepo := do.MustInvoke[repository.UserRepository](i)
		txQueryService := do.MustInvoke[query.TransactionService](i)
		tx, err := txQueryService.StartTx(pgx.TxOptions{})
		if err != nil {
			log.Err(err).Msg("deleteConfigEntry: error starting transaction")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer query.RollbackFunc(txQueryService, tx, w, &err)
		if err = userRepo.DeleteUserConfigTx(user, configId, tx); err != nil {
			log.Err(err).Msg("deleteConfigEntry: could not delete config entry")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("config_id", configId.String()).Msg("deleteConfigEntry: config entry deleted")
	}
}

func DataRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.ConfigureAuth(i))
	r.Get("/", getData(i))
	r.Post("/", addData(i))
	r.Delete("/key/{id}", deleteData(i))
	r.Post("/config", upsertConfigEntry(i))
	r.Delete("/config/{id}", deleteConfigEntry(i))
	return r
}
