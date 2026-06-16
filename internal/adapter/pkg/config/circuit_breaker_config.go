package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

var (
	ErrNoCircuitBreakerInterval          = errors.New("circuit breaker interval should be set with " + circuitBreakerInterval + " environment variable")
	ErrInvalidCircuitBreakerInterval     = errors.New(circuitBreakerInterval + " should be positive duration")
	ErrNoCircuitBreakerTimeout           = errors.New("circuit breaker timeout should be set with " + circuitBreakerTimeout + " environment variable")
	ErrInvalidCircuitBreakerTimeout      = errors.New(circuitBreakerTimeout + " should be positive duration")
	ErrNoCircuitBreakerMaxRequests       = errors.New("circuit breaker max requests should be set with " + circuitBreakerMaxRequests + " environment variable")
	ErrInvalidCircuitBreakerMaxRequests  = errors.New(circuitBreakerMaxRequests + " should be positive integer")
	ErrNoCircuitBreakerFailureRatio      = errors.New("circuit breaker failure ratio should be set with " + circuitBreakerFailureRatio + " environment variable")
	ErrInvalidCircuitBreakerFailureRatio = errors.New(circuitBreakerFailureRatio + " should be float between 0 and 1")
)

const (
	circuitBreakerInterval     = "CIRCUIT_BREAKER_INTERVAL"
	circuitBreakerTimeout      = "CIRCUIT_BREAKER_TIMEOUT"
	circuitBreakerMaxRequests  = "CIRCUIT_BREAKER_MAX_REQUESTS"
	circuitBreakerFailureRatio = "CIRCUIT_BREAKER_FAILURE_RATIO"
)

type CircuitBreakerConfig struct {
	Interval     time.Duration
	Timeout      time.Duration
	MaxRequests  uint32
	FailureRatio float64
}

func ParseCircuitBreakerConfig() (CircuitBreakerConfig, error) {
	intervalStr := os.Getenv(circuitBreakerInterval)
	if intervalStr == "" {
		return CircuitBreakerConfig{}, ErrNoCircuitBreakerInterval
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil || interval <= 0 {
		return CircuitBreakerConfig{}, ErrInvalidCircuitBreakerInterval
	}

	timeoutStr := os.Getenv(circuitBreakerTimeout)
	if timeoutStr == "" {
		return CircuitBreakerConfig{}, ErrNoCircuitBreakerTimeout
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil || timeout <= 0 {
		return CircuitBreakerConfig{}, ErrInvalidCircuitBreakerTimeout
	}

	maxRequestsStr := os.Getenv(circuitBreakerMaxRequests)
	if maxRequestsStr == "" {
		return CircuitBreakerConfig{}, ErrNoCircuitBreakerMaxRequests
	}

	maxRequests, err := strconv.ParseUint(maxRequestsStr, 10, 32)
	if err != nil || maxRequests == 0 {
		return CircuitBreakerConfig{}, ErrInvalidCircuitBreakerMaxRequests
	}

	failureRatioStr := os.Getenv(circuitBreakerFailureRatio)
	if failureRatioStr == "" {
		return CircuitBreakerConfig{}, ErrNoCircuitBreakerFailureRatio
	}

	failureRatio, err := strconv.ParseFloat(failureRatioStr, 64)
	if err != nil || failureRatio <= 0 || failureRatio > 1 {
		return CircuitBreakerConfig{}, ErrInvalidCircuitBreakerFailureRatio
	}

	return CircuitBreakerConfig{
		Interval:     interval,
		Timeout:      timeout,
		MaxRequests:  uint32(maxRequests),
		FailureRatio: failureRatio,
	}, nil
}
