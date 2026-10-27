package user

import (
	"context"
	"errors"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/password"
	"gocms/internal/search"
	"time"
)

type Service struct {
	repo Repository
	se   SearchEngine
	pwd  password.PasswordService
}

func NewService(repo Repository, se SearchEngine, pwd password.PasswordService) *Service {
	return &Service{
		repo: repo,
		se:   se,
		pwd:  pwd,
	}
}

func (s *Service) authorizeEdit(ctx context.Context, username string, action string) error {
	initiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return domain.ErrPerm
	}
	if initiator != username && !auth.HasPermission(ctx, domain.UserEntityName, action) {
		return domain.ErrPerm
	}
	return nil
}

func (s *Service) Create(ctx context.Context, user *domain.User, password password.Password) (string, error) {
	if !auth.HasPermission(ctx, domain.UserEntityName, "create") || !auth.IsSubsetOf(ctx, user.Permissions) {
		return "", domain.ErrPerm
	}

	if err := s.pwd.IsValidPassword(password); err != nil {
		return "", err
	}

	var created domain.User
	username, err := domain.NormalizeUsername(user.Username)
	if err != nil {
		return "", err
	}
	created.Username = username

	if user.Email != "" {
		email, err := domain.NormalizeEmail(user.Email)
		if err != nil {
			return "", err
		}
		created.Email = email
	}

	pwdhash, err := s.pwd.Hash(password)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the user password", err)
		return "", err
	}
	defer clear(pwdhash)

	created.Permissions = user.Permissions
	if created.Permissions == 0 {
		created.Permissions = domain.RoleEditor
	}

	err = s.repo.Create(ctx, &created, pwdhash)
	if err != nil {
		if !errors.Is(err, domain.ErrDuplicate) {
			logger.Error("User.Service", "Failed to create user", err)
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
		if !errors.Is(err, domain.ErrInval) {
			logger.Error("User.Service", "Failed to search users", err)
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
	if !auth.HasPermission(ctx, domain.UserEntityName, "delete") {
		return domain.ErrPerm
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

func (s *Service) PasswordMatches(ctx context.Context, username string, password password.Password) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return domain.ErrInval
	}

	pwdhash, err := s.repo.GetPasswordHash(ctx, username)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to retrieve user's password", err)
			return err
		}
	}
	defer clear(pwdhash)

	if !s.pwd.PasswordMatches(password, pwdhash) {
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

	if err := s.authorizeEdit(ctx, username, "edit"); err != nil {
		return err
	}

	if email != "" {
		email, err = domain.NormalizeEmail(email)
		if err != nil {
			return err
		}
	}

	update := map[string]interface{}{
		"email": email,
	}

	if err := s.repo.Update(ctx, username, update); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to update the user's email address", err)
		}
		return err
	}

	return nil
}

func (s *Service) ChangePermissions(ctx context.Context, username string, permissions uint32) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	initiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return domain.ErrPerm
	}

	if initiator == username || !auth.HasPermission(ctx, domain.UserEntityName, "edit") || !auth.IsSubsetOf(ctx, permissions) {
		return domain.ErrPerm
	}

	update := map[string]interface{}{
		"permissions": permissions,
	}

	if err := s.repo.Update(ctx, username, update); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to update the user's permissions", err)
		}
		return err
	}

	return nil
}

func (s *Service) ChangePassword(ctx context.Context, username string, password password.Password) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	if err := s.authorizeEdit(ctx, username, "edit"); err != nil {
		return err
	}

	if err := s.pwd.IsValidPassword(password); err != nil {
		return err
	}

	pwdhash, err := s.pwd.Hash(password)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the replacement password", err)
		return err
	}
	defer clear(pwdhash)

	err = s.repo.UpdatePassword(ctx, username, pwdhash)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to reset the user's password", err)
		}
		return err
	}

	return nil
}
