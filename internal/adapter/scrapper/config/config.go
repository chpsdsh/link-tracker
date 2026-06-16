package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
)

const (
	githubAPIKey             = "GITHUB_API_KEY"
	stackoverflowAPIKey      = "STACKOVERFLOW_API_KEY"
	botServerAddress         = "BOT_SERVER_ADDRESS"
	assetType                = "ASSET_TYPE"
	scrapperTimeInterval     = "SCRAPPER_TIME_INTERVAL"
	linksBatchSize           = "LINKS_BATCH_SIZE"
	schedulerNumWorkers      = "SCHEDULER_NUM_WORKERS"
	updatesHandleType        = "UPDATES_HANDLE_TYPE"
	metricsCalculateInterval = "METRICS_CALCULATE_INTERVAL"
)

var (
	ErrNoTelegramToken                     = errors.New("telegram token should be set up with " + githubAPIKey + " environment variable")
	ErrNoStackOverflowToken                = errors.New("stackoverflow token should be set up with " + stackoverflowAPIKey + " environment variable")
	ErrNoBotServerAddress                  = errors.New("bot server address should be set with " + botServerAddress + " environment variable")
	ErrNoAssetType                         = errors.New("asset type should be set with " + assetType + " environment variable")
	ErrNoScrapperTimeInterval              = errors.New("scrapper time interval should  be set with " + scrapperTimeInterval + " environment variable")
	ErrInvalidScrapperTimeInterval         = errors.New(scrapperTimeInterval + "should be integer")
	ErrNoLinksBatchSize                    = errors.New("scrapper links batch size should  be set with " + linksBatchSize + " environment variable")
	ErrInvalidBachSize                     = errors.New(linksBatchSize + "should be integer")
	ErrNoNumWorkers                        = errors.New("scrapper num workers should be set with " + schedulerNumWorkers + " environment variable")
	ErrInvalidNumWorkers                   = errors.New(schedulerNumWorkers + "should be integer")
	ErrNoUpdatesSendType                   = errors.New("updates send type should be set with " + updatesHandleType + " environment variable")
	ErrNoMetricsCalculateInterval          = errors.New("metrics calculate interval should be set with " + metricsCalculateInterval + " environment variable")
	ErrInvalidMetricsCalculateIntervalType = errors.New(metricsCalculateInterval + " should have time format")
)

type AssetType int

const (
	AssetTypeSQL AssetType = iota
	AssetTypeBuilder
)

type ScrapperConfig struct {
	GithubToken              string
	StackoverflowToken       string
	BotServerAddr            string
	PostgresURL              string
	AssetType                AssetType
	ScrapperInterval         time.Duration
	BatchSize                int
	NumWorkers               int
	UpdatesSendType          string
	KafkaConfig              KafkaConfig
	PostgresConfig           config.PostgresConfig
	ValkeyConfig             ValkeyConfig
	HTTPClientConfig         config.HTTPClientConfig
	RetryConfig              config.RetryConfig
	CircuitBreakerConfig     config.CircuitBreakerConfig
	RateLimitConfig          config.RateLimitConfig
	MetricsCalculateInterval time.Duration
}

func ParseConfig() (ScrapperConfig, error) { //nolint:funlen // config parsing should be in ine method
	githubToken := os.Getenv(githubAPIKey)
	if githubToken == "" {
		return ScrapperConfig{}, ErrNoTelegramToken
	}

	stackoverflowToken := os.Getenv(stackoverflowAPIKey)
	if stackoverflowToken == "" {
		return ScrapperConfig{}, ErrNoStackOverflowToken
	}

	botServAddr := os.Getenv(botServerAddress)
	if botServAddr == "" {
		return ScrapperConfig{}, ErrNoBotServerAddress
	}

	asset, err := findAssetType(os.Getenv(assetType))
	if err != nil {
		return ScrapperConfig{}, ErrNoAssetType
	}

	scrapperIntervalStr := os.Getenv(scrapperTimeInterval)
	if scrapperIntervalStr == "" {
		return ScrapperConfig{}, ErrNoScrapperTimeInterval
	}
	scrapperInterval, err := strconv.Atoi(scrapperIntervalStr)
	if err != nil {
		return ScrapperConfig{}, ErrInvalidScrapperTimeInterval
	}

	batchSizeStr := os.Getenv(linksBatchSize)
	if batchSizeStr == "" {
		return ScrapperConfig{}, ErrNoLinksBatchSize
	}
	batchSize, err := strconv.Atoi(batchSizeStr)
	if err != nil {
		return ScrapperConfig{}, ErrInvalidBachSize
	}

	numWorkersStr := os.Getenv(schedulerNumWorkers)
	if numWorkersStr == "" {
		return ScrapperConfig{}, ErrNoNumWorkers
	}
	numWorkers, err := strconv.Atoi(numWorkersStr)
	if err != nil {
		return ScrapperConfig{}, ErrInvalidNumWorkers
	}

	updatesSendTypeStr := os.Getenv(updatesHandleType)
	if updatesSendTypeStr == "" {
		return ScrapperConfig{}, ErrNoUpdatesSendType
	}

	metricsCalculateIntervalStr := os.Getenv(metricsCalculateInterval)
	if metricsCalculateIntervalStr == "" {
		return ScrapperConfig{}, ErrNoMetricsCalculateInterval
	}

	metricsCalculateTimeInterval, err := time.ParseDuration(metricsCalculateIntervalStr)
	if err != nil {
		return ScrapperConfig{}, ErrInvalidMetricsCalculateIntervalType
	}

	kafkaConfig, err := ParseKafkaConfig()
	if err != nil {
		return ScrapperConfig{}, fmt.Errorf("parsing kafka config: %w", err)
	}

	postgresConfig, err := config.ParsePostgresConfig()
	if err != nil {
		return ScrapperConfig{}, fmt.Errorf("parsing postgres config: %w", err)
	}

	valkeyConfig, err := ParseValkeyConfig()
	if err != nil {
		return ScrapperConfig{}, fmt.Errorf("parsing valkey config: %w", err)
	}

	httpClientConfig, err := config.ParseHTTPClientConfig()
	if err != nil {
		return ScrapperConfig{}, fmt.Errorf("parsing http client config: %w", err)
	}

	retryConfig, err := config.ParseRetryConfig()
	if err != nil {
		return ScrapperConfig{}, fmt.Errorf("parsing retry config: %w", err)
	}

	circuitBreakerConfig, err := config.ParseCircuitBreakerConfig()
	if err != nil {
		return ScrapperConfig{}, fmt.Errorf("parsing circuit breaker config: %w", err)
	}

	rateLimitConfig, err := config.ParseRateLimitConfig()
	if err != nil {
		return ScrapperConfig{}, fmt.Errorf("parsing rate limit config: %w", err)
	}
	return ScrapperConfig{GithubToken: githubToken,
		StackoverflowToken:       stackoverflowToken,
		BotServerAddr:            botServAddr,
		AssetType:                asset,
		ScrapperInterval:         time.Duration(scrapperInterval) * time.Second,
		BatchSize:                batchSize,
		NumWorkers:               numWorkers,
		UpdatesSendType:          updatesSendTypeStr,
		KafkaConfig:              kafkaConfig,
		PostgresConfig:           postgresConfig,
		ValkeyConfig:             valkeyConfig,
		HTTPClientConfig:         httpClientConfig,
		RetryConfig:              retryConfig,
		CircuitBreakerConfig:     circuitBreakerConfig,
		RateLimitConfig:          rateLimitConfig,
		MetricsCalculateInterval: metricsCalculateTimeInterval,
	}, nil

}

func findAssetType(asset string) (AssetType, error) {
	switch asset {
	case "SQL":
		return AssetTypeSQL, nil
	case "BUILDER":
		return AssetTypeBuilder, nil
	default:
		return 0, ErrNoAssetType
	}
}
