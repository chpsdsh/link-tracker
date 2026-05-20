package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
)

const (
	telegramAPIKey        = "APP_TELEGRAM_TOKEN"
	scrapperServerAddress = "SCRAPPER_SERVER_ADDRESS"
	withTelegramAPI       = "WITH_TELEGRAM_API"
	updatesHandleType     = "UPDATES_HANDLE_TYPE"
)

var (
	ErrNoTelegramAPIFlag    = errors.New("telegram API flag should be set up with " + withTelegramAPI + " environment variable")
	ErrNoTelegramToken      = errors.New("telegram token should be set up with " + telegramAPIKey + " environment variable")
	ErrNoScrapperAddress    = errors.New("scrapper server address should be set up with " + scrapperServerAddress + " environment variable")
	ErrNoUpdatesReceiveType = errors.New("updates handle type should be set up with " + updatesHandleType + " environment variable")
)

type Config struct {
	WithTelegramAPI       bool
	TelegramToken         string
	ScrapperServerAddress string
	UpdatesReceiveType    string
	KafkaConfig           KafkaConfig
	PostgresConfig        config.PostgresConfig
	HTTPClientConfig      config.HTTPClientConfig
	CircuitBreakerConfig  config.CircuitBreakerConfig
	RetryConfig           config.RetryConfig
	RateLimitConfig       config.RateLimitConfig
}

func ParseConfig() (Config, error) {
	token := os.Getenv(telegramAPIKey)
	if token == "" {
		return Config{}, ErrNoTelegramToken
	}

	scrapperAddr := os.Getenv(scrapperServerAddress)
	if scrapperAddr == "" {
		return Config{}, ErrNoScrapperAddress
	}

	withTgAPIEnv := os.Getenv(withTelegramAPI)
	withTgAPI, err := strconv.ParseBool(withTgAPIEnv)
	if err != nil {
		return Config{}, ErrNoTelegramAPIFlag
	}

	updatesReceiveTypeStr := os.Getenv(updatesHandleType)
	if updatesReceiveTypeStr == "" {
		return Config{}, ErrNoUpdatesReceiveType
	}

	kafkaConf, err := ParseKafkaConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing kafka config: %w", err)
	}

	postgresConf, err := config.ParsePostgresConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing postgres config: %w", err)
	}

	timeoutConf, err := config.ParseHTTPClientConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing http client config: %w", err)
	}

	circuitBreakerConf, err := config.ParseCircuitBreakerConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing circuit breaker config: %w", err)
	}
	retryConf, err := config.ParseRetryConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing retry config: %w", err)
	}

	rateLimitConf, err := config.ParseRateLimitConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing rate limit config: %w", err)
	}

	return Config{TelegramToken: token,
		ScrapperServerAddress: scrapperAddr,
		WithTelegramAPI:       withTgAPI,
		KafkaConfig:           kafkaConf,
		UpdatesReceiveType:    updatesReceiveTypeStr,
		PostgresConfig:        postgresConf,
		HTTPClientConfig:      timeoutConf,
		CircuitBreakerConfig:  circuitBreakerConf,
		RetryConfig:           retryConf,
		RateLimitConfig:       rateLimitConf,
	}, nil
}
