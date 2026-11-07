package test

import (
	"context"
	"encoding/json"
	"fmt"
	"gocms/internal/advertisement"
	"gocms/internal/category"
	"gocms/internal/domain"
	"gocms/internal/editor"
	"gocms/internal/lang"
	"gocms/internal/outbox"
	"gocms/internal/password"
	"gocms/internal/post"
	"gocms/internal/user"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

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
	Lang     *lang.LocalService
	Password password.Service
	Outbox   outbox.MessageRepository
	Router   http.Handler
	TestData *Data
}

type TableEntity interface {
	TableName() string
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

func (d *Data) Resolve(v JSONValue) (JSONValue, error) {
	m, ok := v.(map[string]JSONValue)
	if ok {
		for name, arg := range m {
			if strings.HasPrefix(name, "$") {
				val, err := d.ExpandToken(name, arg)
				if err != nil {
					return nil, err
				}
				return val, nil
			}
			return v, nil
		}
	}
	return v, nil
}

func (d *Data) ResolveString(v JSONValue) (string, error) {
	m, ok := v.(map[string]JSONValue)
	if ok {
		for name, arg := range m {
			if strings.HasPrefix(name, "$") {
				val, err := d.ExpandToken(name, arg)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("%v", val), nil
			}
		}
	}
	return fmt.Sprintf("%v", v), nil
}

func (d *Data) ResolveMap(values map[string]JSONValue) (map[string]JSONValue, error) {
	res := make(map[string]JSONValue, len(values))
	for key, value := range values {
		resolved, err := d.Resolve(value)
		if err != nil {
			return nil, err
		}
		res[key] = resolved
	}
	return res, nil
}

func (d *Data) ExpandToken(name string, arg JSONValue) (JSONValue, error) {
	switch name {
	case "$const":
		key, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("$const expects a string")
		}
		val, ok := d.Constants[key]
		if !ok {
			return nil, fmt.Errorf("unknown constant %q", key)
		}
		return val, nil

	case "$repeat":
		m, ok := arg.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("$repeat expects {s, n}")
		}
		s, ok1 := m["s"].(string)
		n, ok2 := ToLen(m["n"])
		if !ok1 || !ok2 || n < 0 {
			return nil, fmt.Errorf("$repeat needs string s and non-negative integer n")
		}
		return strings.Repeat(s, int(n)), nil
	case "$email_of_length":
		n, ok := ToLen(arg)
		const domain = "@example.com"
		if !ok || int(n) <= len(domain) {
			return nil, fmt.Errorf("$email_of_length needs an integer > %d", len(domain))
		}
		return strings.Repeat("a", int(n)-len(domain)) + domain, nil
	case "$raw_session":
		actor, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("$raw_session needs to be a string")
		}
		return d.GetSessionOf(actor), nil
	case "$actor_perms":
		username, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("$actor_perms needs to be a string")
		}
		actor, ok := d.Fixtures.Actors[username]
		if !ok {
			return nil, fmt.Errorf("$actor_perms requests unknown actor")
		}
		return json.Number(strconv.FormatUint(uint64(actor.Permissions), 10)), nil
	case "$printf":
		m, ok := arg.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("$repeat expects {format, values}")
		}
		format, ok1 := m["format"].(string)
		values, ok2 := m["values"].([]any)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("$repeat needs format string and values array")
		}
		resolved := make([]any, 0, len(values))
		for _, val := range values {
			v, err := d.Resolve(val)
			if err != nil {
				return nil, fmt.Errorf("$printf.values: %w", err)
			}
			resolved = append(resolved, v)
		}
		return fmt.Sprintf(format, resolved...), nil
	default:
		return nil, fmt.Errorf("value is not a state token")
	}
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
	entites := []any{
		&lang.Language{},
		&user.Record{},
		&category.Record{},
		&editor.Record{},
		&post.Record{},
		&advertisement.Record{},
		&outbox.Message{},
	}
	incrementables := []any{
		&post.Record{},
		&advertisement.Record{},
	}

	for _, entity := range entites {
		if err := e.DB.Unscoped().Where("1 = 1").Delete(entity).Error; err != nil {
			return err
		}
	}

	for _, entity := range incrementables {
		table, ok := entity.(TableEntity)
		if !ok {
			return fmt.Errorf("%w: entity %T does not hold table name behind itself", domain.ErrInval, entity)
		}
		if err := e.DB.Exec("ALTER TABLE ? AUTO_INCREMENT = 1;", table.TableName()).Error; err != nil {
			return err
		}
	}

	if err := e.Redis.FlushDB(ctx).Err(); err != nil {
		return err
	}

	indexes := []meilisearch.IndexManager{
		e.Search.UsersIndex,
		e.Search.CategoriesIndex,
		e.Search.EditorsIndex,
		e.Search.PostsIndex,
		e.Search.AdsIndex,
	}
	for _, index := range indexes {

		task, err := index.DeleteAllDocumentsWithContext(ctx, nil)
		if err != nil {
			return err
		}
		if err := e.Search.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	return nil
}
