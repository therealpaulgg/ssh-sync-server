package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/therealpaulgg/ssh-sync-common/pkg/dto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/crypto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware/context_keys"
)

type DeleteRequest struct {
	MachineName string `json:"machine_name"`
}

func getMachineById(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("username", user.Username).Msg("getMachineById: request received")
		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		machines, err := machineRepo.GetUserMachines(user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Int("machines_count", len(machines)).Msg("getMachineById: fetched machines for user")
		user.Machines = machines
		machineId, err := uuid.Parse(chi.URLParam(r, "machineId"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Str("machine_id", machineId.String()).Msg("getMachineById: parsed machine id")
		machine, found := lo.Find(user.Machines, func(machine models.Machine) bool {
			return machine.ID == machineId
		})
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Debug().Str("machine_name", machine.Name).Msg("getMachineById: found machine")
		machineDto := dto.MachineDto{
			Name: machine.Name,
		}
		json.NewEncoder(w).Encode(machineDto)
	}
}

func getMachines(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("username", user.Username).Msg("getMachines: request received")
		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		machines, err := machineRepo.GetUserMachines(user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Int("machines_count", len(machines)).Msg("getMachines: fetched machines for user")
		user.Machines = machines
		machineDtos := make([]dto.MachineDto, len(user.Machines))
		for i, machine := range user.Machines {
			machineDtos[i] = dto.MachineDto{
				Name: machine.Name,
			}
		}
		json.NewEncoder(w).Encode(machineDtos)
	}
}

func deleteMachine(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("username", user.Username).Msg("deleteMachine: request received")
		var deleteRequest DeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&deleteRequest); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Str("machine_name", deleteRequest.MachineName).Msg("deleteMachine: parsed request body")
		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		machine, err := machineRepo.GetMachineByNameAndUser(deleteRequest.MachineName, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			log.Err(err).Msg("Error getting machine by name and user")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("machine_id", machine.ID.String()).Msg("deleteMachine: fetched machine")
		if err := machineRepo.DeleteMachine(machine.ID); err != nil {
			log.Err(err).Msg("Error deleting machine")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("machine_id", machine.ID.String()).Msg("deleteMachine: machine deleted")
	}
}

func updateMachineKey(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		machine, ok := r.Context().Value(context_keys.MachineContextKey).(*models.Machine)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("machine_name", machine.Name).Msg("updateMachineKey: request received")
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Msg("updateMachineKey: parsed multipart form")
		file, _, err := r.FormFile("key")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			log.Err(err).Msg("error reading key file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := crypto.ValidatePublicKey(fileBytes); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Msg("updateMachineKey: public key validated")
		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		if err := machineRepo.UpdateMachinePublicKey(machine.ID, fileBytes); err != nil {
			log.Err(err).Msg("error updating machine public key")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("machine_id", machine.ID.String()).Msg("updateMachineKey: machine public key updated")
		ekFile, _, ekErr := r.FormFile("encapsulation_key")
		if ekErr == nil {
			defer ekFile.Close()
			ekBytes, err := io.ReadAll(ekFile)
			if err != nil {
				log.Err(err).Msg("error reading encapsulation key file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := machineRepo.UpdateMachineEncapsulationKey(machine.ID, ekBytes); err != nil {
				log.Err(err).Msg("error updating machine encapsulation key")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			log.Debug().Str("machine_id", machine.ID.String()).Msg("updateMachineKey: encapsulation key updated")
		}
		w.WriteHeader(http.StatusOK)
	}
}

func getMachinePublicKeys(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		machines, err := machineRepo.GetUserMachines(user.ID)
		if err != nil {
			log.Err(err).Msg("getMachinePublicKeys: error fetching machines")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		dtos := make([]dto.MachinePublicKeyDto, len(machines))
		for idx, m := range machines {
			dtos[idx] = dto.MachinePublicKeyDto{
				MachineID:        m.ID,
				Name:             m.Name,
				PublicKey:        m.PublicKey,
				EncapsulationKey: m.EncapsulationKey,
			}
		}
		json.NewEncoder(w).Encode(dto.MachinesPublicKeysDto{Machines: dtos})
	}
}

func MachineRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.ConfigureAuth(i))
	r.Get("/{machineId}", getMachineById(i))
	r.Get("/", getMachines(i))
	r.Get("/public-keys", getMachinePublicKeys(i))
	r.Delete("/", deleteMachine(i))
	r.Put("/key", updateMachineKey(i))
	return r
}
