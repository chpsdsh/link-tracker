package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/logger"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/scrapperserver"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/handler"
)

const (
	scrapperServerPort = ":8081"
	shutdownDuration   = 10 * time.Second
)

func main() {
	baseLogger := logger.NewLogger(os.Stdout, logger.OutputFormatJson, slog.LevelInfo)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	repo := repository.NewLinkRepository()

	scrapperHandler := handler.LinksHandler{Repo: repo}

	scrapperServer := scrapperserver.ScrapperServer{BaseLogger: baseLogger, Handler: scrapperHandler}
	mux := http.NewServeMux()
	h := scrapperserver.HandlerFromMux(scrapperServer, mux)

	server := &http.Server{Handler: h, Addr: scrapperServerPort}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			baseLogger.Error("error starting scrapper http server", slog.String("error", err.Error()))
		}
	}()
	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		baseLogger.Error("error shutting down scrapper http server", slog.String("error", err.Error()))
	}
}
