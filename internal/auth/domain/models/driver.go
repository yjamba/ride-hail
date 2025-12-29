package models

type Driver struct {
	UserID string

	LicenseNumber string
	VehicleType   string
	VehicleAttrs  *VehicleAttrs
}

type VehicleAttrs struct {
	VehicleMake  string
	VehicleModel string
	VehicleColor string
	VehiclePlate string
	VehicleYear  int
}
