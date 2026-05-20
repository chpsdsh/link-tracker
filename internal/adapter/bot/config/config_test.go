package config

import (
	"errors"
	"reflect"
	"testing"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/database/config"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		scrapperAddr  string
		withAPIFlag   string
		expectedCfg   Config
		expectedError error
	}{
		{

			name:         "success",
			token:        "telegram_token",
			scrapperAddr: "http://localhost:8080",
			withAPIFlag:  "true",
			expectedCfg: Config{
				TelegramToken:         "telegram_token",
				ScrapperServerAddress: "http://localhost:8080",
				WithTelegramAPI:       true,
				UpdatesReceiveType:    "kafka",
				KafkaConfig: KafkaConfig{
					Brokers:            []string{"localhost:9092"},
					NotificationsTopic: "topic",
					DLQTopic:           "dlq",
					ConsumerGroup:      "group",
					User:               "user",
					Password:           "pass",
				},
				PostgresConfig: config.PostgresConfig{
					Host:     "POSTGRES_HOST",
					Port:     "POSTGRES_PORT",
					User:     "POSTGRES_USER",
					Password: "POSTGRES_PASSWORD",
					DBName:   "POSTGRES_DB",
				},
			},
			expectedError: nil,
		},
		{
			name:          "missing telegram token",
			token:         "",
			scrapperAddr:  "http://localhost:8080",
			withAPIFlag:   "true",
			expectedCfg:   Config{},
			expectedError: ErrNoTelegramToken,
		},
		{
			name:          "missing scrapper address",
			token:         "telegram_token",
			scrapperAddr:  "",
			expectedCfg:   Config{},
			expectedError: ErrNoScrapperAddress,
		},
		{
			name:          "missing scrapper address",
			token:         "telegram_token",
			scrapperAddr:  "http://localhost:8080",
			withAPIFlag:   "",
			expectedCfg:   Config{},
			expectedError: ErrNoTelegramAPIFlag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(telegramAPIKey, tt.token)
			t.Setenv(scrapperServerAddress, tt.scrapperAddr)
			t.Setenv(withTelegramAPI, tt.withAPIFlag)
			t.Setenv(updatesHandleType, "kafka")

			t.Setenv(kafkaUser, "user")
			t.Setenv(kafkaPassword, "pass")
			t.Setenv(kafkaTopic, "topic")
			t.Setenv(kafkaDLQTopic, "dlq")
			t.Setenv(kafkaBrokers, "localhost:9092")
			t.Setenv(kafkaConsumerGroup, "group")
			t.Setenv("POSTGRES_HOST", "POSTGRES_HOST")
			t.Setenv("POSTGRES_PORT", "POSTGRES_PORT")
			t.Setenv("POSTGRES_USER", "POSTGRES_USER")
			t.Setenv("POSTGRES_PASSWORD", "POSTGRES_PASSWORD")
			t.Setenv("POSTGRES_DB", "POSTGRES_DB")

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
