package dto

type CreateRideRequest struct {
	PassengerID   string  `json:"passenger_id"`
	PickupLat     float64 `json:"pickup_latitude"`
	PickupLon     float64 `json:"pickup_longitude"`
	PickupAddress string  `json:"pickup_address"`
	DestLat       float64 `json:"destination_latitude"`
	DestLon       float64 `json:"destination_longitude"`
	DestAddress   string  `json:"destination_address"`
	RideType      string  `json:"ride_type"`
}
