package config

import (
	"errors"
	"os"
)

const (
	githubAPIKey        = "GITHUB_API_KEY"
	stackoverflowAPIKey = "STACKOVERFLOW_API_KEY"
	botServerAddress    = "BOT_SERVER_ADDRESS"
)

var (
	ErrNoTelegramToken      = errors.New("telegram token should be set up with " + githubAPIKey + " environment variable")
	ErrNoStackOverflowToken = errors.New("stackoverflow token should be set up with " + githubAPIKey + " environment variable")
	ErrNoBotServerAddress   = errors.New("bot server address should not be set with " + botServerAddress + " environment variable")
)

type Config struct {
	GithubToken        string
	StackoverflowToken string
	BotServerAddr      string
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
	return Config{GithubToken: githubToken, StackoverflowToken: stackoverflowToken, BotServerAddr: botServAddr}, nil
}
