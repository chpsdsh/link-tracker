package config

import (
	"errors"
	"os"
	"strconv"
)

const (
	rateLimitRPS   = "RATE_LIMIT_RPS"
	rateLimitBurst = "RATE_LIMIT_BURST"
)

var (
	ErrNoRateLimitRPS        = errors.New("rate limit rps should be set with " + rateLimitRPS + " environment variable")
	ErrInvalidRateLimitRPS   = errors.New(rateLimitRPS + " should be positive float")
	ErrNoRateLimitBurst      = errors.New("rate limit burst should be set with " + rateLimitBurst + " environment variable")
	ErrInvalidRateLimitBurst = errors.New(rateLimitBurst + " should be positive integer")
)

type RateLimitConfig struct {
	RPS   float64
	Burst int
}

func ParseRateLimitConfig() (RateLimitConfig, error) {
	rpsStr := os.Getenv(rateLimitRPS)
	if rpsStr == "" {
		return RateLimitConfig{}, ErrNoRateLimitRPS
	}
	rps, err := strconv.ParseFloat(rpsStr, 64)
	if err != nil || rps <= 0 {
		return RateLimitConfig{}, ErrInvalidRateLimitRPS
	}
	burstStr := os.Getenv(rateLimitBurst)
	if burstStr == "" {
		return RateLimitConfig{}, ErrNoRateLimitBurst
	}
	burst, err := strconv.Atoi(burstStr)
	if err != nil || burst <= 0 {
		return RateLimitConfig{}, ErrInvalidRateLimitBurst
	}
	return RateLimitConfig{
		RPS:   rps,
		Burst: burst,
	}, nil

}
