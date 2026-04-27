package botserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
)

const (
	botServerAddr    = ":8080"
	shutdownDuration = 10 * time.Second
)

type BotHttpServer struct {
	server *http.Server
	logger *slog.Logger
}

func NewBotHttpServer(baseLogger *slog.Logger, telegramHandler handler.TelegramHandler) BotHttpServer {
	router := UpdatesRouter{BaseLogger: baseLogger, Handler: telegramHandler}
	mux := http.NewServeMux()
	h := HandlerFromMux(&router, mux)

	server := &http.Server{
		Handler: h,
		Addr:    botServerAddr,
	}
	return BotHttpServer{server: server, logger: baseLogger}
}

func (s BotHttpServer) Start(_ context.Context) error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("error starting bot http server", slog.String("err", err.Error()))
		}
	}()
	return nil
}

func (s BotHttpServer) Shutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer cancel()
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("error shutting down bot http server", slog.String("err", err.Error()))
		return err
	}
	return nil
}
