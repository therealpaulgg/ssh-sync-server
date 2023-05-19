package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
)

type DeleteRequest struct {
	MachineName string `json:"machine_name"`
}

func getMachineById(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
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
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
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
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
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

func MachineRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.ConfigureAuth(i))
	r.Get("/{machineId}", getMachineById(i))
	r.Get("/", getMachines(i))
	r.Delete("/", deleteMachine(i))
	return r
}
