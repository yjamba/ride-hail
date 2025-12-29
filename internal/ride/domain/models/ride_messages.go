package models

type RideMessage struct {
	RideId         string         `json:"ride_id"`
	RideNumber     string         `json:"ride_number"`
	PickupLocation []PickupLatLon `json:"pickup_location"`
	DestLocation   []DestLatLon   `json:"destination_location"`
	RideType       string         `json:"ride_type"`
	EstimatedFare  float64        `json:"estimated_fare"`
	MaxDestKm      float64        `json:"max_distance_km"`
	TimeoutSeconds int64          `json:"timeout_seconds"`
	CorrelationId  string         `json:"correlation_id"`
}
type PickupLatLon struct {
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Address string  `json:"address"`
}
type DestLatLon struct {
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Address string  `json:"address"`
}
