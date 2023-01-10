package router

import (
	"fmt"

	"net/http"

	"github.com/go-chi/chi"

	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/router/routes"
)

func Router(i *do.Injector) chi.Router {
	baseRouter := chi.NewRouter()
	baseRouter.Use(middleware.Log)

	apiV1Router := chi.NewRouter()
	apiV1Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})
	apiV1Router.Mount("/users", routes.UserRoutes(i))
	apiV1Router.Mount("/setup", routes.SetupRoutes(i))
	apiV1Router.Mount("/machines", routes.MachineRoutes(i))
	apiV1Router.Mount("/data", routes.DataRoutes(i))
	baseRouter.Mount("/api/v1", apiV1Router)
	return baseRouter
}
