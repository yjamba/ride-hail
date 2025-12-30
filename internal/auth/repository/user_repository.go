package repository

import (
	"context"
	"database/sql"

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
	tx, err := u.db.BeginTx(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	query := "INSERT INTO users (email, role, password_hash) VALUES ($1, $2, $3) RETURNING id"

	var id string

	err = tx.QueryRow(ctx, query, user.Email, user.Password, user.Role).Scan(&id)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return id, nil
}

type DB struct {
	db *sql.DB
}
