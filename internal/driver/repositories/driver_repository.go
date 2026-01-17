package repositories

import (
	"context"
	"encoding/json"

	"ride-hail/internal/driver/domain/models"
	"ride-hail/internal/driver/domain/ports"
	"ride-hail/internal/shared/postgres"
)

type DriverRepository struct {
	db *postgres.Database
}

func NewDriverRepository(db *postgres.Database) ports.DriverRepository {
	return &DriverRepository{
		db: db,
	}
}

// GetById implements [ports.DriverRepository].
func (d *DriverRepository) GetById(ctx context.Context, id string) (*models.Driver, error) {
	q := `SELECT 
            id, user_id, license_number, vehicle_type, vehicle_attrs, total_rides, total_earnings, status
        FROM 
            drivers 
        WHERE 
            id = $1`
	var driver *models.Driver
	var vehicleAttrsJSON []byte
	var statusStr string

	err := d.db.QueryRow(ctx, q, id).Scan(
		&driver.ID,
		&driver.UserID,
		&driver.LicenseNumber,
		&driver.VehicleType,
		&vehicleAttrsJSON,
		&driver.TotalRides,
		&driver.TotalEarnings,
		&statusStr,
	)
	if err != nil {
		return nil, err
	}

	if len(vehicleAttrsJSON) > 0 {
		if err := json.Unmarshal(vehicleAttrsJSON, &driver.VehicleAttrs); err != nil {
			return nil, err
		}
	}

	driver.Status = models.DriverStatus(statusStr)

	return driver, nil
}

// Update implements [ports.DriverRepository].
func (d *DriverRepository) Update(ctx context.Context, driver *models.Driver) error {
	q := `UPDATE 
			drivers 
		SET 
			rating = $1,
			total_rides = $2,
			total_earnings = $3,
			updated_at = NOW()
		WHERE id = $4`
	_, err := d.db.Exec(ctx, q,
		driver.TotalRides,
		driver.TotalEarnings,
		driver.ID,
	)
	return err
}

// UpdateStatus implements [ports.DriverRepository].
func (d *DriverRepository) UpdateStatus(ctx context.Context, id string, status models.DriverStatus) error {
	tx := postgres.GetTxFromContext(ctx)
	if tx != nil {
		return d.updateStatusWithTx(ctx, tx, id, status)
	}

	err := d.db.TxManager.WithTx(ctx, func(txCtx context.Context) error {
		return d.updateStatusWithTx(txCtx, postgres.GetTxFromContext(txCtx), id, status)
	})
	return err
}

func (d *DriverRepository) updateStatusWithTx(ctx context.Context, tx *postgres.Tx, id string, status models.DriverStatus) error {
	q := `UPDATE 
			drivers 
		SET status = $1 
		WHERE id = $2`
	_, err := tx.Exec(ctx, q, status.String(), id)
	return err
}
