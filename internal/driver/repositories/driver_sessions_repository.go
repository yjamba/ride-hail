package repositories

import (
	"context"
	"ride-hail/internal/driver/domain/models"
	"ride-hail/internal/driver/domain/ports"
	"ride-hail/internal/shared/postgres"
)

type DriverSessionsRepository struct {
	db *postgres.Database
}

func NewDriverSessionsRepository(db *postgres.Database) ports.DriverSessionsRepository {
	return &DriverSessionsRepository{
		db: db,
	}
}

// GetById implements [ports.DriverSessionsRepository].
func (h *DriverSessionsRepository) GetById(ctx context.Context, id string) (*models.DriverSession, error) {
	q := `SELECT 
			id, driver_id, started_at, ended_at, total_rides, total_earnings
		FROM driver_sessions WHERE id = $1`
	driverSession := &models.DriverSession{}
	err := h.db.QueryRow(ctx, q, id).Scan(
		&driverSession.ID,
		&driverSession.DriverID,
		&driverSession.StartedAt,
		&driverSession.EndedAt,
		&driverSession.TotalRides,
		&driverSession.TotalEarnings,
	)
	if err != nil {
		return nil, err
	}
	return driverSession, nil
}

func (h *DriverSessionsRepository) GetActiveByDriverID(ctx context.Context, driverID string) (*models.DriverSession, error) {
	q := `SELECT 
            id, driver_id, started_at, ended_at, total_rides, total_earnings
        FROM driver_sessions 
        WHERE driver_id = $1 AND ended_at IS NULL
        ORDER BY started_at DESC
        LIMIT 1`

	driverSession := &models.DriverSession{}
	err := h.db.QueryRow(ctx, q, driverID).Scan(
		&driverSession.ID,
		&driverSession.DriverID,
		&driverSession.StartedAt,
		&driverSession.EndedAt,
		&driverSession.TotalRides,
		&driverSession.TotalEarnings,
	)
	if err != nil {
		return nil, err
	}

	// if driverSession.EndedAt != nil {
	// 	return nil, errors.New("no active session found")
	// }

	return driverSession, nil
}

// Create implements [ports.DriverSessionsRepository].
func (h *DriverSessionsRepository) Create(ctx context.Context, driverId string) (string, error) {
	tx := postgres.GetTxFromContext(ctx)
	if tx != nil {
		return h.createWithTx(ctx, tx, driverId)
	}

	var sessionId string
	err := h.db.TxManager.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		sessionId, err = h.createWithTx(txCtx, postgres.GetTxFromContext(txCtx), driverId)
		return err
	})
	if err != nil {
		return "", err
	}

	return sessionId, nil
}

// Update implements [ports.DriverSessionsRepository].
func (h *DriverSessionsRepository) Update(ctx context.Context, driverSession *models.DriverSession) error {
	q := `UPDATE driver_sessions 
		SET total_rides = $1, total_earnings = $2 
		WHERE id = $3`
	_, err := h.db.Exec(ctx, q,
		driverSession.TotalRides,
		driverSession.TotalEarnings,
		driverSession.ID,
	)
	return err
}

// Close implements [ports.DriverSessionsRepository].
func (h *DriverSessionsRepository) Close(ctx context.Context, driverSessionId string) error {
	tx := postgres.GetTxFromContext(ctx)
	if tx != nil {
		return h.closeWithTx(ctx, tx, driverSessionId)
	}

	err := h.db.TxManager.WithTx(ctx, func(txCtx context.Context) error {
		return h.closeWithTx(txCtx, postgres.GetTxFromContext(txCtx), driverSessionId)
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *DriverSessionsRepository) createWithTx(ctx context.Context, tx *postgres.Tx, driverId string) (string, error) {
	q := `INSERT INTO driver_sessions (driver_id, started_at) VALUES ($1, NOW()) RETURNING id`
	var id string
	err := tx.QueryRow(ctx, q, driverId).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *DriverSessionsRepository) closeWithTx(ctx context.Context, tx *postgres.Tx, driverSessionId string) error {
	q := `UPDATE driver_sessions SET ended_at = NOW() WHERE id = $1`
	_, err := tx.Exec(ctx, q, driverSessionId)
	return err
}
