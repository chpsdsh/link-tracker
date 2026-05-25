package main

import (
	"log/slog"
	"os"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/logger"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper"
)

func main() {
	baseLogger := logger.NewLogger(os.Stdout, logger.OutputFormatJSON, slog.LevelInfo)

	if err := scrapper.StartScrapper(baseLogger); err != nil {
		baseLogger.Error("error starting scrapper", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
