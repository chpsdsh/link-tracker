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

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/database"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/logger"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/scrapperclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/scrapperserver"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/scheduler"
)

const (
	envFilename        = "scrapper.env"
	scrapperServerPort = ":8081"
	shutdownDuration   = 10 * time.Second
	clientTimeout      = 15 * time.Second
)

func main() {
	baseLogger := logger.NewLogger(os.Stdout, logger.OutputFormatJSON, slog.LevelInfo)
	if err := godotenv.Load(envFilename); err != nil {
		baseLogger.Error("error loading file", slog.String("file", envFilename), slog.String("err", err.Error()))
		os.Exit(1)
	}

	conf, err := config.ParseConfig()
	if err != nil {
		baseLogger.Error("error parsing environment variables", slog.String("err", err.Error()))
		os.Exit(1)
	}

	dbConf, err := config.ParsePostgresConfig()
	if err != nil {
		baseLogger.Error("error parsing postgres configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	client := &http.Client{Timeout: clientTimeout}

	sched, err := gocron.NewScheduler()
	if err != nil {
		baseLogger.Error("error parsing environment variables", slog.String("err", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := database.NewDB(dbConf)
	if err != nil {
		baseLogger.Error("error connecting to database", slog.String("err", err.Error()))
		os.Exit(1)
	}

	repo := repository.NewLinkRepository()

	linksScheduler := scheduler.LinksRequester{
		Client:     scrapperclient.Client{Client: client, Config: conf},
		Scheduler:  sched,
		Repo:       repo,
		BaseLogger: baseLogger,
	}

	linksScheduler.StartLinkRequester()

	scrapperHandler := handler.LinksHandler{Repo: repo,
		BaseLogger: baseLogger,
	}

	serverImplementation := scrapperserver.ScrapperServer{BaseLogger: baseLogger, Handler: scrapperHandler}
	h := scrapperserver.HandlerWithOptions(serverImplementation,
		scrapperserver.StdHTTPServerOptions{ErrorHandlerFunc: scrapperserver.JSONErrorHandler})

	server := &http.Server{Handler: h, Addr: scrapperServerPort}

	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			baseLogger.Error("error starting scrapper http server", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer shutdownCancel()
	if err = server.Shutdown(shutdownCtx); err != nil {
		baseLogger.Error("error shutting down scrapper http server", slog.String("error", err.Error()))
	}

	if err = sched.Shutdown(); err != nil {
		baseLogger.Error("error shutting down scheduler", slog.String("error", err.Error()))
	}

	db.CloseConnectionPool()
}
