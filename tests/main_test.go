package tests

import (
	"context"
	"fmt"
	"gocms/internal/logger"
	"gocms/internal/test"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"testing"
	"time"
)

const (
	ConfigAddress = ""
	ConfigPort    = "8080"
	Filter        = "^PASSWORD-001"
)

var (
	App *Container
	Env *test.Environment
)

func TestMain(m *testing.M) {
	if err := run(m); err != nil {
		os.Exit(1)
	}

}

func run(m *testing.M) error {
	lg := slog.New(logger.NewReasonedHandler(nil))
	slog.SetDefault(lg)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	App = NewContainer()
	logger.Info("Test.Users", "Initializing services and connections")
	if err := App.Init(ctx); err != nil {
		logger.Error("Main", "Failed to initialize application", err)
		return err
	}
	defer App.TearDown()

	Env = &test.Environment{
		DB: App.DB,
		Search: test.NewMeilisearch(
			App.Meilisearch,
			UsersSearchIndexName,
			PostsSearchIndexName,
			CategoriesSearchIndexName,
			EditorsSearchIndexName,
			AdsSearchIndexName,
		),
		Password: App.PasswordService,
		Router:   App.Router,
		Redis:    App.Redis,
		Outbox:   App.MessageRelay,
	}

	App.Start(ctx)
	defer App.Stop(30 * time.Second)

	status := m.Run()
	cancel()
	if status != 0 {
		return fmt.Errorf("Test run failed")
	}
	return nil
}

func TestUsers(t *testing.T) {
	ran := 0
	succeed := 0
	defer func() {
		logger.Info("Tests", fmt.Sprintf("Finished running user tests. Ran = %d tests. Passed = %d. Failed = %d", ran, succeed, ran-succeed))
	}()

	ctx := t.Context()
	logger.Info("Test.Users", "Loading test data")
	data, err := test.LoadDataFromFile("./test_cases/users.json", "1")
	if err != nil {
		logger.Error("Test.Users", "Failed to load test data", err)
		t.FailNow()
	}
	Env.TestData = data

	for _, testcase := range data.Cases {
		if matched, err := regexp.MatchString(Filter, testcase.ID); err != nil || !matched {
			if err != nil {
				logger.Warn("Test.User", "Failed to apply filter", err)
			}
			continue
		}

		if err = data.Fixtures.ApplyFixtures(ctx, Env); err != nil {
			logger.Error("Test.Users", "Failed to apply fixtures", err)
			t.FailNow()
		}

		logger.Info("Test.User", fmt.Sprintf("Running test: %s", testcase.ID))
		var status error
		ran++
		for _, step := range testcase.Steps {
			if res, err := step.Perform(ctx, data, Env); err != nil {
				status = err
				logger.Error("Test.User", fmt.Sprintf("Test %s failed", testcase.ID), err, "case", testcase, "step", step, "response", res)
				t.Fail()
				break
			}
		}
		if status == nil {
			succeed++
			logger.Info("Test.User", fmt.Sprintf("Test %s passed", testcase.ID))
		}
		if err := Env.Clear(ctx); err != nil {
			logger.Error("Test.User", "Failed to clear environment", err)
			t.FailNow()
		}
	}

}
