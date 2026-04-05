package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

const (
	githubAPIKey         = "GITHUB_API_KEY"
	stackoverflowAPIKey  = "STACKOVERFLOW_API_KEY"
	botServerAddress     = "BOT_SERVER_ADDRESS"
	assetType            = "ASSET_TYPE"
	scrapperTimeInterval = "SCRAPPER_TIME_INTERVAL"
	linksBatchSize       = "LINKS_BATCH_SIZE"
	schedulerNumWorkers  = "SCHEDULER_NUM_WORKERS"
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
)

type AssetType int

const (
	AssetTypeSQL AssetType = iota
	AssetTypeBuilder
)

type Config struct {
	GithubToken        string
	StackoverflowToken string
	BotServerAddr      string
	PostgresURL        string
	AssetType          AssetType
	ScrapperInterval   time.Duration
	BatchSize          int
	NumWorkers         int
}

func ParseConfig() (Config, error) {
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

	return Config{GithubToken: githubToken,
		StackoverflowToken: stackoverflowToken,
		BotServerAddr:      botServAddr,
		AssetType:          asset,
		ScrapperInterval:   time.Duration(scrapperInterval) * time.Second,
		BatchSize:          batchSize,
		NumWorkers:         numWorkers,
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
