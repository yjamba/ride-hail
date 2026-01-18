package repositories

import (
	"context"
	"ride-hail/internal/driver/domain/models"
	"ride-hail/internal/shared/postgres"
)

type CoordinateRepository struct {
	db *postgres.Database
}

func NewCoordinateRepository(db *postgres.Database) *CoordinateRepository {
	return &CoordinateRepository{db: db}
}

// CreateOrUpdate sets previous coordinates to is_current=false and creates new one
func (r *CoordinateRepository) CreateOrUpdate(ctx context.Context, entityID, entityType string,
	lat, lon float64, address string,
) (string, error) {
	tx := postgres.GetTxFromContext(ctx)
	if tx != nil {
		return r.createOrUpdateWithTx(ctx, tx, entityID, entityType, lat, lon, address)
	}

	var coordID string
	err := r.db.TxManager.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		coordID, err = r.createOrUpdateWithTx(txCtx, postgres.GetTxFromContext(txCtx),
			entityID, entityType, lat, lon, address)
		return err
	})

	return coordID, err
}

func (r *CoordinateRepository) createOrUpdateWithTx(ctx context.Context, tx *postgres.Tx,
	entityID, entityType string, lat, lon float64, address string,
) (string, error) {
	// Set previous coordinates to not current
	_, err := tx.Exec(ctx, `
        UPDATE coordinates 
        SET is_current = false, updated_at = NOW()
        WHERE entity_id = $1 AND entity_type = $2 AND is_current = true
    `, entityID, entityType)
	if err != nil {
		return "", err
	}

	// Insert new coordinate
	var coordID string
	err = tx.QueryRow(ctx, `
        INSERT INTO coordinates (entity_id, entity_type, latitude, longitude, address, is_current)
        VALUES ($1, $2, $3, $4, $5, true)
        RETURNING id
    `, entityID, entityType, lat, lon, address).Scan(&coordID)

	return coordID, err
}

func (r *CoordinateRepository) GetCurrent(ctx context.Context, entityID, entityType string) (*models.Coordinate, error) {
	query := `
        SELECT id, latitude, longitude, address, created_at
        FROM coordinates
        WHERE entity_id = $1 AND entity_type = $2 AND is_current = true
    `

	var coord models.Coordinate
	err := r.db.QueryRow(ctx, query, entityID, entityType).Scan(
		&coord.ID,
		&coord.Latitude,
		&coord.Longitude,
		&coord.Address,
		&coord.CreatedAt,
	)

	return &coord, err
}
