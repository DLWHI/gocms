package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/search"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
	se   SearchEngine
}

func NewService(db Repository, se SearchEngine) *Service {
	return &Service{
		repo: db,
		se:   se,
	}
}

func (s *Service) IsAllowedUsernameRune(r rune) bool {
	return strings.ContainsRune(".", r) || unicode.IsSpace(r)
}

func (s *Service) IsAllowedUsername(ctx context.Context, value string) error {
	if strings.IndexFunc(value, s.IsAllowedUsernameRune) >= 0 {
		err := fmt.Errorf("%w: username must not contain spaces or periods", errs.ErrInval)
		logger.Warn("User.Service", "Username must not contain spaces or periods", err)
		return err
	}

	_, err := s.repo.Get(ctx, value)
	if err == nil {
		return fmt.Errorf("%v: Username already exists", errs.ErrInval)
	} else if errors.Is(err, errs.ErrNotFound) {
		return nil
	} else {
		logger.Error("User.Service", "Failed to check username availability", err)
		return err
	}
}

func (s *Service) Create(ctx context.Context, user *User) error {
	if !auth.HasPermission(ctx, "users", "create") {
		clear(user.Passwd)
		return errs.ErrCred
	}

	user.Username = strings.TrimSpace(strings.ToLower(user.Username))
	if user.Username == "" {
		clear(user.Passwd)
		err := fmt.Errorf("%w: username cannot be empty", errs.ErrInval)
		logger.Warn("User.Service", "Username is required to create a user", err)
		return err
	} else if strings.IndexFunc(user.Username, s.IsAllowedUsernameRune) >= 0 {
		clear(user.Passwd)
		err := fmt.Errorf("%w: username must not contain spaces or periods", errs.ErrInval)
		logger.Warn("User.Service", "Username must not contain spaces or periods", err)
		return err
	}

	hash, err := bcrypt.GenerateFromPassword(user.Passwd, 12)
	clear(user.Passwd)
	if err != nil {
		clear(hash)
		logger.Error("User.Service", "Failed to hash the user password", err)
		return err
	}

	user.Passwd = hash
	user.Permissions = auth.RoleEditor

	err = s.repo.Create(ctx, user)
	clear(hash)
	clear(user.Passwd)
	if err != nil {
		logger.Error("User.Service", "Failed to create user", err)
		return err
	}

	return nil
}

func (s *Service) UpdateLogin(ctx context.Context, username string) (*Document, error) {
	user, err := s.repo.UpdateLogin(ctx, username, time.Now())
	if err != nil {
		logger.Error("User.Service", "Failed to update the user's last login time", err)
		return nil, err
	}

	val := user.ToDocument()
	return &val, nil
}

func (s *Service) Get(ctx context.Context, username string) (*Document, error) {
	user, err := s.repo.Get(ctx, username)
	if err != nil {
		logger.Error("User.Service", "Failed to retrieve user", err)
		return nil, err
	}
	clear(user.Passwd)

	val := user.ToDocument()
	return &val, nil
}

func (s *Service) Search(ctx context.Context, filters *SearchParams) (*SearchResult, error) {
	search.ClampParams(&filters.Params)

	users, err := s.se.Search(ctx, filters)
	if err != nil {
		logger.Error("User.Service", "Failed to search users", err)
		return nil, err
	}
	return users, nil
}

func (s *Service) All(ctx context.Context) ([]User, error) {

	users, err := s.repo.GetAll(ctx)
	if err != nil {
		logger.Error("User.Service", "Failed to list users", err)
		return nil, err
	}

	return users, nil
}

func (s *Service) Suspend(ctx context.Context, username string) error {
	if !auth.HasPermission(ctx, "users", "delete") {
		return errs.ErrCred
	}

	if err := s.repo.Delete(ctx, username); err != nil {
		logger.Error("User.Service", "Failed to suspend user", err)
		return err
	}

	return nil
}

func (s *Service) PasswordMatches(ctx context.Context, username string, passwd Password) error {
	hash, err := s.repo.GetPasswdHash(ctx, username)
	if err != nil {
		logger.Error("User.Service", "Failed to retreive user's password", err)
		return err
	}

	err = bcrypt.CompareHashAndPassword(hash, passwd)
	clear(hash)
	if err != nil {
		return errs.ErrCred
	}

	return nil
}

func (s *Service) ChangeEmail(ctx context.Context, username string, email string) error {
	initiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return errs.ErrCred
	}

	if initiator != username && !auth.HasPermission(ctx, "users", "edit") {
		return errs.ErrCred
	}

	diff := map[string]interface{}{
		"email": email,
	}

	if _, err := s.repo.Update(ctx, username, diff); err != nil {
		logger.Error("User.Service", "Failed to update the user's email address", err)
		return err
	}

	return nil
}

func (s *Service) ChangePasswd(ctx context.Context, username string, passwd domain.Password) error {
	initiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return errs.ErrCred
	}

	if initiator != username && !auth.HasPermission(ctx, "users", "edit") {
		return errs.ErrCred
	}

	hash, err := bcrypt.GenerateFromPassword(passwd, 12)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the replacement password", err)
		return err
	}

	err = s.repo.UpdatePasswd(ctx, username, hash)
	clear(hash)
	if err != nil {
		logger.Error("User.Service", "Failed to reset the user's password", err)
		return err
	}

	return nil
}

func (s *Service) ChangePermissions(ctx context.Context, username string, perms uint32) error {
	if !auth.HasPermission(ctx, "users", "edit") {
		return errs.ErrCred
	}

	diff := map[string]interface{}{
		"permissions": perms,
	}

	if _, err := s.repo.Update(ctx, username, diff); err != nil {
		logger.Error("User.Service", "Failed to update the user's email address", err)
		return err
	}

	return nil
}

func (s *Service) SyncSearch(ctx context.Context, action outbox.Event, data []byte) error {
	var payload MessagePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	user, err := s.repo.Get(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil
		}
		return err
	}

	doc := user.ToDocument()
	switch payload.Action {
	case "create", "update":
		if user.SuspendedAt.Valid {
			return nil
		}
		if err := s.se.Index(ctx, &doc); err != nil {
			return err
		}
		return nil
	case "delete":
		if err := s.se.Delete(ctx, doc.Username, *doc.SuspendedAt); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: unkown action", errs.ErrInval)
}
