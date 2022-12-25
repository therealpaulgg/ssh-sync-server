package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/middleware"
)

type DataDto struct {
	ID        uuid.UUID      `json:"id"`
	Username  string         `json:"username"`
	Keys      []KeyDto       `json:"keys"`
	MasterKey []byte         `json:"master_key"`
	SshConfig []SshConfigDto `json:"ssh_config"`
}

type KeyDto struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	Filename string    `json:"filename"`
	Data     []byte    `json:"data"`
}

type SshConfigDto struct {
	Host   string            `json:"host"`
	Values map[string]string `json:"values"`
}

func DataRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.ConfigureAuth(i))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
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
		err := user.GetUserKeys(i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		masterKey := models.MasterKey{}
		masterKey.UserID = user.ID
		masterKey.MachineID = machine.ID
		err = masterKey.GetMasterKey(i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		dto := DataDto{
			ID:       user.ID,
			Username: user.Username,
			Keys: lo.Map(user.Keys, func(key models.SshKey, index int) KeyDto {
				return KeyDto{
					ID:       key.ID,
					UserID:   key.UserID,
					Filename: key.Filename,
					Data:     key.Data,
				}
			}),
			MasterKey: masterKey.Data,
			SshConfig: lo.Map(user.Config, func(conf models.SshConfig, index int) SshConfigDto {
				return SshConfigDto{
					Host:   conf.Host,
					Values: conf.Values,
				}
			}),
		}
		json.NewEncoder(w).Encode(dto)
	})
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m := r.MultipartForm
		sshConfigDataRaw := r.FormValue("ssh_config")
		if sshConfigDataRaw == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var sshConfig []SshConfigDto
		err = json.NewDecoder(bytes.NewBufferString(sshConfigDataRaw)).Decode(&sshConfig)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sshConfigData := lo.Map(sshConfig, func(conf SshConfigDto, i int) models.SshConfig {
			return models.SshConfig{
				UserID:    user.ID,
				MachineID: machine.ID,
				Host:      conf.Host,
				Values:    conf.Values,
			}
		})
		user.Config = sshConfigData
		err = user.AddAndUpdateConfig(i)
		if err != nil {
			log.Err(err).Msg("could not add config")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var files []*multipart.FileHeader
		for _, filelist := range m.File {
			files = append(files, filelist...)
		}
		if len(files) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
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
			_, err = file.Read(user.Keys[i].Data)
			if err != nil {
				log.Err(err).Msg("could not open file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		err = user.AddAndUpdateConfig(i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = user.AddAndUpdateKeys(i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return r
}
