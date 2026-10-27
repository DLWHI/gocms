package test

import (
	"context"
	"encoding/json"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/outbox"
	"gocms/internal/password"
	"gocms/internal/user"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
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
	Actor      string         `json:"actor"`
	Request    HttpRequest    `json:"request"`
	Expect     ExpectedResult `json:"expect"`
	Eventually uint           `json:"eventually_ms"`
}

type Data struct {
	Meta      map[string]JSONValue `json:"meta"`
	Fixtures  Fixtures             `json:"fixtures"`
	Constants map[string]JSONValue `json:"constants"`
	Cases     []Case               `json:"cases"`
}

type Environment struct {
	DB       *gorm.DB
	Search   *Meilisearch
	Redis    *redis.Client
	Password password.PasswordService
	Outbox   outbox.MessageRepository
	Router   http.Handler
	TestData *Data
}

func LoadDataFromFile(path string, format string) (*Data, error) {
	cases, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data := &Data{}
	decoder := json.NewDecoder(cases)
	decoder.UseNumber()
	if err := decoder.Decode(data); err != nil {
		return nil, err
	}

	if val, ok := data.Meta["format_version"]; !ok || fmt.Sprintf("%v", val) != format {
		return nil, err
	}

	for _, actor := range data.Fixtures.Actors {
		if actor.Permissions == 0 {
			for entity, perms := range actor.Caps {
				actor.Permissions |= domain.PermissionsOnEntity(entity, perms...)
			}
		}
	}
	return data, nil
}

func (e *Environment) Clear(ctx context.Context) error {

	if err := e.DB.Unscoped().Where("1 = 1").Delete(&user.Record{}).Error; err != nil {
		return err
	}
	if err := e.DB.Unscoped().Where("1 = 1").Delete(&outbox.Message{}).Error; err != nil {
		return err
	}

	if err := e.Redis.FlushDB(ctx).Err(); err != nil {
		return err
	}

	task, err := e.Search.UsersIndex.DeleteAllDocumentsWithContext(ctx, nil)
	if err != nil {
		return err
	}
	if err := e.Search.WaitForTask(ctx, task); err != nil {
		return err
	}

	return nil
}

func (d *Data) GetConstant(constant string) (string, error) {
	if value, ok := d.Constants[constant]; ok {
		return fmt.Sprintf("%v", value), nil
	}
	return "", fmt.Errorf("Missing constant %s", constant)
}

func (d *Data) GetSessionOf(actor string) string {
	return d.Fixtures.Actors[actor].Session
}

func (s *Step) Perform(ctx context.Context, d *Data, e *Environment) (*http.Response, error) {
	actor, ok := d.Fixtures.Actors[s.Actor]
	if !ok && s.Actor != "anonymous" {
		return nil, fmt.Errorf("Requested step for unknown actor")
	}

	res, err := e.PerformRequest(ctx, actor, s.Request)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()

	if err := e.Outbox.Dispatch(ctx, 5); err != nil {
		return res, err
	}

	if err := s.Expect.Check(d, res); err != nil {
		return res, err
	}

	for _, external := range s.Expect.ExternalChecks {
		if err := external.Assert(ctx, e); err != nil {
			return res, err
		}
	}
	return nil, nil
}
