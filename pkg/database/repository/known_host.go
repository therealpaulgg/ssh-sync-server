package repository

//go:generate go run go.uber.org/mock/mockgen -source=known_host.go -destination=known_host_mock.go -package=repository

import (
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type KnownHostRepository interface {
	UpsertKnownHost(entry *models.KnownHost) (*models.KnownHost, error)
	UpsertKnownHostTx(entry *models.KnownHost, tx pgx.Tx) (*models.KnownHost, error)
}

type KnownHostRepo struct {
	Injector *do.Injector
}

func (repo *KnownHostRepo) UpsertKnownHost(entry *models.KnownHost) (*models.KnownHost, error) {
	q := do.MustInvoke[query.QueryService[models.KnownHost]](repo.Injector)
	result, err := q.QueryOne(
		"INSERT INTO known_hosts (user_id, host_pattern, key_type, key_data, marker) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (user_id, host_pattern, key_type) DO UPDATE SET key_data = $4, marker = $5 RETURNING *",
		entry.UserID, entry.HostPattern, entry.KeyType, entry.KeyData, entry.Marker,
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, sql.ErrNoRows
	}
	return result, nil
}

func (repo *KnownHostRepo) UpsertKnownHostTx(entry *models.KnownHost, tx pgx.Tx) (*models.KnownHost, error) {
	q := do.MustInvoke[query.QueryServiceTx[models.KnownHost]](repo.Injector)
	result, err := q.QueryOne(tx,
		"INSERT INTO known_hosts (user_id, host_pattern, key_type, key_data, marker) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (user_id, host_pattern, key_type) DO UPDATE SET key_data = $4, marker = $5 RETURNING *",
		entry.UserID, entry.HostPattern, entry.KeyType, entry.KeyData, entry.Marker,
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, sql.ErrNoRows
	}
	return result, nil
}
