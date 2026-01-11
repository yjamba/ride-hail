package models

import "time"

type Driver struct {
	ID     string
	UserID string

	LicenseNumber string
	VehicleType   string
	VehicleAttrs  VehicleAttributes

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
