package dto

type RideResponse struct {
	RideId            string  `json:"ride_id"`
	RideNumber        string  `json:"ride_number"`
	Status            string  `json:"status"`
	EstimatedFare     float64 `json:"estimated_fare"`
	EstimatedDuration int64   `json:"estimated_duration_minutes"`
	EstimatedDistance float64 `json:"estimated_distance_km"`
}
