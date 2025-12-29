package repository

import (
	"context"
	"encoding/json"
	"ride-hail/internal/auth/domain/models"
	"ride-hail/internal/auth/domain/ports"
)

type DriverRepository struct {
	db *DB
}

func NewDriverRepository(db *DB) ports.DriverRepository {
	return &DriverRepository{db: db}
}

func (d *DriverRepository) Save(ctx context.Context, driver *models.Driver) (string, error) {
	tx, err := d.db.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	query := "INSERT INTO drivers (user_id, license_number, vehicle_info) VALUES ($1, $2, $3) RETURNING id"

	var id string

	vechicleInfo, err := json.Marshal(driver.VehicleAttrs)
	if err != nil {
		return "", err
	}

	err = tx.QueryRowContext(ctx, query, driver.UserID, driver.LicenseNumber, vechicleInfo).Scan(&id)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return id, nil
}
