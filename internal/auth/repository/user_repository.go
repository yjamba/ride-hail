package repository

import (
	"context"
	"database/sql"
	"ride-hail/internal/auth/domain/models"
	"ride-hail/internal/auth/domain/ports"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) ports.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Save implements ports.UserRepository.
func (u *UserRepository) Save(ctx context.Context, user *models.User) (string, error) {
	tx, err := u.db.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	query := "INSERT INTO users (email, role, password_hash) VALUES ($1, $2, $3) RETURNING id"

	var id string

	err = tx.QueryRowContext(ctx, query, user.Email, user.Password, user.Role).Scan(&id)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return id, nil
}

type DB struct {
	db *sql.DB
}
