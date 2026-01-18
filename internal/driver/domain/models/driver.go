package models

import "time"

type Driver struct {
	ID     string
	UserID string

	LicenseNumber string
	VehicleType   string
	VehicleAttrs  VehicleAttributes
	Rating        float64
	TotalRides    int
	TotalEarnings float64

	Status    DriverStatus
	IsVerifed bool
}

type VehicleAttributes struct {
	VehicleMake  string
	VehicleModel string
	VehicleColor string
	VehiclePlate string
	VehicleYear  int
}

type DriverSession struct {
	ID       string
	DriverID string

	StartedAt time.Time
	EndedAt   time.Time

	TotalRides    int
	TotalEarnings float64
}

type LocationHistory struct {
	ID             string
	CoordinateID   string
	DriverID       string
	Latitude       float64
	Longitude      float64
	AccuracyMeters *float64
	SpeedKmh       *float64
	HeadingDegrees *float64
	RecordedAt     time.Time
	RideID         *string
}

type Coordinate struct {
	ID        string
	Latitude  float64
	Longitude float64
	Address   string
	CreatedAt time.Time
}

type LocationUpdate struct {
	Latitude       float64
	Longitude      float64
	Address        string
	AccuracyMeters *float64
	SpeedKmh       *float64
	HeadingDegrees *float64
	RideID         *string
}

type SessionSummary struct {
	SessionID      string
	DurationHours  float64
	RidesCompleted int
	Earnings       float64
}
