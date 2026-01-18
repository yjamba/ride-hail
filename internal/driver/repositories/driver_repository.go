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
            id, license_number, vehicle_type, vehicle_attrs, rating, total_rides, total_earnings, status
        FROM 
            drivers 
        WHERE 
            id = $1`
	var driver models.Driver
	var vehicleAttrsJSON []byte
	var statusStr string

	err := d.db.QueryRow(ctx, q, id).Scan(
		&driver.ID,
		&driver.LicenseNumber,
		&driver.VehicleType,
		&vehicleAttrsJSON,
		&driver.Rating,
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

	return &driver, nil
}

// Update implements [ports.DriverRepository].
func (d *DriverRepository) Update(ctx context.Context, driver *models.Driver) error {
	tx := postgres.GetTxFromContext(ctx)
	if tx != nil {
		return d.updateWithTx(ctx, tx, driver)
	}

	return d.db.TxManager.WithTx(ctx, func(txCtx context.Context) error {
		return d.updateWithTx(txCtx, postgres.GetTxFromContext(txCtx), driver)
	})
}

func (d *DriverRepository) updateWithTx(ctx context.Context, tx *postgres.Tx, driver *models.Driver) error {
	q := `UPDATE 
            drivers 
        SET 
            rating = $1,
            total_rides = $2,
            total_earnings = $3,
            updated_at = NOW()
        WHERE id = $4`
	_, err := tx.Exec(ctx, q,
		driver.Rating,
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
        SET status = $1, updated_at = NOW() 
        WHERE id = $2`
	_, err := tx.Exec(ctx, q, status.String(), id)
	return err
}

// UpdateRideStatus implements [ports.DriverRepository].
func (d *DriverRepository) UpdateRideStatus(ctx context.Context, rideID string, status models.RideStatus) error {
	tx := postgres.GetTxFromContext(ctx)
	if tx != nil {
		return d.updateRideStatusWithTx(ctx, tx, rideID, status)
	}

	return d.db.TxManager.WithTx(ctx, func(txCtx context.Context) error {
		return d.updateRideStatusWithTx(txCtx, postgres.GetTxFromContext(txCtx), rideID, status)
	})
}

func (d *DriverRepository) FindAvailableDriversNearby(
	ctx context.Context,
	lat, lon float64,
	vehicleType string,
	radiusMeters int,
) ([]models.DriverWithDistance, error) {
	q := `
SELECT d.id, u.email, d.rating, c.latitude, c.longitude,
       ST_Distance(
         ST_MakePoint(c.longitude, c.latitude)::geography,
         ST_MakePoint($1, $2)::geography
       ) / 1000 as distance_km
FROM drivers d
JOIN users u ON d.id = u.id
JOIN coordinates c ON c.entity_id = d.id
  AND c.entity_type = 'driver'
  AND c.is_current = true
WHERE d.status = 'AVAILABLE'
  AND d.vehicle_type = $3
  AND ST_DWithin(
        ST_MakePoint(c.longitude, c.latitude)::geography,
        ST_MakePoint($1, $2)::geography,
        $4
      )
ORDER BY distance_km, d.rating DESC
LIMIT 10;
`

	rows, err := d.db.Query(ctx, q, lat, lon, vehicleType, radiusMeters)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []models.DriverWithDistance
	for rows.Next() {
		var dr models.DriverWithDistance
		if err := rows.Scan(
			&dr.ID,
			&dr.Email,
			&dr.Rating,
			&dr.Latitude,
			&dr.Longitude,
			&dr.DistanceKm,
		); err != nil {
			return nil, err
		}
		drivers = append(drivers, dr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return drivers, nil
}

func (d *DriverRepository) updateRideStatusWithTx(ctx context.Context, tx *postgres.Tx, rideID string, status models.RideStatus) error {
	var query string
	var args []interface{}

	switch status {
	case models.RideStatusMatched:
		query = `UPDATE rides SET status = $1, matched_at = NOW(), updated_at = NOW() WHERE id = $2`
		args = []interface{}{status.String(), rideID}
	case models.RideStatusInProgress:
		query = `UPDATE rides SET status = $1, started_at = NOW(), updated_at = NOW() WHERE id = $2`
		args = []interface{}{status.String(), rideID}
	case models.RideStatusCompleted:
		query = `UPDATE rides SET status = $1, completed_at = NOW(), updated_at = NOW() WHERE id = $2`
		args = []interface{}{status.String(), rideID}
	case models.RideStatusArrived:
		query = `UPDATE rides SET status = $1, arrived_at = NOW(), updated_at = NOW() WHERE id = $2`
		args = []interface{}{status.String(), rideID}
	default:
		query = `UPDATE rides SET status = $1, updated_at = NOW() WHERE id = $2`
		args = []interface{}{status.String(), rideID}
	}

	_, err := tx.Exec(ctx, query, args...)
	return err
}

// GetRideByID implements [ports.DriverRepository].
func (d *DriverRepository) GetRideByID(ctx context.Context, rideID string) (*models.Ride, error) {
	q := `SELECT 
            id, ride_number, passenger_id, driver_id, vehicle_type, status, 
            estimated_fare, final_fare, created_at
        FROM rides 
        WHERE id = $1`

	var ride models.Ride
	var driverID, finalFare *string
	var statusStr string

	err := d.db.QueryRow(ctx, q, rideID).Scan(
		&ride.ID,
		&ride.RideNumber,
		&ride.PassengerID,
		&driverID,
		&ride.VehicleType,
		&statusStr,
		&ride.EstimatedFare,
		&finalFare,
		&ride.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if driverID != nil {
		ride.DriverID = *driverID
	}

	if finalFare != nil {
		// Parse final fare
		ride.FinalFare = ride.EstimatedFare // Use estimated if final not set
	} else {
		ride.FinalFare = ride.EstimatedFare
	}

	ride.Status = models.RideStatus(statusStr)

	return &ride, nil
}
