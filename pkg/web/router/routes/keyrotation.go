package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-common/pkg/dto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware/context_keys"
)

func postKeyRotation(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		machine, ok := r.Context().Value(context_keys.MachineContextKey).(*models.Machine)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var req dto.MasterKeyRotationRequestDto
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		machineRepo := do.MustInvoke[repository.MachineRepository](i)
		machines, err := machineRepo.GetUserMachines(user.ID)
		if err != nil {
			log.Err(err).Msg("postKeyRotation: error fetching machines")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		machineIDSet := make(map[string]bool, len(machines))
		for _, m := range machines {
			machineIDSet[m.ID.String()] = true
		}

		rotationRepo := do.MustInvoke[repository.MasterKeyRotationRepository](i)

		for _, entry := range req.Keys {
			if !machineIDSet[entry.MachineID.String()] {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			// The rotating machine already has the new key saved locally; skip its entry.
			if entry.MachineID == machine.ID {
				continue
			}
			if err := rotationRepo.UpsertRotation(entry.MachineID, entry.EncryptedMasterKey); err != nil {
				log.Err(err).Msg("postKeyRotation: error upserting rotation")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getKeyRotation(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		machine, ok := r.Context().Value(context_keys.MachineContextKey).(*models.Machine)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rotationRepo := do.MustInvoke[repository.MasterKeyRotationRepository](i)
		rotation, err := rotationRepo.GetRotationForMachine(machine.ID)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Err(err).Msg("getKeyRotation: error fetching rotation")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(dto.EncryptedMasterKeyDto{EncryptedMasterKey: rotation.EncryptedMasterKey})
	}
}

func deleteKeyRotation(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		machine, ok := r.Context().Value(context_keys.MachineContextKey).(*models.Machine)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rotationRepo := do.MustInvoke[repository.MasterKeyRotationRepository](i)
		if err := rotationRepo.DeleteRotationForMachine(machine.ID); err != nil {
			log.Err(err).Msg("deleteKeyRotation: error deleting rotation")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func KeyRotationRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.ConfigureAuth(i))
	r.Post("/", postKeyRotation(i))
	r.Get("/", getKeyRotation(i))
	r.Delete("/", deleteKeyRotation(i))
	return r
}
