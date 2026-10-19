package httpserver

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/logger"
	"net/http"
	"sync"
)

type HttpRunner struct {
	server http.Server
}

func NewHttpRunner(addr string, port string, handler http.Handler) *HttpRunner {
	return &HttpRunner{
		server: http.Server{
			Addr:    fmt.Sprintf("%s:%s", addr, port),
			Handler: handler,
		},
	}
}

func (r *HttpRunner) Start(ctx context.Context, group *sync.WaitGroup, _ ...any) {
	defer group.Done()

	logger.Info("Main", "Starting server on port :8080...")
	if err := r.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Main", "HTTP server stopped unexpectedly", err)
	}
}

func (r *HttpRunner) Stop(ctx context.Context) error {
	if err := r.server.Shutdown(ctx); err != nil {
		logger.Error("Main", "Server graceful shutdown failed", err)
		return err
	}
	logger.Info("Main", "Server stopped gracefully.")
	return nil
}
