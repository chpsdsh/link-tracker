package config

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		scrapperAddr  string
		withAPIFlag   string
		expectedCfg   BotConfig
		expectedError error
	}{
		{
			name:         "success",
			token:        "telegram_token",
			scrapperAddr: "http://localhost:8080",
			withAPIFlag:  "true",
			expectedCfg: BotConfig{
				TelegramToken:         "telegram_token",
				ScrapperServerAddress: "http://localhost:8080",
				WithTelegramAPI:       true,
				UpdatesReceiveType:    "kafka",
				KafkaConfig: config.KafkaConfig{
					Brokers:               []string{"localhost:9092"},
					RawNotificationsTopic: "topic",
					DLQTopic:              "dlq",
					ConsumerGroup:         "group",
					User:                  "user",
					Password:              "pass",
				},
				PostgresConfig: config.PostgresConfig{
					Host:     "POSTGRES_HOST",
					Port:     "POSTGRES_PORT",
					User:     "POSTGRES_USER",
					Password: "POSTGRES_PASSWORD",
					DBName:   "POSTGRES_DB",
				},
				HTTPClientConfig: config.HTTPClientConfig{
					Timeout: 10 * time.Second,
				},

				RetryConfig: config.RetryConfig{
					MaxAttempts:       3,
					Delay:             500 * time.Millisecond,
					RetryableStatuses: []int{500, 502, 503, 504},
				},

				CircuitBreakerConfig: config.CircuitBreakerConfig{
					Interval:     10 * time.Second,
					Timeout:      5 * time.Second,
					MaxRequests:  3,
					FailureRatio: 0.6,
				},

				RateLimitConfig: config.RateLimitConfig{
					RPS:   5,
					Burst: 10,
				},
			},
			expectedError: nil,
		},
		{
			name:          "missing telegram token",
			token:         "",
			scrapperAddr:  "http://localhost:8080",
			withAPIFlag:   "true",
			expectedCfg:   BotConfig{},
			expectedError: ErrNoTelegramToken,
		},
		{
			name:          "missing scrapper address",
			token:         "telegram_token",
			scrapperAddr:  "",
			expectedCfg:   BotConfig{},
			expectedError: ErrNoScrapperAddress,
		},
		{
			name:          "missing scrapper address",
			token:         "telegram_token",
			scrapperAddr:  "http://localhost:8080",
			withAPIFlag:   "",
			expectedCfg:   BotConfig{},
			expectedError: ErrNoTelegramAPIFlag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(telegramAPIKey, tt.token)
			t.Setenv(scrapperServerAddress, tt.scrapperAddr)
			t.Setenv(withTelegramAPI, tt.withAPIFlag)
			t.Setenv(updatesHandleType, "kafka")

			t.Setenv("KAFKA_USER", "user")
			t.Setenv("KAFKA_PASSWORD", "pass")
			t.Setenv("KAFKA_TOPIC", "topic")
			t.Setenv("KAFKA_DLQ_TOPIC", "dlq")
			t.Setenv("KAFKA_BROKERS", "localhost:9092")
			t.Setenv("KAFKA_CONSUMER_GROUP", "group")
			t.Setenv("POSTGRES_HOST", "POSTGRES_HOST")
			t.Setenv("POSTGRES_PORT", "POSTGRES_PORT")
			t.Setenv("POSTGRES_USER", "POSTGRES_USER")
			t.Setenv("POSTGRES_PASSWORD", "POSTGRES_PASSWORD")
			t.Setenv("POSTGRES_DB", "POSTGRES_DB")
			t.Setenv("HTTP_CLIENT_TIMEOUT", "10s")
			t.Setenv("RETRY_MAX_ATTEMPTS", "3")
			t.Setenv("RETRY_DELAY", "500ms")
			t.Setenv("RETRYABLE_STATUSES", "500,502,503,504")
			t.Setenv("CIRCUIT_BREAKER_INTERVAL", "10s")
			t.Setenv("CIRCUIT_BREAKER_TIMEOUT", "5s")
			t.Setenv("CIRCUIT_BREAKER_MAX_REQUESTS", "3")
			t.Setenv("CIRCUIT_BREAKER_FAILURE_RATIO", "0.6")
			t.Setenv("RATE_LIMIT_RPS", "5")
			t.Setenv("RATE_LIMIT_BURST", "10")

			cfg, err := ParseConfig()

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}

			if !reflect.DeepEqual(cfg, tt.expectedCfg) {
				t.Fatalf("expected config %+v, got %+v", tt.expectedCfg, cfg)
			}
		})
	}
}
