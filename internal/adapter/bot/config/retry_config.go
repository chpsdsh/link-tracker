package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNoRetryMaxAttempts       = errors.New("retry max attempts should be set with " + retryMaxAttempts + " environment variable")
	ErrInvalidRetryMaxAttempts  = errors.New(retryMaxAttempts + " should be integer")
	ErrNoRetryDelay             = errors.New("retry delay should be set with " + retryDelay + " environment variable")
	ErrInvalidRetryDelay        = errors.New(retryDelay + " should be duration")
	ErrNoRetryableStatuses      = errors.New("retryable statuses should be set with " + retryableStatuses + " environment variable")
	ErrInvalidRetryableStatuses = errors.New(retryableStatuses + " should be comma-separated integers")
)

const (
	retryMaxAttempts  = "RETRY_MAX_ATTEMPTS"
	retryDelay        = "RETRY_DELAY"
	retryableStatuses = "RETRYABLE_STATUSES"
)

type RetryConfig struct {
	MaxAttempts       uint
	Delay             time.Duration
	RetryableStatuses []int
}

func ParseRetryConfig() (RetryConfig, error) {
	maxAttemptsStr := os.Getenv(retryMaxAttempts)
	if maxAttemptsStr == "" {
		return RetryConfig{}, ErrNoRetryMaxAttempts
	}

	maxAttempts, err := strconv.Atoi(maxAttemptsStr)
	if err != nil || maxAttempts <= 0 {
		return RetryConfig{}, ErrInvalidRetryMaxAttempts
	}

	delayStr := os.Getenv(retryDelay)
	if delayStr == "" {
		return RetryConfig{}, ErrNoRetryDelay
	}

	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		return RetryConfig{}, ErrInvalidRetryDelay
	}

	statusesStr := os.Getenv(retryableStatuses)
	if statusesStr == "" {
		return RetryConfig{}, ErrNoRetryableStatuses
	}

	statuses, err := parseRetryableStatuses(statusesStr)
	if err != nil {
		return RetryConfig{}, fmt.Errorf("%w: %w", ErrInvalidRetryableStatuses, err)
	}

	return RetryConfig{
		MaxAttempts:       uint(maxAttempts),
		Delay:             delay,
		RetryableStatuses: statuses,
	}, nil
}

func parseRetryableStatuses(statusesStr string) ([]int, error) {
	parts := strings.Split(statusesStr, ",")
	statuses := make([]int, 0, len(parts))

	for _, part := range parts {
		statusStr := strings.TrimSpace(part)
		if statusStr == "" {
			return nil, ErrInvalidRetryableStatuses
		}

		status, err := strconv.Atoi(statusStr)
		if err != nil {
			return nil, fmt.Errorf("parse status %q: %w", statusStr, err)
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}
