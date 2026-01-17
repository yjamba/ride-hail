package models

import "time"

// PickupCoordinateID, DestinationCoordinateID — это инфраструктура
// EstimatedFare — бизнес
// timestamps — бизнес
type Ride struct {
	ID          string      `json:"id"`
	RideNumber  string      `json:"ride_number"`
	PassengerID string      `json:"passenger_id"`
	DriverID    string      `json:"driver_id,omitempty"`
	VehicleType VehicleType `json:"vehicle_type"`
	Status      RideStatus  `json:"status"`
	Priority    int         `json:"priority"`

	// Locations
	PickupLocation          Location `json:"pickup_location"`
	DestinationLocation     Location `json:"destination_location"`
	PickupCoordinateID      string   `json:"pickup_coordinate_id,omitempty"`
	DestinationCoordinateID string   `json:"destination_coordinate_id,omitempty"`

	// Timestamps
	RequestedAt        time.Time  `json:"requested_at"`
	MatchedAt          *time.Time `json:"matched_at,omitempty"`
	ArrivedAt          *time.Time `json:"arrived_at,omitempty"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason string     `json:"cancellation_reason,omitempty"`

	// Financial
	EstimatedFare *float64 `json:"estimated_fare,omitempty"`
	FinalFare     *float64 `json:"final_fare,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VehicleType - тип транспортного средства
type VehicleType string

const (
	VehicleTypeEconomy VehicleType = "ECONOMY"
	VehicleTypePremium VehicleType = "PREMIUM"
	VehicleTypeXL      VehicleType = "XL"
)

// Location - модель геолокации
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

// Coordinates - модель координат для хранения в БД
type Coordinates struct {
	ID              string    `json:"id"`
	EntityID        string    `json:"entity_id"`
	EntityType      string    `json:"entity_type"`
	Location        Location  `json:"location"`
	FareAmount      *float64  `json:"fare_amount,omitempty"`
	DistanceKm      *float64  `json:"distance_km,omitempty"`
	DurationMinutes *int      `json:"duration_minutes,omitempty"`
	IsCurrent       bool      `json:"is_current"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// RideStatus - статусы поездки
type RideStatus string

const (
	RideStatusRequested  RideStatus = "REQUESTED"
	RideStatusMatched    RideStatus = "MATCHED"
	RideStatusEnRoute    RideStatus = "EN_ROUTE"
	RideStatusArrived    RideStatus = "ARRIVED"
	RideStatusInProgress RideStatus = "IN_PROGRESS"
	RideStatusCompleted  RideStatus = "COMPLETED"
	RideStatusCancelled  RideStatus = "CANCELLED"
)

// PricingInfo - информация о тарифе
type PricingInfo struct {
	BaseFare   float64 `json:"base_fare"`
	RatePerKm  float64 `json:"rate_per_km"`
	RatePerMin float64 `json:"rate_per_min"`
}

// PricingTable - таблица тарифов
var PricingTable = map[VehicleType]PricingInfo{
	VehicleTypeEconomy: {BaseFare: 500, RatePerKm: 100, RatePerMin: 50},
	VehicleTypePremium: {BaseFare: 800, RatePerKm: 120, RatePerMin: 60},
	VehicleTypeXL:      {BaseFare: 1000, RatePerKm: 150, RatePerMin: 75},
}

type CreateRideCommand struct {
	PassengerID string
	VehicleType VehicleType
	Pickup      Location
	Destination Location
}

func (v VehicleType) IsValid() bool {
	switch v {
	case VehicleTypeEconomy,
		VehicleTypePremium,
		VehicleTypeXL:
		return true
	default:
		return false
	}
}
