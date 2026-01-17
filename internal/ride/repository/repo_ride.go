package repository

import (
	"context"
	"database/sql"
	"errors"

	"ride-hail/internal/ride/domain/models"
	"ride-hail/internal/ride/domain/ports"
	"ride-hail/internal/shared/postgres"
)

var (
	ErrDBNoConnection = errors.New("DB no connection")
	ErrNotFound       = errors.New("ride not found")
)

type RideRepo struct {
	db *postgres.Database
}

func NewRideRepo(db *postgres.Database) ports.RideRepository {
	return &RideRepo{db: db}
}

// CreateRide inserts a new ride into the database
func (r *RideRepo) CreateRide(ctx context.Context, ride *models.Ride) error {
	query := `INSERT INTO rides (
		passenger_id,
		status,
		pickup_location,
		destination_location
	)
	VALUES ($1, $2, $3, $4)
	RETURNING id, requested_at, created_at, updated_at`

	err := r.db.QueryRow(
		ctx,
		query,
		ride.PassengerID,
		ride.Status,
		ride.PickupLocation,
		ride.DestinationLocation,
	).Scan(&ride.ID, &ride.RequestedAt, &ride.CreatedAt, &ride.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

// GetRide fetches a ride by its ID
func (r *RideRepo) GetRide(ctx context.Context, id string) (models.Ride, error) {
	query := `SELECT id, passenger_id, status, pickup_location, destination_location, requested_at, created_at, updated_at
	FROM rides WHERE id = $1`

	var ride models.Ride
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ride.ID,
		&ride.PassengerID,
		&ride.Status,
		&ride.PickupLocation,
		&ride.DestinationLocation,
		&ride.RequestedAt,
		&ride.CreatedAt,
		&ride.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Ride{}, ErrNotFound
		}
		return models.Ride{}, err
	}

	return ride, nil
}

// ListByPassenger fetches all rides for a specific passenger
func (r *RideRepo) ListByPassenger(ctx context.Context, passengerID string) ([]models.Ride, error) {
	query := `SELECT id, passenger_id, status, pickup_location, destination_location, requested_at, created_at, updated_at
    FROM rides WHERE passenger_id = $1`

	rows, err := r.db.Query(ctx, query, passengerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []models.Ride
	for rows.Next() {
		var ride models.Ride
		if err := rows.Scan(
			&ride.ID,
			&ride.PassengerID,
			&ride.Status,
			&ride.PickupLocation,
			&ride.DestinationLocation,
			&ride.RequestedAt,
			&ride.CreatedAt,
			&ride.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rides = append(rides, ride)
	}

	// Check for errors from the rows iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rides, nil
}

// ListByStatus fetches all rides for a specific passenger with a specific status
func (r *RideRepo) ListByStatus(ctx context.Context, passengerID string, status string) ([]models.Ride, error) {
	query := `SELECT id, passenger_id, status, pickup_location, destination_location, requested_at, created_at, updated_at
    FROM rides WHERE passenger_id = $1 AND status = $2`

	rows, err := r.db.Query(ctx, query, passengerID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []models.Ride
	for rows.Next() {
		var ride models.Ride
		if err := rows.Scan(
			&ride.ID,
			&ride.PassengerID,
			&ride.Status,
			&ride.PickupLocation,
			&ride.DestinationLocation,
			&ride.RequestedAt,
			&ride.CreatedAt,
			&ride.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rides = append(rides, ride)
	}

	// Check for errors from the rows iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rides, nil
}

// UpdateRide updates an existing ride in the database
func (r *RideRepo) UpdateRide(ctx context.Context, ride models.Ride) error {
	query := `UPDATE rides 
	SET 
		status = $1, 
		pickup_location = $2, 
		destination_location = $3, 
		updated_at = NOW() 
	WHERE id = $4`

	result, err := r.db.Exec(
		ctx,
		query,
		ride.Status,
		ride.PickupLocation,
		ride.DestinationLocation,
		ride.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// CloseRide marks a ride as closed in the database (not implemented yet)
func (r *RideRepo) CloseRide(ctx context.Context, id string, reason string) error {
	query := `UPDATE rides SET status = 'CANCELLED', cancellation_reason = $1, cancelled_at = NOW(), updated_at = NOW() WHERE id = $2`

	result, err := r.db.Exec(ctx, query, reason, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *RideRepo) UpdateStatus(ctx context.Context, rideID string, status string) error {
	query := `UPDATE rides SET status = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.db.Exec(ctx, query, status, rideID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

type DB struct {
	db *sql.DB
}
