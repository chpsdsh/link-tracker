package scrapper

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/cache"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/senderfactory"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/database"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/scrapperclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/scrapperserver"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/scheduler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service"
)

const (
	envFilename = "scrapper.env"

	clientTimeout           = 15 * time.Second
	notificationChanBufSize = 10
)

func StartScrapper(baseLogger *slog.Logger) error {
	if err := godotenv.Load(envFilename); err != nil {
		baseLogger.Error("error loading file", slog.String("file", envFilename), slog.String("err", err.Error()))
		return fmt.Errorf("loading .env file: %w", err)
	}

	conf, dbConf, err := parseConfigs()
	if err != nil {
		baseLogger.Error("error parsing config", slog.String("err", err.Error()))
		return fmt.Errorf("parsing config: %w", err)
	}

	client := &http.Client{Timeout: clientTimeout}

	sched, err := gocron.NewScheduler()
	if err != nil {
		baseLogger.Error("error creating scheduler", slog.String("err", err.Error()))
		return fmt.Errorf("creating scheduler: %w", err)
	}

	db, err := database.NewDB(dbConf)
	if err != nil {
		baseLogger.Error("error connecting to database", slog.String("err", err.Error()))
		return fmt.Errorf("connecting to database: %w", err)
	}

	chatRepo, linkRepo, err := repository.CreateRepositories(db.GetDBPool(), conf.AssetType)
	if err != nil {
		baseLogger.Error("error creating repository", slog.String("err", err.Error()))
		return fmt.Errorf("creating repository: %w", err)
	}

	notificationsChan := make(chan pkg.KafkaLinkUpdate, notificationChanBufSize)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	sender, err := senderfactory.NewSender(ctx, conf, baseLogger, notificationsChan)
	if err != nil {
		baseLogger.Error("error creating sender", slog.String("err", err.Error()))
		return fmt.Errorf("creating sender: %w", err)
	}

	linksRequester := service.NewLinkRequester(
		scrapperclient.Client{Client: client, Config: conf},
		sender,
		linkRepo,
		conf.NumWorkers,
		conf.BatchSize,
		baseLogger,
	)

	wg := &sync.WaitGroup{}
	linksRequester.Start(ctx, wg)

	waitLinkRequester(wg, linksRequester)

	scrapperScheduler := scheduler.ScrapperScheduler{Scheduler: sched, LinksRequester: linksRequester}
	scrapperScheduler.StartScrapperScheduler()

	scrapperCache := cache.NewScrapperCacheClient(conf)

	scrapperHandler := service.LinksService{
		LinkRepo:   linkRepo,
		ChatsRepo:  chatRepo,
		Transactor: db,
		BaseLogger: baseLogger,
		CacheRepo:  scrapperCache,
	}

	scrapperServer := scrapperserver.NewScrapperHTTPServer(baseLogger, scrapperHandler)
	_ = scrapperServer.Start(ctx)

	<-ctx.Done()
	shutdown(scrapperServer, baseLogger, sched, db, notificationsChan, sender, scrapperCache)
	return nil
}

func waitLinkRequester(wg *sync.WaitGroup, linksRequester service.LinksRequester) {
	go func() {
		wg.Wait()
		close(linksRequester.LinksPool.LinksChan)
	}()
}

func shutdown(server scrapperserver.ScrapperHTTPServer,
	baseLogger *slog.Logger, sched gocron.Scheduler,
	db *database.DB,
	updatedChan chan pkg.KafkaLinkUpdate,
	sender service.Sender,
	cache cache.ScrapperCacheClient) {
	if err := server.Shutdown(); err != nil {
		baseLogger.Error("error shutting down scrapper http server", slog.String("error", err.Error()))
	}

	if err := sched.Shutdown(); err != nil {
		baseLogger.Error("error shutting down scheduler", slog.String("error", err.Error()))
	}
	sender.Close()
	close(updatedChan)
	db.CloseConnectionPool()

	if err := cache.Close(); err != nil {
		baseLogger.Error("error closing cache", slog.String("error", err.Error()))
	}
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
