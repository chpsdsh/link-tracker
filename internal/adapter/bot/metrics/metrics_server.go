package metrics

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	metricsServerAddr = ":8081"
)

type Server struct {
	server *http.Server
	logger *slog.Logger
}

func NewMetricsServer(logger *slog.Logger) Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    metricsServerAddr,
		Handler: mux,
	}
	return Server{
		server: server,
		logger: logger,
	}
}

func (s Server) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("error starting bot http server", slog.String("err", err.Error()))
		}
	}()
}

func (s Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("error shutting down bot http server", slog.String("err", err.Error()))
		return fmt.Errorf("shutting down bot http server: %w", err)
	}
	return nil
}
