package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/permissions"
	"gocms/internal/search"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const ()

type Service struct {
	repo     Repository
	se       SearchEngine
	sessions SessionManager
}

func NewService(db Repository, se SearchEngine, sm SessionManager) *Service {
	return &Service{
		repo:     db,
		se:       se,
		sessions: sm,
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

func (s *Service) Register(ctx context.Context, user *Model) error {
	level := ctx.Value("userperm")

	if level == nil || !permissions.Has("users", "create", level.(uint32)) {
		clear(user.Passwd)
		return errs.ErrPerm
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
	user.Permissions = permissions.RoleEditor

	err = s.repo.Create(ctx, user)
	clear(hash)
	clear(user.Passwd)
	if err != nil {
		logger.Error("User.Service", "Failed to create user", err)
		return err
	}

	return nil
}

func (s *Service) Login(ctx context.Context, username string, passwd Bytes) (*Session, error) {
	hash, err := s.repo.GetPasswdHash(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrCred
		}
		logger.Error("User.Service", "Failed to retrieve user credentials for login", err)
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(hash, passwd)
	clear(passwd)
	clear(hash)
	if err != nil {
		return nil, errs.ErrCred
	}

	user, err := s.repo.UpdateLogin(ctx, username, time.Now())
	if err != nil {
		logger.Error("User.Service", "Failed to update the user's last login time", err)
		return nil, err
	}

	session, err := s.sessions.Create(ctx, user)
	if err != nil {
		logger.Error("User.Service", "Failed to create a user session", err)
		return nil, err
	}

	return session, nil
}

func (s *Service) Logout(ctx context.Context) error {
	username := ctx.Value("username")
	session := ctx.Value("session")

	if username == nil || session == nil {
		return nil
	}

	if err := s.sessions.Revoke(ctx, username.(string), session.(string)); err != nil {
		logger.Error("User.Service", "Failed to revoke the user session", err)
		return err
	}
	return nil
}

func (s *Service) CurrentUser(ctx context.Context) (*Model, error) {
	username := ctx.Value("username")

	if username == nil {
		return nil, errs.ErrExpired
	}

	user, err := s.repo.Get(ctx, username.(string))
	if err != nil {
		logger.Error("User.Service", "Failed to retrieve the current user", err)
		return nil, err
	}
	clear(user.Passwd)

	return user, nil
}

func (s *Service) Get(ctx context.Context, username string) (*Model, error) {
	user, err := s.repo.Get(ctx, username)
	if err != nil {
		logger.Error("User.Service", "Failed to retrieve user", err)
		return nil, err
	}
	clear(user.Passwd)

	return user, nil
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

func (s *Service) All(ctx context.Context) ([]Model, error) {

	users, err := s.repo.GetAll(ctx)
	if err != nil {
		logger.Error("User.Service", "Failed to list users", err)
		return nil, err
	}

	return users, nil
}

func (s *Service) Suspend(ctx context.Context, username string) error {
	level := ctx.Value("userperm")

	if level == nil || !permissions.Has("users", "delete", level.(uint32)) {
		return errs.ErrPerm
	}
	if err := s.repo.Delete(ctx, username); err != nil {
		logger.Error("User.Service", "Failed to suspend user", err)
		return err
	}

	if err := s.sessions.RevokeUser(ctx, username); err != nil {
		logger.Error("User.Service", "Failed to revoke the user's sessions after suspension", err)
		return err
	}

	return nil
}

func (s *Service) ChangeEmail(ctx context.Context, username string, email string) error {
	initiator := ctx.Value("username")

	if initiator == nil || initiator.(string) != username {
		return errs.ErrPerm
	}

	if err := s.repo.UpdateEmail(ctx, username, email); err != nil {
		logger.Error("User.Service", "Failed to update the user's email address", err)
		return err
	}

	return nil
}

func (s *Service) ChangePasswd(ctx context.Context, username string, oldPwd Bytes, newPwd Bytes) error {
	initiator := ctx.Value("username")

	if initiator == nil || initiator.(string) != username {
		return errs.ErrPerm
	}

	oldHash, err := s.repo.GetPasswdHash(ctx, username)
	if err != nil {
		logger.Error("User.Service", "Failed to retrieve the user's password for password change", err)
		return err
	}

	err = bcrypt.CompareHashAndPassword(oldHash, oldPwd)
	clear(oldPwd)
	clear(oldHash)
	if err != nil {
		return errs.ErrPerm
	}

	newHash, err := bcrypt.GenerateFromPassword(newPwd, 12)
	clear(newPwd)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the new password", err)
		return err
	}

	err = s.repo.UpdatePasswd(ctx, username, newHash)
	clear(newHash)
	if err != nil {
		logger.Error("User.Service", "Failed to update the user's password", err)
		return err
	}

	return nil
}

func (s *Service) ResetPasswd(ctx context.Context, username string, newPwd Bytes) error {
	level := ctx.Value("userperm")

	if level == nil || !permissions.Has("users", "edit", level.(uint32)) {
		return errs.ErrPerm
	}

	newHash, err := bcrypt.GenerateFromPassword(newPwd, 12)
	clear(newPwd)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the replacement password", err)
		return err
	}

	err = s.repo.UpdatePasswd(ctx, username, newHash)
	clear(newHash)
	if err != nil {
		logger.Error("User.Service", "Failed to reset the user's password", err)
		return err
	}

	if err := s.sessions.RevokeUser(ctx, username); err != nil {
		logger.Error("User.Service", "Failed to revoke the user's sessions after password reset", err)
		return err
	}

	return nil
}

func (s *Service) ChangePermissions(ctx context.Context, username string, perms uint32) error {
	level := ctx.Value("userperm").(uint32)

	if !permissions.Has("users", "edit", level) {
		return errs.ErrPerm
	}

	user, err := s.repo.UpdatePermissions(ctx, username, perms)
	if err != nil {
		logger.Error("User.Service", "Failed to update user permissions", err)
		return err
	}

	if _, err := s.sessions.Update(ctx, user); err != nil {
		if !errors.Is(err, errs.ErrExpired) {
			logger.Error("User.Service", "Failed to update the user's active session permissions", err)
			return err
		}
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

	doc := user.ToDoc()
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
