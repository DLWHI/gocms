package main

import (
	"context"
	"gocms/internal/logger"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

const (
	CodeErrorOperation = -1
	CodeOk             = 1 << iota
	CodeErrorInit
	CodeErrorStart
	CodeErrorShutdown
)

func config() Config {
	init := pflag.BoolP("init", "i", false, "Migrate database and create index")
	user := pflag.BoolP("create-user", "", false, "Create user")
	wipe := pflag.BoolP("wipe", "W", false, "Wipe and re-create database and search index")
	port := pflag.StringP("port", "p", "8080", "CMS bound port")
	addr := pflag.StringP("address", "a", "", "CMS bound address")
	storage := pflag.StringP("storage", "s", "", "CMS media storage path")

	pflag.Parse()

	return Config{
		Init:    *init,
		Address: *addr,
		Wipe:    *wipe,
		Port:    *port,
		User:    *user,
		Storage: *storage,
	}
}

func main() {
	cfg := config()
	code := run(cfg)
	os.Exit(code)
}

func run(cfg Config) int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app := NewContainer()
	if err := app.Init(ctx, cfg); err != nil {
		logger.Error("Main", "Failed to initialize application", err)
		return CodeErrorInit
	}
	defer app.TearDown()

	if cfg.Init {
		if err := Migrate(ctx, app); err != nil {
			return CodeErrorOperation
		}
	} else if cfg.Wipe {
		if err := Wipe(ctx, app); err != nil {
			return CodeErrorOperation
		}
	} else if cfg.User {
		if err := CreateUser(ctx, app); err != nil {
			return CodeErrorOperation
		}
	} else {
		app.Start(ctx)

		<-ctx.Done()

		logger.Info("Main", "Shutdown signal received. Shutting down...")
		if !app.Stop(30 * time.Second) {
			return CodeErrorShutdown
		}
	}

	return CodeOk
}
