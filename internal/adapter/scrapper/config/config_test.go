package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name            string
		githubToken     string
		stackToken      string
		botServerAddr   string
		configAssetType string

		expectedCfg Config
		expectedErr error
	}{
		{
			name:            "success",
			githubToken:     "github_token",
			stackToken:      "stack_token",
			botServerAddr:   "http://localhost:8080",
			configAssetType: "SQL",
			expectedCfg: Config{
				GithubToken:        "github_token",
				StackoverflowToken: "stack_token",
				BotServerAddr:      "http://localhost:8080",
			},
			expectedErr: nil,
		},
		{
			name:          "no asset type",
			githubToken:   "github_token",
			stackToken:    "stack_token",
			botServerAddr: "http://localhost:8080",
			expectedCfg:   Config{},
			expectedErr:   ErrNoAssetType,
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
			t.Setenv(assetType, tt.configAssetType)

			cfg, err := ParseConfig()

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCfg, cfg)
		})
	}
}

func TestFindAssetType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    AssetType
		expectedErr error
	}{
		{
			name:        "SQL asset",
			input:       "SQL",
			expected:    AssetTypeSQL,
			expectedErr: nil,
		},
		{
			name:        "BUILDER asset",
			input:       "BUILDER",
			expected:    AssetTypeBuilder,
			expectedErr: nil,
		},
		{
			name:        "unknown asset",
			input:       "UNKNOWN",
			expected:    0,
			expectedErr: ErrNoAssetType,
		},
		{
			name:        "empty asset",
			input:       "",
			expected:    0,
			expectedErr: ErrNoAssetType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := findAssetType(tt.input)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}
