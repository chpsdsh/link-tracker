package config

import (
	"errors"
	"os"
)

const (
	githubApiKey        = "GITHUB_API_KEY"
	stackoverflowApiKey = "STACKOVERFLOW_API_KEY"
)

type Config struct {
	GithubToken        string
	StackoverflowToken string
}

func ParseConfig() (Config, error) {
	githubToken := os.Getenv(githubApiKey)
	if githubToken == "" {
		return Config{}, errors.New("telegram token should be set up with " + githubToken + " environment variable")
	}
	stackoverflowToken := os.Getenv(stackoverflowApiKey)
	if stackoverflowToken == "" {
		return Config{}, errors.New("telegram token should be set up with " + stackoverflowApiKey + " environment variable")
	}
	return Config{GithubToken: githubToken, StackoverflowToken: stackoverflowToken}, nil
}
