package repositories

import (
	"context"

	"ride-hail/internal/driver/domain/models"
	"ride-hail/internal/driver/domain/ports"
	"ride-hail/internal/shared/postgres"
)

type HistoryLocationRepository struct {
	db *postgres.Database
}

func NewHistoryLocationRepository(db *postgres.Database) ports.HistoryLocationRepository {
	return &HistoryLocationRepository{
		db: db,
	}
}

// AddLocation implements [ports.HistoryLocationRepository].
func (h *HistoryLocationRepository) AddLocation(ctx context.Context, historyLocation *models.LocationHistory) error {
	q := `INSERT INTO location_histories 
			(coordinate_id, driver_id, latitude, longitude, accuracy_meters, speed_kmh, heading_degrees, recorded_at, ride_id) 
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := h.db.Exec(ctx, q,
		historyLocation.CoordinateID,
		historyLocation.DriverID,
		historyLocation.Latitude,
		historyLocation.Longitude,
		historyLocation.AccuracyMeters,
		historyLocation.SpeedKmh,
		historyLocation.HeadingDegrees,
		historyLocation.RecordedAt,
		historyLocation.RideID,
	)
	return err
}
