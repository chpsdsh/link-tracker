package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/scheduler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service"
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

	conf, dbConf, err := parseConfigs()
	if err != nil {
		baseLogger.Error("error parsing config", slog.String("err", err.Error()))
		os.Exit(1)
	}

	client := &http.Client{Timeout: clientTimeout}

	sched, err := gocron.NewScheduler()
	if err != nil {
		baseLogger.Error("error parsing environment variables", slog.String("err", err.Error()))
		os.Exit(1)
	}

	db, err := database.NewDB(dbConf)
	if err != nil {
		baseLogger.Error("error connecting to database", slog.String("err", err.Error()))
		os.Exit(1)
	}

	chatRepo, linkRepo, err := repository.CreateRepositories(db.GetDBPool(), conf.AssetType)
	if err != nil {
		baseLogger.Error("error creating repository", slog.String("err", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	linksRequester := service.NewLinkRequester(
		scrapperclient.Client{Client: client, Config: conf},
		linkRepo,
		conf.NumWorkers,
		conf.BatchSize,
		baseLogger,
	)

	wg := &sync.WaitGroup{}
	linksRequester.Start(ctx, wg)

	go func() {
		wg.Wait()
		close(linksRequester.LinksPool.LinksChan)
	}()

	scrapperScheduler := scheduler.ScrapperScheduler{Scheduler: sched, LinksRequester: linksRequester}
	scrapperScheduler.StartScrapperScheduler()

	scrapperHandler := service.LinksService{
		LinkRepo:   linkRepo,
		ChatsRepo:  chatRepo,
		Transactor: db,
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

func parseConfigs() (config.Config, config.PostgresConfig, error) {
	conf, err := config.ParseConfig()
	if err != nil {
		return config.Config{}, config.PostgresConfig{}, fmt.Errorf("error parsing scrapper config: %w", err)
	}

	dbConf, err := config.ParsePostgresConfig()
	if err != nil {
		return config.Config{}, config.PostgresConfig{}, fmt.Errorf("error parsing postgres config: %w", err)
	}
	return conf, dbConf, nil
}
