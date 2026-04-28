package botserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
)

const (
	botServerAddr    = ":8080"
	shutdownDuration = 10 * time.Second
)

type BotHTTPServer struct {
	server *http.Server
	logger *slog.Logger
}

func NewBotHTTPServer(baseLogger *slog.Logger, telegramHandler handler.TelegramHandler) BotHTTPServer {
	router := UpdatesRouter{BaseLogger: baseLogger, Handler: telegramHandler}
	mux := http.NewServeMux()
	h := HandlerFromMux(&router, mux)

	server := &http.Server{
		Handler: h,
		Addr:    botServerAddr,
	}
	return BotHTTPServer{server: server, logger: baseLogger}
}

func (s BotHTTPServer) Start(_ context.Context) error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("error starting bot http server", slog.String("err", err.Error()))
		}
	}()
	return nil
}

func (s BotHTTPServer) Shutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer cancel()
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("error shutting down bot http server", slog.String("err", err.Error()))
		return fmt.Errorf("shutting down bot http server: %w", err)
	}
	return nil
}
