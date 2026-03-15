package config

import (
	"errors"
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

			cfg, err := ParseConfig()

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}

			if cfg != tt.expectedCfg {
				t.Fatalf("expected config %+v, got %+v", tt.expectedCfg, cfg)
			}
		})
	}
}
