package config

import (
	"errors"
	"os"
	"strconv"
)

const (
	telegramAPIKey        = "APP_TELEGRAM_TOKEN"
	scrapperServerAddress = "SCRAPPER_SERVER_ADDRESS"
	withTelegramAPI       = "WITH_TELEGRAM_API"
)

var (
	ErrNoTelegramAPIFlag = errors.New("telegram API flag should be set up with " + withTelegramAPI + " environment variable")
	ErrNoTelegramToken   = errors.New("telegram token should be set up with " + telegramAPIKey + " environment variable")
	ErrNoScrapperAddress = errors.New("scrapper server address should be set up with " + scrapperServerAddress + " environment variable")
)

type Config struct {
	WithTelegramAPI       bool
	TelegramToken         string
	ScrapperServerAddress string
}

func ParseConfig() (Config, error) {
	token := os.Getenv(telegramAPIKey)
	if token == "" {
		return Config{}, ErrNoTelegramToken
	}

	scrapperAddr := os.Getenv(scrapperServerAddress)
	if scrapperAddr == "" {
		return Config{}, ErrNoScrapperAddress
	}

	withTgAPIEnv := os.Getenv(withTelegramAPI)
	withTgAPI, err := strconv.ParseBool(withTgAPIEnv)
	if err != nil {
		return Config{}, ErrNoTelegramAPIFlag
	}
	return Config{TelegramToken: token, ScrapperServerAddress: scrapperAddr, WithTelegramAPI: withTgAPI}, nil
}
