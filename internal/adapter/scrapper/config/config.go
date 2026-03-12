package config

import (
	"errors"
	"os"
)

const (
	githubApiKey        = "GITHUB_API_KEY"
	stackoverflowApiKey = "STACKOVERFLOW_API_KEY"
	botServerAddress    = "BOT_SERVER_ADDRESS"
)

var (
	NoTelegramTokenErr      = errors.New("telegram token should be set up with " + githubApiKey + " environment variable")
	NoStackOverflowTokenErr = errors.New("stackoverflow token should be set up with " + githubApiKey + " environment variable")
	NoBotServerAddressErr   = errors.New("bot server address should not be set with " + botServerAddress + " environment variable")
)

type Config struct {
	GithubToken        string
	StackoverflowToken string
	BotServerAddr      string
}

func ParseConfig() (Config, error) {
	githubToken := os.Getenv(githubApiKey)
	if githubToken == "" {
		return Config{}, NoTelegramTokenErr
	}
	stackoverflowToken := os.Getenv(stackoverflowApiKey)
	if stackoverflowToken == "" {
		return Config{}, NoStackOverflowTokenErr
	}
	botServAddr := os.Getenv(botServerAddress)
	if botServAddr == "" {
		return Config{}, NoBotServerAddressErr
	}
	return Config{GithubToken: githubToken, StackoverflowToken: stackoverflowToken, BotServerAddr: botServAddr}, nil
}
