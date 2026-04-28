package config

import (
	"errors"
	"reflect"
	"testing"
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
			t.Setenv(updatesReceiverType, "kafka")

			t.Setenv(kafkaUser, "user")
			t.Setenv(kafkaPassword, "pass")
			t.Setenv(kafkaTopic, "topic")
			t.Setenv(kafkaDLQTopic, "dlq")
			t.Setenv(kafkaBrokers, "localhost:9092")
			t.Setenv(kafkaConsumerGroup, "group")
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
