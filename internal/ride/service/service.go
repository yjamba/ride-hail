package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ride-hail/internal/ride/domain/models"
	"ride-hail/internal/ride/domain/ports"
	"ride-hail/internal/shared/broker/messages"
	"ride-hail/internal/shared/logger"
	"time"
)

type RideService struct {
	repo      ports.RideRepository
	publisher ports.Publish
	logger    *logger.Logger
	secretKey []byte
}

func NewRideService(repo ports.RideRepository, publisher ports.Publish, log *logger.Logger, secretKey []byte) *RideService {
	return &RideService{
		repo:      repo,
		publisher: publisher,
		logger:    log,
		secretKey: secretKey,
	}
}

func (s *RideService) CreateRide(ctx context.Context, cmd models.CreateRideCommand) (*models.Ride, error) {
	// 1. Валидация координат
	if err := validateLanLon(cmd.Pickup.Latitude, cmd.Pickup.Longitude); err != nil {
		s.logError(ctx, "validation_error", "invalid pickup coordinates", err)
		return nil, err
	}
	if err := validateLanLon(cmd.Destination.Latitude, cmd.Destination.Longitude); err != nil {
		s.logError(ctx, "validation_error", "invalid destination coordinates", err)
		return nil, err
	}

	// 2. Расчёты
	distanceKm := calculateDistance(cmd.Pickup.Latitude, cmd.Pickup.Longitude,
		cmd.Destination.Latitude, cmd.Destination.Longitude)
	durationMin := estimateDuration(distanceKm)

	vehicleType := cmd.VehicleType
	if !vehicleType.IsValid() {
		vehicleType = models.VehicleTypeEconomy
	}

	pricing := models.PricingTable[vehicleType]
	estimatedFare := pricing.BaseFare + distanceKm*pricing.RatePerKm + float64(durationMin)*pricing.RatePerMin

	// 3. Формируем Ride
	ride := &models.Ride{
		PassengerID:              cmd.PassengerID,
		VehicleType:              vehicleType,
		Status:                   models.RideStatusRequested,
		PickupLocation:           cmd.Pickup,
		DestinationLocation:      cmd.Destination,
		RequestedAt:              time.Now(),
		EstimatedFare:            &estimatedFare,
		EstimatedDistanceKm:      distanceKm,
		EstimatedDurationMinutes: durationMin,
	}

	// 4. Генерация ride_number
	ride.RideNumber = fmt.Sprintf("RIDE-%d", time.Now().UnixNano())

	// 5. Сохраняем в репозитории (внутри транзакции repo создаст coordinates)
	if err := s.repo.CreateRide(ctx, ride); err != nil {
		s.logError(ctx, "db_error", "failed to create ride", err)
		return nil, err
	}

	// 6. Логирование
	ctx = logger.WithRideID(ctx, ride.ID)
	s.logInfo(ctx, "ride_created", "ride successfully created", map[string]any{
		"passenger_id":   ride.PassengerID,
		"ride_number":    ride.RideNumber,
		"vehicle_type":   ride.VehicleType,
		"estimated_fare": estimatedFare,
	})

	// 7. Публикуем событие в брокер (если есть)
	if err := s.publishRideMatchRequest(ctx, ride); err != nil {
		s.logError(ctx, "publish_error", "failed to publish ride match request", err)
	}

	return ride, nil
}

func (s *RideService) GetRideById(ctx context.Context, rideID, passengerID string) (models.Ride, error) {
	ride, err := s.repo.GetRide(ctx, rideID)
	if err != nil {
		return models.Ride{}, err
	}
	return ride, nil
}

func (s *RideService) GetRideByStatus(ctx context.Context, passengerID string, status string) ([]models.Ride, error) {
	rides, err := s.repo.ListByStatus(ctx, passengerID, status)
	if err != nil {
		return nil, err
	}
	return rides, nil
}

func (s *RideService) UpdateRideStatus(ctx context.Context, rideID string, status string) error {
	validStatuses := []string{"REQUESTED", "IN_PROGRESS", "COMPLETED", "CANCELLED"}
	if !contains(validStatuses, status) {
		return errors.New("invalid ride status")
	}
	if err := s.repo.UpdateStatus(ctx, rideID, status); err != nil {
		s.logError(ctx, "db_error", "Failed to update ride status", err)
		return err
	}

	ctx = logger.WithRideID(ctx, rideID)
	s.logInfo(ctx, "status_changed", "Ride status updated", map[string]interface{}{
		"new_status": status,
	})
	return nil
}

func (s *RideService) CloseRide(ctx context.Context, id string, reason string) error {
	if err := s.repo.CloseRide(ctx, id, reason); err != nil {
		return err
	}
	return nil
}

func validateLanLon(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}

	if lon < -180 || lon > 180 {
		return errors.New("longitude must be between -180 and 180")
	}

	return nil
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// calculateDistance calculates the distance in km between two coordinates using Haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := lat1 * (3.141592653589793 / 180.0)
	lat2Rad := lat2 * (3.141592653589793 / 180.0)
	deltaLat := (lat2 - lat1) * (3.141592653589793 / 180.0)
	deltaLon := (lon2 - lon1) * (3.141592653589793 / 180.0)

	a := sin(deltaLat/2)*sin(deltaLat/2) +
		cos(lat1Rad)*cos(lat2Rad)*sin(deltaLon/2)*sin(deltaLon/2)
	c := 2 * atan2(sqrt(a), sqrt(1-a))

	return earthRadiusKm * c
}

// estimateDuration estimates ride duration in minutes based on distance
// Assumes average speed of 30 km/h in city traffic
func estimateDuration(distanceKm float64) int {
	avgSpeedKmh := 30.0
	durationHours := distanceKm / avgSpeedKmh
	return int(durationHours * 60)
}

// Math helper functions to avoid math package import
func sin(x float64) float64 {
	// Taylor series approximation
	x = fmod(x, 2*3.141592653589793)
	sum := x
	term := x
	for i := 1; i < 10; i++ {
		term *= -x * x / float64((2*i)*(2*i+1))
		sum += term
	}
	return sum
}

func cos(x float64) float64 {
	return sin(x + 3.141592653589793/2)
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func atan2(y, x float64) float64 {
	if x > 0 {
		return atan(y / x)
	}
	if x < 0 && y >= 0 {
		return atan(y/x) + 3.141592653589793
	}
	if x < 0 && y < 0 {
		return atan(y/x) - 3.141592653589793
	}
	if x == 0 && y > 0 {
		return 3.141592653589793 / 2
	}
	if x == 0 && y < 0 {
		return -3.141592653589793 / 2
	}
	return 0
}

func atan(x float64) float64 {
	// Taylor series for small x
	if x > 1 {
		return 3.141592653589793/2 - atan(1/x)
	}
	if x < -1 {
		return -3.141592653589793/2 - atan(1/x)
	}
	sum := x
	term := x
	for i := 1; i < 20; i++ {
		term *= -x * x
		sum += term / float64(2*i+1)
	}
	return sum
}

func fmod(x, y float64) float64 {
	return x - float64(int(x/y))*y
}

// publishRideMatchRequest publishes a ride match request to the message broker
func (s *RideService) publishRideMatchRequest(ctx context.Context, ride *models.Ride) error {
	if s.publisher == nil {
		return nil
	}

	msg := messages.RideMatchRequest{
		RideID:     ride.ID,
		RideNumber: ride.RideNumber,
		PickupLocation: messages.Coordinate{
			Lat:     ride.PickupLocation.Latitude,
			Lng:     ride.PickupLocation.Longitude,
			Address: ride.PickupLocation.Address,
		},
		Destination: messages.Coordinate{
			Lat:     ride.DestinationLocation.Latitude,
			Lng:     ride.DestinationLocation.Longitude,
			Address: ride.DestinationLocation.Address,
		},
		RideType:       string(ride.VehicleType),
		EstimatedFare:  getEstimatedFare(ride.EstimatedFare),
		MaxDistanceKm:  10.0, // Default max distance for driver matching
		TimeoutSeconds: 60,   // Default timeout for driver response
		RequestedAt:    ride.RequestedAt,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	routingKey := messages.RideRequestRoutingKey(string(ride.VehicleType))
	return s.publisher.Publish(ctx, messages.ExchangeRideTopic, routingKey, body)
}

// publishRideStatusUpdate publishes a ride status update to the message broker
func (s *RideService) publishRideStatusUpdate(ctx context.Context, ride *models.Ride) error {
	if s.publisher == nil {
		return nil
	}

	msg := messages.RideStatusUpdate{
		RideID:    ride.ID,
		DriverID:  ride.DriverID,
		Status:    string(ride.Status),
		Timestamp: time.Now(),
	}

	if ride.Status == models.RideStatusCompleted && ride.FinalFare != nil {
		msg.FinalFare = ride.FinalFare
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	routingKey := messages.RideStatusRoutingKey(string(ride.Status))
	return s.publisher.Publish(ctx, messages.ExchangeRideTopic, routingKey, body)
}

func getEstimatedFare(fare *float64) float64 {
	if fare == nil {
		return 0
	}
	return *fare
}

// Logging helper methods
func (s *RideService) logInfo(ctx context.Context, action, message string, extra interface{}) {
	if s.logger != nil {
		s.logger.InfoWithFields(ctx, action, message, extra)
	}
}

func (s *RideService) logError(ctx context.Context, action, message string, err error) {
	if s.logger != nil {
		s.logger.Error(ctx, action, message, err)
	}
}

func (s *RideService) logDebug(ctx context.Context, action, message string) {
	if s.logger != nil {
		s.logger.Debug(ctx, action, message)
	}
}
