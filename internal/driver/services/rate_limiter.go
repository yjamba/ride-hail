package services

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu          sync.Mutex
	lastUpdate  map[string]time.Time
	minInterval time.Duration
}

func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		lastUpdate:  make(map[string]time.Time),
		minInterval: minInterval,
	}
}

func (rl *RateLimiter) Allow(driverID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	lastTime, exists := rl.lastUpdate[driverID]

	if !exists || now.Sub(lastTime) >= rl.minInterval {
		rl.lastUpdate[driverID] = now
		return true
	}

	return false
}

// Cleanup old entries periodically
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	for driverID, lastTime := range rl.lastUpdate {
		if lastTime.Before(cutoff) {
			delete(rl.lastUpdate, driverID)
		}
	}
}
