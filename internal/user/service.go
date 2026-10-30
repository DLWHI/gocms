package user

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/password"
	"gocms/internal/search"
)

type Service struct {
	repo      Repository
	se        SearchEngine
	auth      domain.AuthService
	passwords password.PasswordService
}

func NewService(repo Repository, se SearchEngine, auth domain.AuthService, passwords password.PasswordService) *Service {
	return &Service{
		repo:      repo,
		se:        se,
		auth:      auth,
		passwords: passwords,
	}
}

func (s *Service) Create(ctx context.Context, user *domain.User, password password.Password) (string, error) {
	if !s.auth.Authorize(ctx, domain.UserEntityName, domain.ActionCreateName) || !s.auth.IsSubsetOf(ctx, user.Permissions) {
		return "", domain.ErrPerm
	}

	if err := s.passwords.IsValidPassword(password); err != nil {
		return "", err
	}

	var created domain.User
	username, err := domain.NormalizeUsername(user.Username)
	if err != nil {
		return "", err
	}
	created.Username = username

	email, err := domain.NormalizeEmail(user.Email)
	if err != nil {
		return "", err
	}
	created.Email = email

	hash, err := s.passwords.Hash(password)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the user password", err)
		return "", err
	}
	defer clear(hash)

	created.Permissions = user.Permissions
	if created.Permissions == 0 {
		created.Permissions = domain.RoleEditor
	}

	err = s.repo.Create(ctx, &created, hash)
	if err != nil {
		if !errors.Is(err, domain.ErrDuplicate) {
			logger.Error("User.Service", "Failed to create user", err, "data", user)
		}
		return "", err
	}

	return created.Username, nil
}

func (s *Service) Get(ctx context.Context, username string) (*domain.User, error) {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.Get(ctx, username)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to retrieve user", err, "username", username)
		}
		return nil, err
	}

	return user, nil
}

func (s *Service) Search(ctx context.Context, filters *SearchParams) (*SearchResult, error) {
	search.ClampParams(&filters.Params)

	users, err := s.se.Search(ctx, filters)
	if err != nil {
		if !errors.Is(err, domain.ErrInval) {
			logger.Error("User.Service", "Failed to search users", err, "filters", filters)
		}
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
	if !s.auth.Authorize(ctx, domain.UserEntityName, domain.ActionDeleteName) {
		return domain.ErrPerm
	}

	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	if err := s.repo.Suspend(ctx, username); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to suspend user", err, "username", username)
		}
		return err
	}

	return nil
}

func (s *Service) ChangeEmail(ctx context.Context, username string, diff *domain.UserDiff) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	initiator, ok := ctx.Value(domain.UsernameKey).(string)
	if !ok {
		return domain.ErrPerm
	}
	if initiator != username && !s.auth.Authorize(ctx, domain.UserEntityName, domain.ActionEditName) {
		return domain.ErrPerm
	}

	if diff.Email == nil {
		return fmt.Errorf("%w: missing email", domain.ErrInval)
	}
	email, err := domain.NormalizeEmail(*diff.Email)
	if err != nil {
		return err
	}
	diff.Email = &email
	diff.Permissions = nil

	if err := s.repo.Update(ctx, username, diff); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to update the user's email address", err, "value", email)
		}
		return err
	}

	return nil
}

func (s *Service) ChangePermissions(ctx context.Context, username string, diff *domain.UserDiff) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	initiator, ok := ctx.Value(domain.UsernameKey).(string)
	if !ok {
		return domain.ErrPerm
	}

	if initiator == username || !s.auth.Authorize(ctx, domain.UserEntityName, domain.ActionEditName) {
		return domain.ErrPerm
	}

	if diff.Permissions == nil {
		return fmt.Errorf("%w: missing permissions", domain.ErrInval)
	}

	if !s.auth.IsSubsetOf(ctx, *diff.Permissions) {
		return domain.ErrPerm
	}
	diff.Email = nil

	if err := s.repo.Update(ctx, username, diff); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to update the user's permissions", err, "value", *diff.Permissions)
		}
		return err
	}

	return nil
}

func (s *Service) OverridePassword(ctx context.Context, username string, password password.Password) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	initiator, ok := ctx.Value(domain.UsernameKey).(string)
	if !ok {
		return domain.ErrPerm
	}

	if initiator == username || !s.auth.Authorize(ctx, domain.UserEntityName, domain.ActionEditName) {
		return domain.ErrPerm
	}

	if err := s.passwords.IsValidPassword(password); err != nil {
		return err
	}

	pwdhash, err := s.passwords.Hash(password)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the replacement password", err)
		return err
	}
	defer clear(pwdhash)

	err = s.repo.UpdatePassword(ctx, username, pwdhash)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to reset the user's password", err, "username", username)
		}
		return err
	}

	return nil
}
