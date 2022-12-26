package routes

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
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
		fmt.Fprintln(w, user)
	})

	return r
}
