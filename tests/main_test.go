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
)

const (
	ConfigAddress = ""
	ConfigPort    = "8080"
	Filter        = ""
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

	ctx, cancel := signal.NotifyContext(t.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
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
				logger.Warn("Test.Users", "Failed to apply filter", err)
			}
			continue
		}

		if err = data.Fixtures.ApplyFixtures(ctx, Env); err != nil {
			logger.Error("Test.Users", "Failed to apply fixtures", err)
			t.FailNow()
		}

		logger.Info("Test.Users", fmt.Sprintf("Running test: %s", testcase.ID))
		var result *test.Result
		ran++
		for _, step := range testcase.Steps {
			result = step.Perform(ctx, data, Env)
			if result != nil {
				failMessage := fmt.Sprintf("Test %s failed", testcase.ID)
				logger.Error("Test.Users", failMessage, result.Reason, "case", testcase, "step", step, "response_code", result.ResponseStatus, "message", result.Message)
				t.Fail()
				break
			}
		}
		if result == nil {
			succeed++
			logger.Info("Test.Users", fmt.Sprintf("Test %s passed", testcase.ID))
		}
		if err := Env.Clear(ctx); err != nil {
			logger.Error("Test.Users", "Failed to clear environment", err)
			t.FailNow()
		}
	}

}

func TestLocales(t *testing.T) {
	ran := 0
	succeed := 0
	defer func() {
		logger.Info("Tests", fmt.Sprintf("Finished running locale tests. Ran = %d tests. Passed = %d. Failed = %d", ran, succeed, ran-succeed))
	}()

	ctx, cancel := signal.NotifyContext(t.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	logger.Info("Test.Locales", "Loading test data")
	data, err := test.LoadDataFromFile("./test_cases/locales.json", "1")
	if err != nil {
		logger.Error("Test.Locales", "Failed to load test data", err)
		t.FailNow()
	}
	Env.TestData = data

	for _, testcase := range data.Cases {
		if matched, err := regexp.MatchString(Filter, testcase.ID); err != nil || !matched {
			if err != nil {
				logger.Warn("Test.Locales", "Failed to apply filter", err)
			}
			continue
		}

		if err = data.Fixtures.ApplyFixtures(ctx, Env); err != nil {
			logger.Error("Test.Locales", "Failed to apply fixtures", err)
			t.FailNow()
		}

		logger.Info("Test.Locales", fmt.Sprintf("Running test: %s", testcase.ID))
		var result *test.Result
		ran++
		for _, step := range testcase.Steps {
			result = step.Perform(ctx, data, Env)
			if result != nil {
				failMessage := fmt.Sprintf("Test %s failed", testcase.ID)
				logger.Error("Test.Locales", failMessage, result.Reason, "case", testcase, "step", step, "response_code", result.ResponseStatus, "message", result.Message)
				t.Fail()
				break
			}
		}
		if result == nil {
			succeed++
			logger.Info("Test.Locales", fmt.Sprintf("Test %s passed", testcase.ID))
		}
		if err := Env.Clear(ctx); err != nil {
			logger.Error("Test.Locales", "Failed to clear environment", err)
			t.FailNow()
		}
	}

}

func TestCategories(t *testing.T) {
	ran := 0
	succeed := 0
	defer func() {
		logger.Info("Tests", fmt.Sprintf("Finished running category tests. Ran = %d tests. Passed = %d. Failed = %d", ran, succeed, ran-succeed))
	}()

	ctx, cancel := signal.NotifyContext(t.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	logger.Info("Test.Categories", "Loading test data")
	data, err := test.LoadDataFromFile("./test_cases/categories.json", "1")
	if err != nil {
		logger.Error("Test.Categories", "Failed to load test data", err)
		t.FailNow()
	}
	Env.TestData = data

	for _, testcase := range data.Cases {
		if matched, err := regexp.MatchString(Filter, testcase.ID); err != nil || !matched {
			if err != nil {
				logger.Warn("Test.Categories", "Failed to apply filter", err)
			}
			continue
		}

		if err = data.Fixtures.ApplyFixtures(ctx, Env); err != nil {
			logger.Error("Test.Categories", "Failed to apply fixtures", err)
			t.FailNow()
		}

		logger.Info("Test.Categories", fmt.Sprintf("Running test: %s", testcase.ID))
		var result *test.Result
		ran++
		for _, step := range testcase.Steps {
			result = step.Perform(ctx, data, Env)
			if result != nil {
				failMessage := fmt.Sprintf("Test %s failed", testcase.ID)
				logger.Error("Test.Categories", failMessage, result.Reason, "case", testcase, "step", step, "response_code", result.ResponseStatus, "message", result.Message)
				t.Fail()
				break
			}
		}
		if result == nil {
			succeed++
			logger.Info("Test.Categories", fmt.Sprintf("Test %s passed", testcase.ID))
		}
		if err := Env.Clear(ctx); err != nil {
			logger.Error("Test.Categories", "Failed to clear environment", err)
			t.FailNow()
		}
	}

}
