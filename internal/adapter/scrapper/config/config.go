package config

import (
	"errors"
	"os"
)

const (
	githubAPIKey        = "GITHUB_API_KEY"
	stackoverflowAPIKey = "STACKOVERFLOW_API_KEY"
	botServerAddress    = "BOT_SERVER_ADDRESS"
	assetType           = "ASSET_TYPE"
)

var (
	ErrNoTelegramToken      = errors.New("telegram token should be set up with " + githubAPIKey + " environment variable")
	ErrNoStackOverflowToken = errors.New("stackoverflow token should be set up with " + stackoverflowAPIKey + " environment variable")
	ErrNoBotServerAddress   = errors.New("bot server address should not be set with " + botServerAddress + " environment variable")
	ErrNoAssetType          = errors.New("asset type should not be set with " + assetType + " environment variable")
)

type AssetType int

const (
	AssetTypeSQL AssetType = iota
	AssetTypeORM
)

type Config struct {
	GithubToken        string
	StackoverflowToken string
	BotServerAddr      string
	PostgresURL        string
	AssetType          AssetType
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
	return Config{GithubToken: githubToken, StackoverflowToken: stackoverflowToken, BotServerAddr: botServAddr, AssetType: asset}, nil
}

func findAssetType(asset string) (AssetType, error) {
	switch asset {
	case "SQL":
		return AssetTypeSQL, nil
	case "ORM":
		return AssetTypeORM, nil
	default:
		return 0, ErrNoAssetType
	}
}
