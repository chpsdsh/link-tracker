package scrapper

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/cache"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/notificationsender"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/repository/outboxrepo"
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
	envFilename             = "scrapper.env"
	notificationChanBufSize = 10
)

type App struct {
	logger *slog.Logger
	conf   config.ScrapperConfig

	ctx    context.Context
	cancel context.CancelFunc

	db *database.DB

	chatRepo   service.ChatRepository
	linkRepo   service.LinkRepository
	outboxRepo *outboxrepo.OutboxRepository

	sender service.Sender
	cache  cache.ScrapperCacheClient

	linksRequester service.LinksRequester
	updatesSender  service.UpdatesSender

	scheduler gocron.Scheduler
	server    scrapperserver.ScrapperHTTPServer

	notificationsChan chan pkg.KafkaLinkUpdate
	wg                sync.WaitGroup
}

func StartScrapper(baseLogger *slog.Logger) error {
	app, err := NewApp(baseLogger)
	if err != nil {
		return err
	}

	return app.Run()
}

func NewApp(logger *slog.Logger) (*App, error) {
	if err := godotenv.Load(envFilename); err != nil {
		logger.Error("error loading file", slog.String("file", envFilename), slog.String("err", err.Error()))
		return nil, fmt.Errorf("loading .env file: %w", err)
	}

	conf, err := config.ParseConfig()
	if err != nil {
		logger.Error("error parsing config", slog.String("err", err.Error()))
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	app := &App{
		logger:            logger,
		conf:              conf,
		ctx:               ctx,
		cancel:            cancel,
		notificationsChan: make(chan pkg.KafkaLinkUpdate, notificationChanBufSize),
	}

	if err = app.initDB(); err != nil {
		cancel()
		return nil, err
	}

	if err = app.initRepositories(); err != nil {
		cancel()
		app.closeDB()
		return nil, err
	}

	if err = app.initSender(); err != nil {
		cancel()
		app.closeDB()
		return nil, err
	}

	if err = app.initCache(); err != nil {
		cancel()
		app.closeSender()
		app.closeDB()
		return nil, err
	}

	if err = app.initScheduler(); err != nil {
		cancel()
		app.closeCache()
		app.closeSender()
		app.closeDB()
		return nil, err
	}

	app.initServices()
	app.initServer()

	return app, nil
}

func (a *App) Run() error {
	a.linksRequester.Start(a.ctx, &a.wg)

	go func() {
		a.wg.Wait()
		close(a.linksRequester.LinksPool.LinksChan)
	}()

	scrapperScheduler := scheduler.ScrapperScheduler{
		Scheduler:      a.scheduler,
		LinksRequester: a.linksRequester,
		UpdatesSender:  a.updatesSender,
	}
	scrapperScheduler.StartScrapperScheduler()

	if err := a.server.Start(a.ctx); err != nil {
		a.logger.Error("error starting server", slog.String("err", err.Error()))
		return fmt.Errorf("starting scrapper server: %w", err)
	}

	<-a.ctx.Done()

	a.Shutdown()
	return nil
}

func (a *App) Shutdown() {
	a.cancel()

	if err := a.server.Shutdown(); err != nil {
		a.logger.Error("error shutting down scrapper http server", slog.String("error", err.Error()))
	}

	if err := a.scheduler.Shutdown(); err != nil {
		a.logger.Error("error shutting down scheduler", slog.String("error", err.Error()))
	}

	a.closeSender()

	close(a.notificationsChan)

	a.closeCache()
	a.closeDB()
}

func (a *App) initDB() error {
	db, err := database.NewDB(a.conf.PostgresConfig)
	if err != nil {
		a.logger.Error("error connecting to database", slog.String("err", err.Error()))
		return fmt.Errorf("connecting to database: %w", err)
	}

	a.db = db
	return nil
}

func (a *App) initRepositories() error {
	chatRepo, linkRepo, err := repository.CreateRepositories(
		a.db.GetDBPool(),
		a.conf.AssetType,
	)
	if err != nil {
		a.logger.Error("error creating repository", slog.String("err", err.Error()))
		return fmt.Errorf("creating repository: %w", err)
	}

	a.chatRepo = chatRepo
	a.linkRepo = linkRepo
	a.outboxRepo = outboxrepo.NewOutboxRepository(a.db.GetDBPool())

	return nil
}

func (a *App) initSender() error {
	sender, err := notificationsender.NewSender(
		a.ctx,
		a.conf,
		a.logger,
		a.notificationsChan,
	)
	if err != nil {
		a.logger.Error("error creating sender", slog.String("err", err.Error()))
		return fmt.Errorf("creating sender: %w", err)
	}

	a.sender = sender
	return nil
}

func (a *App) initCache() error {
	scrapperCache, err := cache.NewScrapperCacheClient(a.conf)
	if err != nil {
		a.logger.Error("error creating cache", slog.String("err", err.Error()))
		return fmt.Errorf("creating cache: %w", err)
	}

	a.cache = scrapperCache
	return nil
}

func (a *App) initScheduler() error {
	sched, err := gocron.NewScheduler()
	if err != nil {
		a.logger.Error("error creating scheduler", slog.String("err", err.Error()))
		return fmt.Errorf("creating scheduler: %w", err)
	}

	a.scheduler = sched
	return nil
}

func (a *App) initServices() {
	a.linksRequester = service.NewLinkRequester(
		scrapperclient.NewScrapperClient(a.conf),
		a.sender,
		a.linkRepo,
		a.outboxRepo,
		a.db,
		a.conf.NumWorkers,
		a.conf.BatchSize,
		a.logger,
	)

	a.updatesSender = service.UpdatesSender{
		OutboxRepo:         a.outboxRepo,
		Transactor:         a.db,
		NotificationSender: a.sender,
		BaseLogger:         a.logger,
	}
}

func (a *App) initServer() {
	scrapperHandler := service.LinksService{
		LinkRepo:   a.linkRepo,
		ChatsRepo:  a.chatRepo,
		Transactor: a.db,
		BaseLogger: a.logger,
		CacheRepo:  a.cache,
	}

	a.server = scrapperserver.NewScrapperHTTPServer(a.logger, scrapperHandler, a.conf)
}

func (a *App) closeSender() {
	if a.sender != nil {
		a.sender.Close()
	}
}

func (a *App) closeCache() {
	if err := a.cache.Close(); err != nil {
		a.logger.Error("error closing cache", slog.String("error", err.Error()))
	}
}

func (a *App) closeDB() {
	if a.db != nil {
		a.db.CloseConnectionPool()
	}
}
