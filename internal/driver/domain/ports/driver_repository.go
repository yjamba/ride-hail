package ports

import "ride-hail/internal/driver/domain/models"

type DriverRepository interface {
	GetById(id string) (models.Driver, error)
	Update(id string) error
}
