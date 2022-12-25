package routes

import (
	"github.com/go-chi/chi"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

func MachineRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	do.Provide(i, func(i *do.Injector) (query.QueryService[models.Machine], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceImpl[models.Machine]{DataAccessor: dataAccessor}, nil
	})
	return r
}
