package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	app       *Container
	mux       *http.ServeMux
	sleepTime time.Duration = 3 * time.Second // time before test starts pulling from meilisearch index
)

func TestMain(m *testing.M) {
	if err := run(m); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}

func run(m *testing.M) error {
	app = NewContainer()
	mux = http.NewServeMux()

	err := app.Init()
	if err != nil {
		return err
	}
	defer app.TearDown()

	router := NewRouter()

	router.AddUserRoutes(mux, app.UserHandler)
	router.AddEditorRoutes(mux, app.EditorHandler)
	router.AddCategoryRoutes(mux, app.CategoryHandler)
	router.AddPostRoutes(mux, app.PostHandler)
	router.AddAdsRoutes(mux, app.AdHandler)
	router.AddLocaleRoutes(mux, app.LocalesHandler)

	if m.Run() != 0 {
		return fmt.Errorf("Test run failed")
	}
	return nil
}
