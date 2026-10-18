package auth

import (
	"context"
	"errors"
	"gocms/internal/domain"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/session"
)

type Service struct {
	users    UserManager
	sessions session.Repository
}

func NewService(users UserManager, sessions session.Repository) *Service {
	return &Service{
		users:    users,
		sessions: sessions,
	}
}

func (s *Service) Login(ctx context.Context, username string, passwd domain.Password) (*session.Session, error) {
	if err := s.users.PasswordMatches(ctx, username, passwd); err != nil {
		if !errors.Is(err, errs.ErrCred) {
			logger.Error("User.Service", "Failed to retrieve user credentials for login", err)
		}
		return nil, err
	}

	user, err := s.users.UpdateLogin(ctx, username)
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
	session, ok := ctx.Value("username").(string)
	if !ok {
		return nil
	}

	if err := s.sessions.RevokeID(ctx, session); err != nil {
		logger.Error("User.Service", "Failed to revoke the user session", err)
		return err
	}
	return nil
}

func (s *Service) UpdateProfile(ctx context.Context) (*session.Session, error) {
	session, ok := ctx.Value("username").(string)
	if !ok {
		return nil, errs.ErrExpired
	}

	user, err := s.sessions.Validate(ctx, session)
	if err != nil {
		logger.Error("User.Service", "Failed to retrieve the current user", err)
		return nil, err
	}

	return user, nil
}

func (s *Service) CurrentUser(ctx context.Context) (*session.Session, error) {
	session, ok := ctx.Value("username").(string)
	if !ok {
		return nil, errs.ErrExpired
	}

	user, err := s.sessions.Validate(ctx, session)
	if err != nil {
		logger.Error("User.Service", "Failed to retrieve the current user", err)
		return nil, err
	}

	return user, nil
}

func (s *Service) ChangePasswd(ctx context.Context, username string, oldPwd domain.Password, newPwd domain.Password) error {
	initiator := ctx.Value("username")

	if initiator == nil || initiator.(string) != username {
		return errs.ErrPerm
	}

	err := s.users.PasswordMatches(ctx, username, oldPwd)
	clear(oldPwd)
	if err != nil {
		logger.Error("User.Service", "Failed to retrieve the user's password for password change", err)
		return err
	}

	err = s.users.ChangePasswd(ctx, username, newPwd)
	clear(newPwd)
	if err != nil {
		logger.Error("User.Service", "Failed to update the user's password", err)
		return err
	}

	return nil
}
