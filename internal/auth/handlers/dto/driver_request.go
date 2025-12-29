package dto

type DriverRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`

	LicenseNumber string              `json:"license_number"`
	VehicleType   string              `json:"vehicle_type"`
	VehicleAttrs  VehicleAttrsRequest `json:"vehicle_attrs"`
}

type VehicleAttrsRequest struct {
	VehicleMake  string `json:"vehicle_make"`
	VehicleModel string `json:"vehicle_model"`
	VehicleColor string `json:"vehicle_color"`
	VehiclePlate string `json:"vehicle_plate"`
	VehicleYear  int    `json:"vehicle_year"`
}
