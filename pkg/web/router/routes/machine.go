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
	pqc "github.com/therealpaulgg/ssh-sync-server/pkg/crypto"
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
		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		machines, err := machineRepo.GetUserMachines(user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		user.Machines = machines
		machineId, err := uuid.Parse(chi.URLParam(r, "machineId"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		machine, found := lo.Find(user.Machines, func(machine models.Machine) bool {
			return machine.ID == machineId
		})
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}
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
		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		machines, err := machineRepo.GetUserMachines(user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
		var deleteRequest DeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&deleteRequest); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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
		if err := machineRepo.DeleteMachine(machine.ID); err != nil {
			log.Err(err).Msg("Error deleting machine")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func updateMachineKey(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		machine, ok := r.Context().Value(context_keys.MachineContextKey).(*models.Machine)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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
		if _, err := pqc.ValidatePublicKey(fileBytes); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		if err := machineRepo.UpdateMachinePublicKey(machine.ID, fileBytes); err != nil {
			log.Err(err).Msg("error updating machine public key")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func MachineRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.ConfigureAuth(i))
	r.Get("/{machineId}", getMachineById(i))
	r.Get("/", getMachines(i))
	r.Delete("/", deleteMachine(i))
	r.Put("/key", updateMachineKey(i))
	return r
}
