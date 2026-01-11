package dto

import "ride-hail/internal/ride/domain/models"

type CreateRideRequest struct {
	PickupLat     float64            `json:"pickup_latitude"`
	PickupLon     float64            `json:"pickup_longitude"`
	PickupAddress string             `json:"pickup_address"`
	DestLat       float64            `json:"destination_latitude"`
	DestLon       float64            `json:"destination_longitude"`
	DestAddress   string             `json:"destination_address"`
	RideType      models.VehicleType `json:"ride_type"`
}

type CancelRideRequest struct {
	RideId string `json:"ride_id"`
	Reason string `json:"reason"`
}