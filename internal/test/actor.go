package test

import (
	"bytes"
	"context"
	"encoding/json"
	"gocms/internal/domain"
	"gocms/internal/session"
	"gocms/internal/user"
	"io"
	"net/http"
	"net/http/httptest"
)

type Actor struct {
	Username    string   `json:"username"`
	UsersCaps   []string `json:"users_caps"`
	Permissions uint32   `json:"permissions"`
	Session     string   `json:"-"`
}

func GetActorPermissions(entity string, actor *Actor) uint32 {
	return domain.PermissionsOnEntity(entity, actor.UsersCaps...)
}

func (a *Actor) PerformLogin(ctx context.Context, e *Environment, passwd Password) error {
	var payload struct {
		Username string   `json:"username"`
		Passwd   Password `json:"passwd"`
	}
	payload.Username = a.Username
	payload.Passwd = passwd

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(&payload); err != nil {
		return err
	}
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/cms/login", io.NopCloser(buf))

	req.Header.Set("User-Agent", "Test-Client")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	e.Router.ServeHTTP(w, req)

	var response session.Session
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		return err
	}
	a.Session = response.ID
	return nil
}

func (a *Actor) ToRecord(pwdhash []byte) user.Record {
	var u user.Record
	u.Password = pwdhash
	u.Username = a.Username
	u.Permissions = a.Permissions
	return u
}

func (a *Actor) ToDomain() domain.User {
	var u domain.User
	u.Username = a.Username
	u.Permissions = a.Permissions
	return u
}
