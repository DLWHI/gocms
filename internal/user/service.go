package user

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/search"
	"net/mail"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	repo Repository
	se   SearchEngine
}

func NewService(repo Repository, se SearchEngine) *Service {
	return &Service{
		repo: repo,
		se:   se,
	}
}

func (s *Service) Create(ctx context.Context, user *Record) error {
	if !auth.HasPermission(ctx, "users", "create") {
		clear(user.Passwd)
		return errs.ErrCred
	}

	if len(user.Passwd) == 0 || len(user.Passwd) > 72 {
		clear(user.Passwd)
		return fmt.Errorf("%w: invalid password", errs.ErrInval)
	}

	username, err := domain.NormalizeUsername(user.Username)
	if err != nil {
		return err
	}
	user.Username = username

	parsed, err := mail.ParseAddress(user.Email)
	if err != nil || parsed.Name != "" {
		clear(user.Passwd)
		return fmt.Errorf("%w: invalid email", errs.ErrInval)
	}
	user.Email = parsed.Address

	hash, err := bcrypt.GenerateFromPassword(user.Passwd, 12)
	clear(user.Passwd)
	if err != nil {
		clear(hash)
		logger.Error("User.Service", "Failed to hash the user password", err)
		return err
	}

	user.Passwd = hash
	user.Permissions = auth.RoleEditor
	user.LastLogin = nil
	user.SuspendedAt = gorm.DeletedAt{Valid: false}

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
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.UpdateLogin(ctx, username, time.Now())
	if err != nil {
		logger.Error("User.Service", "Failed to update the user's last login time", err)
		return nil, err
	}

	val := user.ToDomain()
	return &val, nil
}

func (s *Service) Get(ctx context.Context, username string) (*Document, error) {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.Get(ctx, username)
	if err != nil {
		if !errors.Is(err, errs.ErrNotFound) {
			logger.Error("User.Service", "Failed to retrieve user", err)
		}
		return nil, err
	}

	val := user.ToDomain()
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

func (s *Service) All(ctx context.Context) ([]Document, error) {

	users, err := s.repo.GetAll(ctx)
	if err != nil {
		logger.Error("User.Service", "Failed to list users", err)
		return nil, err
	}

	vals := make([]Document, 0, len(users))

	for _, user := range users {
		vals = append(vals, user.ToDomain())
	}

	return vals, nil
}

func (s *Service) Suspend(ctx context.Context, username string) error {
	if !auth.HasPermission(ctx, "users", "delete") {
		return errs.ErrCred
	}

	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, username); err != nil {
		logger.Error("User.Service", "Failed to suspend user", err)
		return err
	}

	return nil
}

func (s *Service) PasswordMatches(ctx context.Context, username string, passwd Password) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return errs.ErrCred
	}

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
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	initiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return errs.ErrCred
	}

	if initiator != username && !auth.HasPermission(ctx, "users", "edit") {
		return errs.ErrCred
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Name != "" {
		return fmt.Errorf("%w: invalid email", errs.ErrInval)
	}

	diff := map[string]interface{}{
		"email": parsed.Address,
	}

	if err := s.repo.Update(ctx, username, diff); err != nil {
		logger.Error("User.Service", "Failed to update the user's email address", err)
		return err
	}

	return nil
}

func (s *Service) ChangePasswd(ctx context.Context, username string, passwd domain.Password) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	initiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return errs.ErrCred
	}

	if initiator != username && !auth.HasPermission(ctx, "users", "edit") {
		return errs.ErrCred
	}

	if len(passwd) == 0 || len(passwd) > 72 {
		clear(passwd)
		return fmt.Errorf("%w: invalid password", errs.ErrInval)
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
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	intitiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return errs.ErrCred
	}

	if intitiator == username || !auth.HasPermission(ctx, "users", "edit") {
		return errs.ErrCred
	}

	diff := map[string]interface{}{
		"permissions": perms,
	}

	if err := s.repo.Update(ctx, username, diff); err != nil {
		logger.Error("User.Service", "Failed to update the user's permissions", err)
		return err
	}

	return nil
}
