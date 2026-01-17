package repositories

import (
	"context"

	"ride-hail/internal/driver/domain/models"
	"ride-hail/internal/driver/domain/ports"
	"ride-hail/internal/shared/postgres"
)

type DriverRepository struct {
	db *postgres.Database
}

func NewDriverRepository(db *postgres.Database) ports.DriverRepository {
	return &DriverRepository{db: db}
}

// GetById implements [ports.DriverRepository].
func (d *DriverRepository) GetById(ctx context.Context, id string) (models.Driver, error) {
	panic("unimplemented")
}

// Update implements [ports.DriverRepository].
func (d *DriverRepository) Update(ctx context.Context, id string) error {
	panic("unimplemented")
}

// UpdateStatus implements [ports.DriverRepository].
func (d *DriverRepository) UpdateStatus(ctx context.Context, id string, status models.DriverStatus) error {
	q := `UPDATE drivers SET status = $1 WHERE id = $2`
	_, err := d.db.Exec(ctx, q, status.String(), id)
	return err
}
