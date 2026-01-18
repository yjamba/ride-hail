package repository

import (
	"context"

	"ride-hail/internal/admin/domain/models"
	"ride-hail/internal/admin/domain/ports"
	"ride-hail/internal/shared/postgres"
)

type RidesRepository struct {
	db *postgres.Database
}

func NewRidesRepository(db *postgres.Database) ports.RidesRepository {
	return &RidesRepository{
		db: db,
	}
}

// FetchRidesList implements [ports.RidesRepository].
func (r *RidesRepository) FetchRidesList(ctx context.Context, page int, pageSize int) (*models.RidesList, error) {
	offset := (page - 1) * pageSize

	ridesList := &models.RidesList{
		Rides:    []models.Ride{},
		Page:     page,
		PageSize: pageSize,
	}

	// Fetch total count
	countQuery := `
        SELECT COUNT(*) 
        FROM rides 
        WHERE status IN ('IN_PROGRESS', 'EN_ROUTE', 'ARRIVED')
    `
	err := r.db.QueryRow(ctx, countQuery).Scan(&ridesList.TotalCount)
	if err != nil {
		return nil, err
	}

	// Fetch active rides with detailed information
	ridesQuery := `
        SELECT 
            r.id as ride_id,
            r.ride_number,
            r.status,
            r.passenger_id,
            r.driver_id,
            pc.address as pickup_address,
            dc.address as destination_address,
            r.started_at,
            r.started_at + (COALESCE(dc.duration_minutes, 30) * INTERVAL '1 minute') as estimated_completion,
            COALESCE(lh.latitude, pc.latitude) as current_lat,
            COALESCE(lh.longitude, pc.longitude) as current_lng,
            COALESCE(dc.distance_km, 0) as total_distance,
            COALESCE(dc.distance_km, 0) - COALESCE(
                (SELECT SUM(
                    6371 * acos(
                        cos(radians(lh2.latitude)) * cos(radians(LAG(lh2.latitude) OVER (ORDER BY lh2.recorded_at))) * 
                        cos(radians(LAG(lh2.longitude) OVER (ORDER BY lh2.recorded_at)) - radians(lh2.longitude)) + 
                        sin(radians(lh2.latitude)) * sin(radians(LAG(lh2.latitude) OVER (ORDER BY lh2.recorded_at)))
                    )
                )
                FROM location_history lh2
                WHERE lh2.ride_id = r.id
                ), 0
            ) as distance_remaining
        FROM rides r
        JOIN coordinates pc ON r.pickup_coordinate_id = pc.id
        JOIN coordinates dc ON r.dropoff_coordinate_id = dc.id
        LEFT JOIN LATERAL (
            SELECT latitude, longitude 
            FROM location_history 
            WHERE ride_id = r.id 
            ORDER BY recorded_at DESC 
            LIMIT 1
        ) lh ON true
        WHERE r.status IN ('IN_PROGRESS', 'EN_ROUTE', 'ARRIVED')
        ORDER BY r.started_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.Query(ctx, ridesQuery, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ride models.Ride
		err := rows.Scan(
			&ride.RideID,
			&ride.RideNumber,
			&ride.Status,
			&ride.PassengerID,
			&ride.DriverID,
			&ride.PickupAddress,
			&ride.DestinationAddress,
			&ride.StartedAt,
			&ride.EstimatedCompletion,
			&ride.CurrentDriverLocation.Latitude,
			&ride.CurrentDriverLocation.Longitude,
			&ride.DistanceCompletedKM,
			&ride.DistanceRemainingKM,
		)
		if err != nil {
			return nil, err
		}

		// Calculate completed distance
		ride.DistanceCompletedKM = ride.DistanceCompletedKM - ride.DistanceRemainingKM

		ridesList.Rides = append(ridesList.Rides, ride)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ridesList, nil
}
