package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/dto"
)

func UserRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Get("/{username}", func(w http.ResponseWriter, r *http.Request) {
		user := models.User{}
		user.Username = chi.URLParam(r, "username")
		err := user.GetUserByUsername(i)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err = user.GetUserMachines(i); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		userDto := dto.UserDto{
			Username: user.Username,
			Machines: lo.Map(user.Machines, func(m models.Machine, i int) dto.MachineDto {
				return dto.MachineDto{
					Name: m.Name,
				}
			}),
		}
		json.NewEncoder(w).Encode(userDto)
	})

	return r
}
