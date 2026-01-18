package service

import "errors"

var (
	ErrInvalidDistance = errors.New("invalid distance")
	ErrInvalidDuration = errors.New("invalid duration")
)

type FareCalculator struct {
	BaseFare       float64
	PricePerKM     float64
	PricePerMinute float64
	MinFare        float64
}

func NewFareCalculator(
	baseFare float64,
	pricePerKM float64,
	pricePerMinute float64,
	minFare float64,
) *FareCalculator {
	return &FareCalculator{
		BaseFare:       baseFare,
		PricePerKM:     pricePerKM,
		PricePerMinute: pricePerMinute,
		MinFare:        minFare,
	}
}

func (f *FareCalculator) Calculate(distanceKM float64, durationMin float64) (float64, error) {
	if distanceKM < 0 {
		return 0, ErrInvalidDistance
	}

	if durationMin < 0 {
		return 0, ErrInvalidDuration
	}

	total := f.BaseFare +
		distanceKM*f.PricePerKM +
		durationMin*f.PricePerMinute

	if total < f.MinFare {
		return f.MinFare, nil
	}

	return total, nil
}
