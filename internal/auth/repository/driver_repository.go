package repository

import (
	"context"
	"encoding/json"

	"ride-hail/internal/auth/domain/models"
	"ride-hail/internal/auth/domain/ports"
	"ride-hail/internal/shared/postgres"
)

type DriverRepository struct {
	db *postgres.Database
}

func NewDriverRepository(db *postgres.Database) ports.DriverRepository {
	return &DriverRepository{db: db}
}

func (d *DriverRepository) Save(ctx context.Context, driver *models.Driver) (string, error) {
	tx := postgres.GetTxFromContext(ctx)
	if tx != nil {
		// Используем существующую транзакцию
		return d.saveWithTx(ctx, tx, driver)
	}

	var id string
	err := d.db.TxManager.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		id, err = d.saveWithTx(txCtx, postgres.GetTxFromContext(txCtx), driver)
		return err
	})
	return id, err
}

func (d *DriverRepository) saveWithTx(ctx context.Context, tx *postgres.Tx, driver *models.Driver) (string, error) {
	query := "INSERT INTO drivers (user_id, license_number, vehicle_attrs) VALUES ($1, $2, $3) RETURNING id"
	var id string

	vehicleInfo, err := json.Marshal(driver.VehicleAttrs)
	if err != nil {
		return "", err
	}

	err = tx.QueryRow(ctx, query, driver.UserID, driver.LicenseNumber, vehicleInfo).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}
