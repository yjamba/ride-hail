package repository

import (
	"context"

	"ride-hail/internal/auth/domain/models"
	"ride-hail/internal/auth/domain/ports"
	"ride-hail/internal/shared/postgres"
)

type UserRepository struct {
	db *postgres.Database
}

func NewUserRepository(db *postgres.Database) ports.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Save implements ports.UserRepository.
func (u *UserRepository) Save(ctx context.Context, user *models.User) (string, error) {
	tx := postgres.GetTxFromContext(ctx)
	if tx != nil {
		return u.saveUser(ctx, tx, user)
	}

	var id string
	err := u.db.TxManager.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		id, err = u.saveUser(txCtx, postgres.GetTxFromContext(txCtx), user)
		return err
	})

	return id, err
}

func (u *UserRepository) saveUser(ctx context.Context, tx *postgres.Tx, user *models.User) (string, error) {
	query := "INSERT INTO users (email, role, password_hash) VALUES ($1, $2, $3) RETURNING id"

	var id string

	err := tx.QueryRow(ctx, query, user.Email, user.Role, user.Password).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}
