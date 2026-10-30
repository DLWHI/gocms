package auth

import (
	"context"
	"errors"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/password"
	"time"
)

type Service struct {
	users     UserRepository
	passwords password.PasswordService
	sessions  SessionRepository
}

func NewService(sessions SessionRepository, users UserRepository, passwords password.PasswordService) *Service {
	return &Service{
		users:     users,
		sessions:  sessions,
		passwords: passwords,
	}
}

func (s *Service) PasswordMatches(ctx context.Context, username string, passwd password.Password) error {
	hash, err := s.users.GetPasswordHash(ctx, username)
	defer clear(hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrCred
		}
		logger.Error("Auth.Service", "Failed to retrieve user credentials for login", err)
		return err
	}

	if !s.passwords.PasswordMatches(passwd, hash) {
		return domain.ErrCred
	}
	return nil
}

func (s *Service) Login(ctx context.Context, username string, passwd password.Password) (string, error) {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return "", err
	}

	err = s.PasswordMatches(ctx, username, passwd)
	clear(passwd)
	if err != nil {
		return "", err
	}

	if err := s.users.UpdateLogin(ctx, username, time.Now()); err != nil {
		logger.Error("Auth.Service", "Failed to update the user's last login time", err)
		return "", err
	}

	token, err := s.sessions.Create(ctx, username)
	if err != nil {
		logger.Error("Auth.Service", "Failed to create a user session", err)
		return "", err
	}

	return token, nil
}

func (s *Service) Logout(ctx context.Context) error {
	username, ok := ctx.Value(domain.UsernameKey).(string)
	if !ok {
		return nil
	}
	session, ok := ctx.Value(domain.SessionIDKey).(string)
	if !ok {
		return nil
	}

	if err := s.sessions.Revoke(ctx, username, session); err != nil {
		logger.Error("Auth.Service", "Failed to revoke the user session", err)
		return err
	}
	return nil
}

func (s *Service) CurrentUser(ctx context.Context) (*domain.User, error) {
	session, ok := ctx.Value(domain.SessionIDKey).(string)
	if !ok {
		return nil, domain.ErrExpired
	}

	username, _, err := s.sessions.Validate(ctx, session)
	if err != nil {
		logger.Error("Auth.Service", "Failed to retrieve the current user", err)
		return nil, err
	}

	user, err := s.users.Get(ctx, username)
	if err != nil {
		logger.Error("Auth.Service", "Failed to retrieve the current user", err)
		return nil, err
	}

	return user, nil
}

func (s *Service) ChangePasswd(ctx context.Context, old password.Password, new password.Password) error {
	initiator, ok := ctx.Value(domain.UsernameKey).(string)
	if !ok {
		return domain.ErrCred
	}

	if err := s.PasswordMatches(ctx, initiator, old); err != nil {
		return err
	}
	clear(old)

	if err := s.passwords.IsValidPassword(new); err != nil {
		return err
	}

	hash, err := s.passwords.Hash(new)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the replacement password", err)
		return err
	}
	defer clear(hash)

	err = s.users.UpdatePassword(ctx, initiator, hash)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("User.Service", "Failed to reset the user's password", err, "username", initiator)
		}
		return err
	}

	return nil
}

func (s *Service) Authorize(ctx context.Context, entity string, action string) bool {
	username, ok := ctx.Value(domain.UsernameKey).(string)
	if !ok {
		return false
	}

	user, err := s.users.Get(ctx, username)
	if err != nil {
		logger.Error("Auth.Service", "Failed to retrieve user data to check permissions", err)
	}

	return domain.HasPermission(user.Permissions, entity, action)
}

func (s *Service) IsSubsetOf(ctx context.Context, permissions uint32) bool {
	username, ok := ctx.Value(domain.UsernameKey).(string)
	if !ok {
		return false
	}

	user, err := s.users.Get(ctx, username)
	if err != nil {
		logger.Error("Auth.Service", "Failed to retrieve user data to check permissions", err)
	}
	return (user.Permissions & permissions) == permissions
}
