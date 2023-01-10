package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
)

type DeleteRequest struct {
	MachineName string `json:"machine_name"`
}

func MachineRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.ConfigureAuth(i))
	r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
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
		machine := models.Machine{}
		machine.Name = deleteRequest.MachineName
		machine.UserID = user.ID
		if err := machine.GetMachineByNameAndUser(i); errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			log.Err(err).Msg("Error getting machine by name and user")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := machine.DeleteMachine(i); err != nil {
			log.Err(err).Msg("Error deleting machine")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return r
}
