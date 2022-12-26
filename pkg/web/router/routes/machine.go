package routes

import (
	"github.com/go-chi/chi"
	"github.com/samber/do"
)

func MachineRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	return r
}
