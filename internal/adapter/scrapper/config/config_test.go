package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	config2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name string

		githubToken     string
		stackToken      string
		botServerAddr   string
		configAssetType string
		updateSendType  string

		interval   string
		batchSize  string
		numWorkers string
		valkeyTTL  string

		expectedCfg ScrapperConfig
		expectedErr error
	}{
		{
			name:            "success",
			githubToken:     "github_token",
			stackToken:      "stack_token",
			botServerAddr:   "http://localhost:8080",
			configAssetType: "SQL",
			interval:        "10",
			batchSize:       "100",
			numWorkers:      "4",
			updateSendType:  "http",
			valkeyTTL:       "5",
			expectedCfg: ScrapperConfig{
				GithubToken:        "github_token",
				StackoverflowToken: "stack_token",
				BotServerAddr:      "http://localhost:8080",
				AssetType:          AssetTypeSQL,
				ScrapperInterval:   10 * time.Second,
				BatchSize:          100,
				NumWorkers:         4,
				UpdatesSendType:    "http",
				KafkaConfig: KafkaConfig{
					Brokers:            []string{"localhost:9092"},
					NotificationsTopic: "topic",
					User:               "user",
					Password:           "pass",
				},
				PostgresConfig: config2.PostgresConfig{
					Host:     "localhost",
					Port:     "5432",
					User:     "user",
					Password: "pass",
					DBName:   "db",
				},
				ValkeyConfig: ValkeyConfig{
					Addresses: []string{"valkey-0:6379", "valkey-1:6379"},
					Password:  "pass",
					ValkeyTTL: 5 * time.Minute,
				},
				HTTPClientConfig: config2.HTTPClientConfig{
					Timeout: 10 * time.Second,
				},

				RetryConfig: config2.RetryConfig{
					MaxAttempts:       3,
					Delay:             500 * time.Millisecond,
					RetryableStatuses: []int{500, 502, 503, 504},
				},

				CircuitBreakerConfig: config2.CircuitBreakerConfig{
					Interval:     10 * time.Second,
					Timeout:      5 * time.Second,
					MaxRequests:  3,
					FailureRatio: 0.6,
				},

				RateLimitConfig: config2.RateLimitConfig{
					RPS:   5,
					Burst: 10,
				},
			},
			expectedErr: nil,
		},
		{
			name:            "missing scrapper interval",
			githubToken:     "github_token",
			stackToken:      "stack_token",
			botServerAddr:   "http://localhost:8080",
			configAssetType: "SQL",
			interval:        "",
			batchSize:       "100",
			numWorkers:      "4",
			expectedErr:     ErrNoScrapperTimeInterval,
		},
		{
			name:            "invalid scrapper interval",
			githubToken:     "github_token",
			stackToken:      "stack_token",
			botServerAddr:   "http://localhost:8080",
			configAssetType: "SQL",
			interval:        "abc",
			batchSize:       "100",
			numWorkers:      "4",
			expectedErr:     ErrInvalidScrapperTimeInterval,
		},
		{
			name:            "missing batch size",
			githubToken:     "github_token",
			stackToken:      "stack_token",
			botServerAddr:   "http://localhost:8080",
			configAssetType: "SQL",
			interval:        "10",
			batchSize:       "",
			numWorkers:      "4",
			expectedErr:     ErrNoLinksBatchSize,
		},
		{
			name:            "invalid batch size",
			githubToken:     "github_token",
			stackToken:      "stack_token",
			botServerAddr:   "http://localhost:8080",
			configAssetType: "SQL",
			interval:        "10",
			batchSize:       "abc",
			numWorkers:      "4",
			expectedErr:     ErrInvalidBachSize,
		},
		{
			name:            "missing num workers",
			githubToken:     "github_token",
			stackToken:      "stack_token",
			botServerAddr:   "http://localhost:8080",
			configAssetType: "SQL",
			interval:        "10",
			batchSize:       "100",
			numWorkers:      "",
			expectedErr:     ErrNoNumWorkers,
		},
		{
			name:            "invalid num workers",
			githubToken:     "github_token",
			stackToken:      "stack_token",
			botServerAddr:   "http://localhost:8080",
			configAssetType: "SQL",
			interval:        "10",
			batchSize:       "100",
			numWorkers:      "abc",
			expectedErr:     ErrInvalidNumWorkers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(githubAPIKey, tt.githubToken)
			t.Setenv(stackoverflowAPIKey, tt.stackToken)
			t.Setenv(botServerAddress, tt.botServerAddr)
			t.Setenv(assetType, tt.configAssetType)

			t.Setenv(scrapperTimeInterval, tt.interval)
			t.Setenv(linksBatchSize, tt.batchSize)
			t.Setenv(schedulerNumWorkers, tt.numWorkers)
			t.Setenv(updatesHandleType, tt.updateSendType)

			t.Setenv(kafkaUser, "user")
			t.Setenv(kafkaPassword, "pass")
			t.Setenv(kafkaTopic, "topic")
			t.Setenv(kafkaBrokers, "localhost:9092")

			t.Setenv("POSTGRES_HOST", "localhost")
			t.Setenv("POSTGRES_PORT", "5432")
			t.Setenv("POSTGRES_USER", "user")
			t.Setenv("POSTGRES_PASSWORD", "pass")
			t.Setenv("POSTGRES_DB", "db")
			t.Setenv(valkeyAddressesEnv, "valkey-0:6379,valkey-1:6379")
			t.Setenv(valkeyPasswordEnv, "pass")
			t.Setenv(valkeyTTLEnv, tt.valkeyTTL)
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

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
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
