package repository

import (
	"context"
	"time"

	"ride-hail/internal/admin/domain/models"
	"ride-hail/internal/admin/domain/ports"
	"ride-hail/internal/shared/postgres"
)

type MetricsRepository struct {
	db *postgres.Database
}

func NewMetricsRepository(db *postgres.Database) ports.MetricsRepository {
	return &MetricsRepository{
		db: db,
	}
}

// FetchOverview implements [ports.MetricsRepository].
func (m *MetricsRepository) FetchOverview(ctx context.Context) (*models.Overview, error) {
	overview := &models.Overview{
		Time:    time.Now(),
		Metrics: &models.Metrics{},
	}

	// Fetch main metrics
	metricsQuery := `
        SELECT 
            COUNT(CASE WHEN r.status IN ('IN_PROGRESS', 'EN_ROUTE', 'ARRIVED') THEN 1 END) as active_rides,
            COUNT(CASE WHEN d.status = 'AVAILABLE' THEN 1 END) as available_drivers,
            COUNT(CASE WHEN d.status = 'BUSY' THEN 1 END) as busy_drivers,
            COUNT(CASE WHEN r.created_at::date = CURRENT_DATE THEN 1 END) as total_rides_today,
            COALESCE(SUM(CASE WHEN r.created_at::date = CURRENT_DATE THEN r.final_fare END), 0) as total_revenue_today,
            COALESCE(AVG(CASE WHEN r.matched_at IS NOT NULL THEN EXTRACT(EPOCH FROM (r.matched_at - r.requested_at))/60 END), 0) as avg_wait_time,
            COALESCE(AVG(CASE WHEN r.completed_at IS NOT NULL THEN EXTRACT(EPOCH FROM (r.completed_at - r.started_at))/60 END), 0) as avg_ride_duration,
            CASE 
                WHEN COUNT(r.id) > 0 THEN 
                    (COUNT(CASE WHEN r.status = 'CANCELLED' AND r.created_at::date = CURRENT_DATE THEN 1 END)::float / COUNT(CASE WHEN r.created_at::date = CURRENT_DATE THEN 1 END)) * 100
                ELSE 0 
            END as cancellation_rate
        FROM rides r
        FULL OUTER JOIN drivers d ON d.user_id = r.driver_id
    `

	err := m.db.QueryRow(ctx, metricsQuery).Scan(
		&overview.Metrics.ActiveRides,
		&overview.Metrics.AvailableDrivers,
		&overview.Metrics.BusyDrivers,
		&overview.Metrics.TotalRidesToday,
		&overview.Metrics.TotalRevenueToday,
		&overview.Metrics.AverageWaitTimeMinutes,
		&overview.Metrics.AverageRideDurationMinutes,
		&overview.Metrics.CancellationRate,
	)
	if err != nil {
		return nil, err
	}

	// Fetch driver distribution by vehicle type
	distributionQuery := `
        SELECT 
            vehicle_type,
            COUNT(*) as count
        FROM drivers
        WHERE status IN ('AVAILABLE', 'BUSY')
        GROUP BY vehicle_type
    `

	rows, err := m.db.Query(ctx, distributionQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	overview.DriverDistribution = &models.DriverDistribution{}
	for rows.Next() {
		var vehicleType string
		var count int
		if err := rows.Scan(&vehicleType, &count); err != nil {
			return nil, err
		}

		switch vehicleType {
		case "ECONOMY":
			overview.DriverDistribution.Economy = count
		case "PREMIUM":
			overview.DriverDistribution.Premium = count
		case "XL":
			overview.DriverDistribution.XL = count
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch hotspots (areas with high ride activity)
	hotspotsQuery := `
        SELECT 
            c.address as location,
            COUNT(CASE WHEN r.status IN ('IN_PROGRESS', 'EN_ROUTE', 'ARRIVED') THEN 1 END) as active_rides,
            COUNT(CASE WHEN r.status = 'REQUESTED' THEN 1 END) as waiting_rides
        FROM rides r
        JOIN coordinates c ON r.pickup_coordinate_id = c.id
        WHERE r.created_at >= NOW() - INTERVAL '1 hour'
        GROUP BY c.address
        HAVING COUNT(r.id) > 0
        ORDER BY active_rides DESC, waiting_rides DESC
        LIMIT 10
    `

	hotspotRows, err := m.db.Query(ctx, hotspotsQuery)
	if err != nil {
		return nil, err
	}
	defer hotspotRows.Close()

	overview.Hotspots = []models.Hotspots{}
	for hotspotRows.Next() {
		var hotspot models.Hotspots
		if err := hotspotRows.Scan(
			&hotspot.Location,
			&hotspot.ActiveRides,
			&hotspot.WaitingRides,
		); err != nil {
			return nil, err
		}
		overview.Hotspots = append(overview.Hotspots, hotspot)
	}

	if err := hotspotRows.Err(); err != nil {
		return nil, err
	}

	return overview, nil
}
