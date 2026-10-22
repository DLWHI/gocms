package user

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/hash"
	"gocms/internal/logger"
	"gocms/internal/search"
	"time"
)

type Service struct {
	repo Repository
	se   SearchEngine
	pwd  hash.PasswordService
}

func NewService(repo Repository, se SearchEngine, pwd hash.PasswordService) *Service {
	return &Service{
		repo: repo,
		se:   se,
		pwd:  pwd,
	}
}

func (s *Service) Create(ctx context.Context, user *domain.User, passwd hash.Password) (string, error) {
	level, ok := ctx.Value(auth.UserPermissionKey).(uint32)
	if !ok {
		return "", domain.ErrCred
	}

	if !auth.HasPermission(ctx, domain.UserEntityName, "create") || (level&user.Permissions) != user.Permissions {
		return "", domain.ErrCred
	}

	if !s.pwd.IsValidPassword(passwd) {
		return "", fmt.Errorf("%w: invalid password", domain.ErrInval)
	}

	username, err := domain.NormalizeUsername(user.Username)
	if err != nil {
		return "", err
	}
	user.Username = username

	if user.Email != "" {
		email, err := domain.ParseEmail(user.Email)
		if err != nil {
			return "", err
		}
		user.Email = email
	}

	pwdhash, err := s.pwd.Hash(passwd)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the user password", err)
		return "", err
	}
	defer clear(pwdhash)

	if user.Permissions == 0 {
		user.Permissions = domain.RoleEditor
	}

	err = s.repo.Create(ctx, user, pwdhash)
	if err != nil {
		logger.Error("User.Service", "Failed to create user", err)
		return "", err
	}

	return user.Username, nil
}

func (s *Service) Get(ctx context.Context, username string) (*domain.User, error) {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.Get(ctx, username)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to retrieve user", err)
		}
		return nil, err
	}

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

func (s *Service) AllActive(ctx context.Context) ([]domain.User, error) {
	users, err := s.repo.GetActive(ctx)
	if err != nil {
		logger.Error("User.Service", "Failed to list users", err)
		return nil, err
	}

	return users, nil
}

func (s *Service) Suspend(ctx context.Context, username string) error {
	if !auth.HasPermission(ctx, domain.UserEntityName, "delete") {
		return domain.ErrCred
	}

	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	if err := s.repo.Suspend(ctx, username); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to suspend user", err)
		}
		return err
	}

	return nil
}

func (s *Service) PasswordMatches(ctx context.Context, username string, passwd hash.Password) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return domain.ErrCred
	}

	pwdhash, err := s.repo.GetPasswdHash(ctx, username)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to retrieve user's password", err)
		}
		return err
	}
	defer clear(pwdhash)

	if !s.pwd.PasswordMatches(passwd, pwdhash) {
		return domain.ErrCred
	}
	return nil
}

func (s *Service) UpdateLogin(ctx context.Context, username string) (*domain.User, error) {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.UpdateLogin(ctx, username, time.Now())
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to update the user's last login time", err)
		}
		return nil, err
	}

	return user, nil
}

func (s *Service) ChangeEmail(ctx context.Context, username string, email string) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	initiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return domain.ErrCred
	}

	if initiator != username && !auth.HasPermission(ctx, domain.UserEntityName, "edit") {
		return domain.ErrCred
	}

	if email != "" {
		email, err = domain.ParseEmail(email)
		if err != nil {
			return err
		}
	}

	update := domain.User{
		Email: email,
	}

	if err := s.repo.Update(ctx, username, &update); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to update the user's email address", err)
		}
		return err
	}

	return nil
}

func (s *Service) ChangePasswd(ctx context.Context, username string, passwd hash.Password) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	initiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return domain.ErrCred
	}

	if initiator != username && !auth.HasPermission(ctx, domain.UserEntityName, "edit") {
		return domain.ErrCred
	}

	if !s.pwd.IsValidPassword(passwd) {
		return fmt.Errorf("%w: invalid password", domain.ErrInval)
	}

	pwdhash, err := s.pwd.Hash(passwd)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the replacement password", err)
		return err
	}
	defer clear(pwdhash)

	err = s.repo.UpdatePasswd(ctx, username, pwdhash)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to reset the user's password", err)
		}
		return err
	}

	return nil
}

func (s *Service) ChangePermissions(ctx context.Context, username string, perms uint32) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	level, ok := ctx.Value(auth.UserPermissionKey).(uint32)
	if !ok {
		return domain.ErrCred
	}
	initiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return domain.ErrCred
	}

	if initiator == username || !auth.HasPermission(ctx, domain.UserEntityName, "edit") || (level&perms) != perms {
		return domain.ErrCred
	}

	update := domain.User{
		Permissions: perms,
	}

	if err := s.repo.Update(ctx, username, &update); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to update the user's permissions", err)
		}
		return err
	}

	return nil
}
