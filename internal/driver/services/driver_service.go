package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ride-hail/internal/driver/domain/models"
	"ride-hail/internal/driver/domain/ports"
)

type DriverService struct {
	repo           ports.DriverRepository
	sessionRepo    ports.DriverSessionsRepository
	locationRepo   ports.HistoryLocationRepository
	coordinateRepo ports.CoordinateRepository
	consume        ports.Consume
	publish        ports.Publish
	txManager      ports.TransactionManager
}

func NewDriverService(
	repo ports.DriverRepository,
	sessionRepo ports.DriverSessionsRepository,
	locationRepo ports.HistoryLocationRepository,
	coordinateRepo ports.CoordinateRepository,
	consume ports.Consume,
	publish ports.Publish,
	txManager ports.TransactionManager,
) *DriverService {
	return &DriverService{
		repo:           repo,
		sessionRepo:    sessionRepo,
		locationRepo:   locationRepo,
		coordinateRepo: coordinateRepo,
		consume:        consume,
		publish:        publish,
		txManager:      txManager,
	}
}

func (s *DriverService) GoOnline(ctx context.Context, driverID string, lat, lon float64) (string, error) {
	if driverID == "" {
		return "", errors.New("driverID cannot be empty")
	}

	if err := validateLatLon(lat, lon); err != nil {
		return "", err
	}

	// Start transaction
	var sessionID string
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Get current driver status
		driver, err := s.repo.GetById(txCtx, driverID)
		if err != nil {
			return fmt.Errorf("failed to get driver: %w", err)
		}

		// Verify driver can go online
		if driver.Status != models.Offline {
			return fmt.Errorf("cannot go online: current status %s", driver.Status)
		}

		// Create new session
		sessionID, err = s.sessionRepo.Create(txCtx, driverID)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Update driver status
		if err := s.repo.UpdateStatus(txCtx, driverID, models.Available); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		// Set initial location
		coordID, err := s.coordinateRepo.CreateOrUpdate(txCtx, driverID, "driver", lat, lon, "")
		if err != nil {
			return fmt.Errorf("failed to set location: %w", err)
		}

		// Record in location history
		historyLoc := &models.LocationHistory{
			CoordinateID: coordID,
			DriverID:     driverID,
			Latitude:     lat,
			Longitude:    lon,
			RecordedAt:   time.Now(),
		}

		if err := s.locationRepo.AddLocation(txCtx, historyLoc); err != nil {
			return fmt.Errorf("failed to add location history: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	// Publish driver status update
	statusUpdate := map[string]interface{}{
		"driver_id": driverID,
		"status":    models.Available.String(),
		"timestamp": time.Now(),
	}

	data, _ := json.Marshal(statusUpdate)
	routingKey := fmt.Sprintf("driver.status.%s", driverID)
	_ = s.publish.Publish("driver_topic", routingKey, data)

	return sessionID, nil
}

func (s *DriverService) GoOffline(ctx context.Context, driverID string) (*models.SessionSummary, error) {
	if driverID == "" {
		return nil, errors.New("driverID cannot be empty")
	}

	var summary *models.SessionSummary

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Get current driver
		driver, err := s.repo.GetById(txCtx, driverID)
		if err != nil {
			return fmt.Errorf("failed to get driver: %w", err)
		}

		// Verify driver can go offline (must be AVAILABLE, not BUSY or EN_ROUTE)
		if driver.Status != models.Available {
			return fmt.Errorf("cannot go offline: current status %s (complete active ride first)", driver.Status)
		}

		// Get active session
		session, err := s.sessionRepo.GetActiveByDriverID(txCtx, driverID)
		if err != nil {
			return fmt.Errorf("failed to get active session: %w", err)
		}

		// Close the session
		if err := s.sessionRepo.Close(txCtx, session.ID); err != nil {
			return fmt.Errorf("failed to close session: %w", err)
		}

		// Update driver status
		if err := s.repo.UpdateStatus(txCtx, driverID, models.Offline); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		// Calculate session summary
		duration := time.Since(session.StartedAt).Hours()
		summary = &models.SessionSummary{
			SessionID:      session.ID,
			DurationHours:  duration,
			RidesCompleted: session.TotalRides,
			Earnings:       session.TotalEarnings,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Publish driver status update
	statusUpdate := map[string]interface{}{
		"driver_id": driverID,
		"status":    models.Offline.String(),
		"timestamp": time.Now(),
	}

	data, _ := json.Marshal(statusUpdate)
	routingKey := fmt.Sprintf("driver.status.%s", driverID)
	_ = s.publish.Publish("driver_topic", routingKey, data)

	return summary, nil
}

func (s *DriverService) UpdateLocation(ctx context.Context, driverID string, update *models.LocationUpdate) error {
	if driverID == "" {
		return errors.New("driverID cannot be empty")
	}

	if err := validateLatLon(update.Latitude, update.Longitude); err != nil {
		return err
	}

	// Verify driver is online
	driver, err := s.repo.GetById(ctx, driverID)
	if err != nil {
		return fmt.Errorf("failed to get driver: %w", err)
	}

	if driver.Status == models.Offline {
		return errors.New("cannot update location: driver offline")
	}

	// Update location in transaction
	var coordID string
	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Create/Update coordinate
		coordID, err = s.coordinateRepo.CreateOrUpdate(txCtx, driverID, "driver",
			update.Latitude, update.Longitude, update.Address)
		if err != nil {
			return fmt.Errorf("failed to update coordinate: %w", err)
		}

		// Add to location history
		historyLoc := &models.LocationHistory{
			CoordinateID:   coordID,
			DriverID:       driverID,
			Latitude:       update.Latitude,
			Longitude:      update.Longitude,
			AccuracyMeters: update.AccuracyMeters,
			SpeedKmh:       update.SpeedKmh,
			HeadingDegrees: update.HeadingDegrees,
			RideID:         update.RideID,
			RecordedAt:     time.Now(),
		}

		return s.locationRepo.AddLocation(txCtx, historyLoc)
	})
	if err != nil {
		return err
	}

	// Broadcast location update to fanout exchange
	locationMsg := map[string]interface{}{
		"driver_id": driverID,
		"ride_id":   update.RideID,
		"location": map[string]float64{
			"lat": update.Latitude,
			"lng": update.Longitude,
		},
		"speed_kmh":       update.SpeedKmh,
		"heading_degrees": update.HeadingDegrees,
		"timestamp":       time.Now(),
	}

	data, _ := json.Marshal(locationMsg)
	_ = s.publish.Publish("location_fanout", "", data)

	return nil
}

func (s *DriverService) StartRide(ctx context.Context, driverID, rideID string, lat, lon float64) error {
	if rideID == "" {
		return errors.New("rideID cannot be empty")
	}

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Verify driver status
		driver, err := s.repo.GetById(txCtx, driverID)
		if err != nil {
			return fmt.Errorf("failed to get driver: %w", err)
		}

		if driver.Status != models.Busy {
			return fmt.Errorf("cannot start ride: driver status is %s, must be BUSY", driver.Status)
		}

		// Update ride status to IN_PROGRESS
		if err := s.repo.UpdateRideStatus(txCtx, rideID, models.RideStatusInProgress); err != nil {
			return fmt.Errorf("failed to update ride status: %w", err)
		}

		// Record start location
		coordID, err := s.coordinateRepo.CreateOrUpdate(txCtx, driverID, "driver", lat, lon, "")
		if err != nil {
			return fmt.Errorf("failed to record start location: %w", err)
		}

		historyLoc := &models.LocationHistory{
			CoordinateID: coordID,
			DriverID:     driverID,
			Latitude:     lat,
			Longitude:    lon,
			RideID:       &rideID,
			RecordedAt:   time.Now(),
		}

		return s.locationRepo.AddLocation(txCtx, historyLoc)
	})
	if err != nil {
		return err
	}

	// Publish ride status update
	statusUpdate := map[string]interface{}{
		"ride_id":   rideID,
		"status":    models.RideStatusInProgress.String(),
		"driver_id": driverID,
		"timestamp": time.Now(),
	}

	data, _ := json.Marshal(statusUpdate)
	routingKey := fmt.Sprintf("ride.status.%s", models.RideStatusInProgress.String())
	_ = s.publish.Publish("ride_topic", routingKey, data)

	return nil
}

func (s *DriverService) CompleteRide(ctx context.Context, driverID, rideID string,
	finalLat, finalLon, actualDistance float64, actualDuration int,
) error {
	if rideID == "" {
		return errors.New("rideID cannot be empty")
	}

	var driverEarnings float64

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Update ride status to COMPLETED
		if err := s.repo.UpdateRideStatus(txCtx, rideID, models.RideStatusCompleted); err != nil {
			return fmt.Errorf("failed to update ride status: %w", err)
		}

		// Record final location
		coordID, err := s.coordinateRepo.CreateOrUpdate(txCtx, driverID, "driver",
			finalLat, finalLon, "")
		if err != nil {
			return fmt.Errorf("failed to record final location: %w", err)
		}

		historyLoc := &models.LocationHistory{
			CoordinateID: coordID,
			DriverID:     driverID,
			Latitude:     finalLat,
			Longitude:    finalLon,
			RideID:       &rideID,
			RecordedAt:   time.Now(),
		}

		if err := s.locationRepo.AddLocation(txCtx, historyLoc); err != nil {
			return err
		}

		// Update driver status back to AVAILABLE
		if err := s.repo.UpdateStatus(txCtx, driverID, models.Available); err != nil {
			return fmt.Errorf("failed to update driver status: %w", err)
		}

		// Get ride details for earnings calculation (80% to driver)
		ride, err := s.repo.GetRideByID(txCtx, rideID)
		if err != nil {
			return fmt.Errorf("failed to get ride: %w", err)
		}

		driverEarnings = ride.FinalFare * 0.80 // 80% to driver, 20% commission

		// Update driver totals
		driver, err := s.repo.GetById(txCtx, driverID)
		if err != nil {
			return fmt.Errorf("failed to get driver: %w", err)
		}

		driver.TotalRides++
		driver.TotalEarnings += driverEarnings

		if err := s.repo.Update(txCtx, driver); err != nil {
			return fmt.Errorf("failed to update driver totals: %w", err)
		}

		// Update session totals
		session, err := s.sessionRepo.GetActiveByDriverID(txCtx, driverID)
		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}

		session.TotalRides++
		session.TotalEarnings += driverEarnings

		return s.sessionRepo.Update(txCtx, session)
	})
	if err != nil {
		return err
	}

	// Publish ride completion
	statusUpdate := map[string]interface{}{
		"ride_id":         rideID,
		"status":          models.RideStatusCompleted.String(),
		"driver_id":       driverID,
		"driver_earnings": driverEarnings,
		"timestamp":       time.Now(),
	}

	data, _ := json.Marshal(statusUpdate)
	routingKey := fmt.Sprintf("ride.status.%s", models.RideStatusCompleted.String())
	_ = s.publish.Publish("ride_topic", routingKey, data)

	return nil
}

func validateLatLon(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if lon < -180 || lon > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}
