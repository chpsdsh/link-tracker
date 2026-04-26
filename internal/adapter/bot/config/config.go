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
	updatesReceiveType    = "UPDATES_RECEIVE_TYPE"
)

var (
	ErrNoTelegramAPIFlag    = errors.New("telegram API flag should be set up with " + withTelegramAPI + " environment variable")
	ErrNoTelegramToken      = errors.New("telegram token should be set up with " + telegramAPIKey + " environment variable")
	ErrNoScrapperAddress    = errors.New("scrapper server address should be set up with " + scrapperServerAddress + " environment variable")
	ErrNoUpdatesReceiveType = errors.New("updates receive type should be set up with " + updatesReceiveType + " environment variable")
)

type Config struct {
	WithTelegramAPI       bool
	TelegramToken         string
	ScrapperServerAddress string
	UpdatesReceiveType    string
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

	updatesReceiveTypeStr := os.Getenv(updatesReceiveType)
	if updatesReceiveTypeStr == "" {
		return Config{}, ErrNoUpdatesReceiveType
	}
	return Config{TelegramToken: token, ScrapperServerAddress: scrapperAddr, WithTelegramAPI: withTgAPI}, nil
}
