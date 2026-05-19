package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	config2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/database/config"
)

const (
	githubAPIKey         = "GITHUB_API_KEY"
	stackoverflowAPIKey  = "STACKOVERFLOW_API_KEY"
	botServerAddress     = "BOT_SERVER_ADDRESS"
	assetType            = "ASSET_TYPE"
	scrapperTimeInterval = "SCRAPPER_TIME_INTERVAL"
	linksBatchSize       = "LINKS_BATCH_SIZE"
	schedulerNumWorkers  = "SCHEDULER_NUM_WORKERS"
	updatesSendType      = "UPDATES_SEND_TYPE"
)

var (
	ErrNoTelegramToken             = errors.New("telegram token should be set up with " + githubAPIKey + " environment variable")
	ErrNoStackOverflowToken        = errors.New("stackoverflow token should be set up with " + stackoverflowAPIKey + " environment variable")
	ErrNoBotServerAddress          = errors.New("bot server address should be set with " + botServerAddress + " environment variable")
	ErrNoAssetType                 = errors.New("asset type should be set with " + assetType + " environment variable")
	ErrNoScrapperTimeInterval      = errors.New("scrapper time interval should  be set with " + scrapperTimeInterval + " environment variable")
	ErrInvalidScrapperTimeInterval = errors.New(scrapperTimeInterval + "should be integer")
	ErrNoLinksBatchSize            = errors.New("scrapper links batch size should  be set with " + linksBatchSize + " environment variable")
	ErrInvalidBachSize             = errors.New(linksBatchSize + "should be integer")
	ErrNoNumWorkers                = errors.New("scrapper num workers should be set with " + schedulerNumWorkers + " environment variable")
	ErrInvalidNumWorkers           = errors.New(schedulerNumWorkers + "should be integer")
	ErrNoUpdatesSendType           = errors.New("updates send type should be set with " + updatesSendType + " environment variable")
)

type AssetType int

const (
	AssetTypeSQL AssetType = iota
	AssetTypeBuilder
)

type Config struct {
	GithubToken          string
	StackoverflowToken   string
	BotServerAddr        string
	PostgresURL          string
	AssetType            AssetType
	ScrapperInterval     time.Duration
	BatchSize            int
	NumWorkers           int
	UpdatesSendType      string
	KafkaConfig          KafkaConfig
	PostgresConfig       config2.PostgresConfig
	ValkeyConfig         ValkeyConfig
	HTTPClientConfig     HTTPClientConfig
	RetryConfig          RetryConfig
	CircuitBreakerConfig CircuitBreakerConfig
}

func ParseConfig() (Config, error) { //nolint:funlen // config parsing should be in ine method
	githubToken := os.Getenv(githubAPIKey)
	if githubToken == "" {
		return Config{}, ErrNoTelegramToken
	}

	stackoverflowToken := os.Getenv(stackoverflowAPIKey)
	if stackoverflowToken == "" {
		return Config{}, ErrNoStackOverflowToken
	}

	botServAddr := os.Getenv(botServerAddress)
	if botServAddr == "" {
		return Config{}, ErrNoBotServerAddress
	}

	asset, err := findAssetType(os.Getenv(assetType))
	if err != nil {
		return Config{}, ErrNoAssetType
	}

	scrapperIntervalStr := os.Getenv(scrapperTimeInterval)
	if scrapperIntervalStr == "" {
		return Config{}, ErrNoScrapperTimeInterval
	}
	scrapperInterval, err := strconv.Atoi(scrapperIntervalStr)
	if err != nil {
		return Config{}, ErrInvalidScrapperTimeInterval
	}

	batchSizeStr := os.Getenv(linksBatchSize)
	if batchSizeStr == "" {
		return Config{}, ErrNoLinksBatchSize
	}
	batchSize, err := strconv.Atoi(batchSizeStr)
	if err != nil {
		return Config{}, ErrInvalidBachSize
	}

	numWorkersStr := os.Getenv(schedulerNumWorkers)
	if numWorkersStr == "" {
		return Config{}, ErrNoNumWorkers
	}
	numWorkers, err := strconv.Atoi(numWorkersStr)
	if err != nil {
		return Config{}, ErrInvalidNumWorkers
	}

	updatesSendTypeStr := os.Getenv(updatesSendType)
	if updatesSendTypeStr == "" {
		return Config{}, ErrNoUpdatesSendType
	}

	kafkaConfig, err := ParseKafkaConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing kafka config: %w", err)
	}

	postgresConfig, err := config2.ParsePostgresConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing postgres config: %w", err)
	}

	valkeyConfig, err := ParseValkeyConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing valkey config: %w", err)
	}

	httpClientConfig, err := ParseHTTPClientConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing http client config: %w", err)
	}

	retryConfig, err := ParseRetryConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing retry config: %w", err)
	}

	circuitBreakerConfig, err := ParseCircuitBreakerConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parsing circuit breaker config: %w", err)
	}
	return Config{GithubToken: githubToken,
		StackoverflowToken:   stackoverflowToken,
		BotServerAddr:        botServAddr,
		AssetType:            asset,
		ScrapperInterval:     time.Duration(scrapperInterval) * time.Second,
		BatchSize:            batchSize,
		NumWorkers:           numWorkers,
		UpdatesSendType:      updatesSendTypeStr,
		KafkaConfig:          kafkaConfig,
		PostgresConfig:       postgresConfig,
		ValkeyConfig:         valkeyConfig,
		HTTPClientConfig:     httpClientConfig,
		RetryConfig:          retryConfig,
		CircuitBreakerConfig: circuitBreakerConfig,
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
