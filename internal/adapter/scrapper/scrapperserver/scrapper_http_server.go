package scrapperserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service"
)

const (
	scrapperServerPort = ":8081"
	shutdownDuration   = 10 * time.Second
)

type ScrapperHTTPServer struct {
	server *http.Server
	logger *slog.Logger
}

func NewScrapperHTTPServer(baseLogger *slog.Logger, scrapperHandler service.LinksService) ScrapperHTTPServer {
	serverImplementation := ScrapperServer{BaseLogger: baseLogger, Handler: scrapperHandler}
	h := HandlerWithOptions(serverImplementation,
		StdHTTPServerOptions{ErrorHandlerFunc: JSONErrorHandler})

	server := &http.Server{Handler: h, Addr: scrapperServerPort}
	return ScrapperHTTPServer{server: server, logger: baseLogger}
}

func (s ScrapperHTTPServer) Start(_ context.Context) error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("error starting bot http server", slog.String("err", err.Error()))
		}
	}()
	return nil
}

func (s ScrapperHTTPServer) Shutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer cancel()
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("error shutting down bot http server", slog.String("err", err.Error()))
		return fmt.Errorf("shutting down bot http server: %w", err)
	}
	return nil
}
