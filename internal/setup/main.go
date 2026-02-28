package setup

import (
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
)

func SetupServices(i *do.Injector) {
	do.Provide(i, database.NewDataAccessorService)
	do.Provide(i, func(i *do.Injector) (query.TransactionService, error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.TransactionServiceImpl{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (query.QueryService[models.Machine], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceImpl[models.Machine]{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (query.QueryServiceTx[models.Machine], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceTxImpl[models.Machine]{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (query.QueryService[models.User], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceImpl[models.User]{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (query.QueryServiceTx[models.User], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceTxImpl[models.User]{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (query.QueryService[models.SshKey], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceImpl[models.SshKey]{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (query.QueryServiceTx[models.SshKey], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceTxImpl[models.SshKey]{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceImpl[models.SshConfig]{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (query.QueryServiceTx[models.SshConfig], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceTxImpl[models.SshConfig]{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (query.QueryService[models.KnownHost], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceImpl[models.KnownHost]{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (query.QueryServiceTx[models.KnownHost], error) {
		dataAccessor := do.MustInvoke[database.DataAccessor](i)
		return &query.QueryServiceTxImpl[models.KnownHost]{DataAccessor: dataAccessor}, nil
	})
	do.Provide(i, func(i *do.Injector) (repository.UserRepository, error) {
		return &repository.UserRepo{Injector: i}, nil
	})
	do.Provide(i, func(i *do.Injector) (repository.MachineRepository, error) {
		return &repository.MachineRepo{Injector: i}, nil
	})
	do.Provide(i, func(i *do.Injector) (repository.SshKeyRepository, error) {
		return &repository.SshKeyRepo{Injector: i}, nil
	})
	do.Provide(i, func(i *do.Injector) (repository.SshConfigRepository, error) {
		return &repository.SshConfigRepo{Injector: i}, nil
	})
	do.Provide(i, func(i *do.Injector) (repository.KnownHostRepository, error) {
		return &repository.KnownHostRepo{Injector: i}, nil
	})

}
