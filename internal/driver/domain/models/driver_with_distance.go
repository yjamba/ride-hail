package models

type DriverWithDistance struct {
	ID         string  `json:"id"`
	Email      string  `json:"email"`
	Rating     float64 `json:"rating"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	DistanceKm float64 `json:"distance_km"`
}
