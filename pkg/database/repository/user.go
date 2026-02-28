package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"

	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

// UserRepository interface for User repository
type UserRepository interface {
	GetUser(id uuid.UUID) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
	CreateUserTx(user *models.User, tx pgx.Tx) (*models.User, error)
	DeleteUser(id uuid.UUID) error
	GetUserConfig(id uuid.UUID) ([]models.SshConfig, error)
	GetUserKeys(id uuid.UUID) ([]models.SshKey, error)
	GetUserKey(userId uuid.UUID, keyId uuid.UUID) (*models.SshKey, error)
	GetUserKnownHosts(id uuid.UUID) ([]models.KnownHost, error)
	AddAndUpdateKeys(user *models.User) error
	AddAndUpdateKeysTx(user *models.User, tx pgx.Tx) error
	AddAndUpdateConfig(user *models.User) error
	AddAndUpdateConfigTx(user *models.User, tx pgx.Tx) error
	AddAndUpdateKnownHostsTx(user *models.User, tx pgx.Tx) error
	DeleteUserKeyTx(user *models.User, id uuid.UUID, tx pgx.Tx) error
}

type UserRepo struct {
	Injector *do.Injector
}

var ErrUserAlreadyExists = errors.New("user already exists")

func (repo *UserRepo) GetUser(userId uuid.UUID) (*models.User, error) {
	q := do.MustInvoke[query.QueryService[models.User]](repo.Injector)
	user, err := q.QueryOne("select * from users where id = $1", userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (repo *UserRepo) GetUserByUsername(username string) (*models.User, error) {
	q := do.MustInvoke[query.QueryService[models.User]](repo.Injector)
	user, err := q.QueryOne("select * from users where username = $1", username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (repo *UserRepo) CreateUser(user *models.User) (*models.User, error) {
	q := do.MustInvoke[query.QueryService[models.User]](repo.Injector)
	existingUser, err := q.QueryOne("select * from users where username = $1", user.Username)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}
	newUser, err := q.QueryOne("insert into users (username) values ($1) returning *", user.Username)
	if err != nil {
		return nil, err
	}
	if newUser == nil {
		return nil, sql.ErrNoRows
	}
	return newUser, nil
}

func (repo *UserRepo) CreateUserTx(user *models.User, tx pgx.Tx) (*models.User, error) {
	q := do.MustInvoke[query.QueryServiceTx[models.User]](repo.Injector)
	existingUser, err := q.QueryOne(tx, "select * from users where username = $1", user.Username)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}
	newUser, err := q.QueryOne(tx, "insert into users (username) values ($1) returning *", user.Username)
	if err != nil {
		return nil, err
	}
	if newUser == nil {
		return nil, sql.ErrNoRows
	}
	return newUser, nil
}

func (repo *UserRepo) DeleteUser(id uuid.UUID) error {
	q := do.MustInvoke[database.DataAccessor](repo.Injector)
	tx, err := q.GetConnection().BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil && !errors.Is(err, pgx.ErrTxCommitRollback) {
			tx.Rollback(context.TODO())
		}
	}()
	if _, err := tx.Exec(context.TODO(), "delete from ssh_keys where user_id = $1", id); err != nil {
		return err
	}
	if _, err := tx.Exec(context.TODO(), "delete from ssh_configs where user_id = $1", id); err != nil {
		return err
	}
	if _, err := tx.Exec(context.TODO(), "delete from known_hosts where user_id = $1", id); err != nil {
		return err
	}
	if _, err := tx.Exec(context.TODO(), "delete from machines where user_id = $1", id); err != nil {
		return err
	}
	if _, err := tx.Exec(context.TODO(), "delete from users where id = $1", id); err != nil {
		return err
	}
	err = tx.Commit(context.TODO())
	return nil
}

func (repo *UserRepo) GetUserConfig(id uuid.UUID) ([]models.SshConfig, error) {
	q := do.MustInvoke[query.QueryService[models.SshConfig]](repo.Injector)
	config, err := q.Query("select * from ssh_configs where user_id = $1", id)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (repo *UserRepo) GetUserKeys(id uuid.UUID) ([]models.SshKey, error) {
	q := do.MustInvoke[query.QueryService[models.SshKey]](repo.Injector)
	keys, err := q.Query("select * from ssh_keys where user_id = $1", id)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (repo *UserRepo) AddAndUpdateKeys(user *models.User) error {
	keyRepo := do.MustInvoke[SshKeyRepository](repo.Injector)
	for i := range user.Keys {
		keyPtr := &user.Keys[i]
		newKey, err := keyRepo.UpsertSshKey(keyPtr)
		if err != nil {
			return err
		}
		*keyPtr = *newKey
	}
	return nil
}

func (repo *UserRepo) AddAndUpdateKeysTx(user *models.User, tx pgx.Tx) error {
	keyRepo := do.MustInvoke[SshKeyRepository](repo.Injector)
	for i := range user.Keys {
		keyPtr := &user.Keys[i]
		newKey, err := keyRepo.UpsertSshKeyTx(keyPtr, tx)
		if err != nil {
			return err
		}
		*keyPtr = *newKey
	}
	return nil
}

func (repo *UserRepo) AddAndUpdateConfig(user *models.User) error {
	configRepo := do.MustInvoke[SshConfigRepository](repo.Injector)
	for i := range user.Config {
		configPtr := &user.Config[i]
		newConfig, err := configRepo.UpsertSshConfig(configPtr)
		if err != nil {
			return err
		}
		*configPtr = *newConfig
	}
	return nil
}

func (repo *UserRepo) AddAndUpdateConfigTx(user *models.User, tx pgx.Tx) error {
	configRepo := do.MustInvoke[SshConfigRepository](repo.Injector)
	for i := range user.Config {
		configPtr := &user.Config[i]
		newConfig, err := configRepo.UpsertSshConfigTx(configPtr, tx)
		if err != nil {
			return err
		}
		*configPtr = *newConfig
	}
	return nil
}

func (repo *UserRepo) GetUserKnownHosts(id uuid.UUID) ([]models.KnownHost, error) {
	q := do.MustInvoke[query.QueryService[models.KnownHost]](repo.Injector)
	entries, err := q.Query("select * from known_hosts where user_id = $1", id)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (repo *UserRepo) AddAndUpdateKnownHostsTx(user *models.User, tx pgx.Tx) error {
	knownHostRepo := do.MustInvoke[KnownHostRepository](repo.Injector)
	for i := range user.KnownHosts {
		entryPtr := &user.KnownHosts[i]
		newEntry, err := knownHostRepo.UpsertKnownHostTx(entryPtr, tx)
		if err != nil {
			return err
		}
		*entryPtr = *newEntry
	}
	return nil
}

func (repo *UserRepo) GetUserKey(userId uuid.UUID, keyId uuid.UUID) (*models.SshKey, error) {
	q := do.MustInvoke[query.QueryService[models.SshKey]](repo.Injector)
	key, err := q.QueryOne("select * from ssh_keys where user_id = $1 and id = $2", userId, keyId)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, sql.ErrNoRows
	}
	return key, nil
}

func (repo *UserRepo) DeleteUserKeyTx(user *models.User, id uuid.UUID, tx pgx.Tx) error {
	_, err := tx.Exec(context.TODO(), "delete from ssh_keys where user_id = $1 and id = $2", user.ID, id)
	return err
}
