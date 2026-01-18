package ports

import (
	"context"

	"ride-hail/internal/auth/domain/models"
)

type UserRepository interface {
	Save(ctx context.Context, user *models.User) (string, error)
}
