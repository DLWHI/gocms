package test

import (
	"context"
	"encoding/json"
	"fmt"
	"gocms/internal/category"
	"gocms/internal/domain"
	"gocms/internal/language"
	"gocms/internal/outbox"
	"gocms/internal/password"
	"gocms/internal/user"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	Password password.PasswordService
	Outbox   outbox.MessageRepository
	Router   http.Handler
	TestData *Data
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

	if err := e.DB.Unscoped().Where("1 = 1").Delete(&user.Record{}).Error; err != nil {
		return err
	}
	if err := e.DB.Unscoped().Where("1 = 1").Delete(&language.Record{}).Error; err != nil {
		return err
	}
	if err := e.DB.Unscoped().Where("1 = 1").Delete(&category.Record{}).Error; err != nil {
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
