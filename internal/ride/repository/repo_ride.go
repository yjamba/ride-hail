package repository

import (
	"context"
	"errors"
	"ride-hail/internal/ride/domain/models"
)

var ErrDBNoConnection = errors.New("DB no connection")

type RideRepo struct {
	db *DB
	// logger
}

func NewRideRepo(db *DB) *RideRepo {
	return &RideRepo{db: db}
}

func (r *RideRepo) CreateRide(ctx context.Context, ride *models.RideDB) error {
	return nil
}

func (r *RideRepo) ListByPassenger(ctx context.Context, passengerID string) error {
return nil
}

func (r *RideRepo) GetRide(ctx context.Context, id string) error {
	return nil
}

func (r *RideRepo) UpdateRide(ctx context.Context, ride models.RideDB) error {
return nil
}

func (r *RideRepo) CloseRide(ctx context.Context, id string) error {
	return nil
}

type DB struct { // заглушка
}
