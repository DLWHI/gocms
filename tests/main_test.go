package tests

import (
	"context"
	"fmt"
	"gocms/internal/logger"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

var (
	app       *Container
	address                 = ""
	port                    = "8080"
	sleepTime time.Duration = 3 * time.Second // time before test starts pulling from meilisearch index
)

func TestMain(m *testing.M) {
	if err := run(m); err != nil {
		os.Exit(1)
	}

}

func run(m *testing.M) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app = NewContainer()
	if err := app.Init(ctx); err != nil {
		logger.Error("Main", "Failed to initialize application", err)
		return err
	}
	defer app.TearDown()

	app.Start(ctx)

	<-ctx.Done()

	logger.Info("Main", "Shutdown signal received. Shutting down...")
	app.Stop(30 * time.Second)

	if m.Run() != 0 {
		return fmt.Errorf("Test run failed")
	}
	return nil
}
