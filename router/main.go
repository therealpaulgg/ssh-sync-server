package router

import (
	"fmt"
	"net/http"

	"github.com/therealpaulgg/ssh-sync-server/middleware"

	"github.com/go-chi/chi"
)

func Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Log)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})
	r.Get("/upload", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "TODO")
	})
	r.Get("/download", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "TODO")
	})
	return r
}
