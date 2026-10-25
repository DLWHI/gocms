package test

import (
	"fmt"
	"gocms/internal/password"
	"net/http"

	"gorm.io/gorm"
)

type Case struct {
	ID     string   `json:"id"`
	Group  string   `json:"group"`
	Title  string   `json:"title"`
	Source string   `json:"source"`
	Tags   []string `json:"tags"`
	Steps  []Step   `json:"steps"`
}

type Step struct {
	Actor   string         `json:"actor"`
	Request HttpRequest    `json:"request"`
	Expect  ExpectedResult `json:"expect"`
}

type Data struct {
	Meta      map[string]Value `json:"meta"`
	Fixtures  Fixtures         `json:"fixtures"`
	Constants map[string]Value `json:"constants"`
	Cases     []Case           `json:"cases"`
}

type Environment struct {
	DB       *gorm.DB
	Search   *Meilisearch
	Password password.PasswordService
	Router   http.Handler
	TestData *Data
}

func (e *Data) GetConstant(constant string) (string, error) {
	if value, ok := e.Constants[constant]; ok {
		return fmt.Sprintf("%v", value), nil
	}
	return "", fmt.Errorf("Missing constant %s", constant)
}

func (e *Data) GetSessionOf(actor string) string {
	return e.Fixtures.Actors[actor].Session
}
