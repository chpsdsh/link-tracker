package config

import (
	"errors"
	"testing"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name          string
		githubToken   string
		stackToken    string
		botServerAddr string
		expectedCfg   Config
		expectedErr   error
	}{
		{
			name:          "success",
			githubToken:   "github_token",
			stackToken:    "stack_token",
			botServerAddr: "http://localhost:8080",
			expectedCfg: Config{
				GithubToken:        "github_token",
				StackoverflowToken: "stack_token",
				BotServerAddr:      "http://localhost:8080",
			},
			expectedErr: nil,
		},
		{
			name:          "missing github token",
			githubToken:   "",
			stackToken:    "stack_token",
			botServerAddr: "http://localhost:8080",
			expectedCfg:   Config{},
			expectedErr:   ErrNoTelegramToken,
		},
		{
			name:          "missing stackoverflow token",
			githubToken:   "github_token",
			stackToken:    "",
			botServerAddr: "http://localhost:8080",
			expectedCfg:   Config{},
			expectedErr:   ErrNoStackOverflowToken,
		},
		{
			name:          "missing bot server address",
			githubToken:   "github_token",
			stackToken:    "stack_token",
			botServerAddr: "",
			expectedCfg:   Config{},
			expectedErr:   ErrNoBotServerAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(githubAPIKey, tt.githubToken)
			t.Setenv(stackoverflowAPIKey, tt.stackToken)
			t.Setenv(botServerAddress, tt.botServerAddr)

			cfg, err := ParseConfig()

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if cfg != tt.expectedCfg {
				t.Fatalf("expected config %+v, got %+v", tt.expectedCfg, cfg)
			}
		})
	}
}
