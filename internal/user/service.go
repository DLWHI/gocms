package user

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/search"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	repo       Repository
	se         SearchEngine
	bcryptCost int
	dummyHash  []byte
}

func NewService(repo Repository, se SearchEngine) *Service {
	pwd, err := domain.GeneratePassword(16)
	if err != nil {
		logger.Error("User.Service", "Failed to generate dummy password", err)
		return nil
	}
	dummyHash, err := bcrypt.GenerateFromPassword(pwd, 12)
	clear(pwd)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the dummy password", err)
		return nil
	}
	return &Service{
		repo:       repo,
		se:         se,
		bcryptCost: 12,
		dummyHash:  dummyHash,
	}
}

func (s *Service) Create(ctx context.Context, user *Record) error {
	if !auth.HasPermission(ctx, "users", "create") {
		clear(user.Passwd)
		return domain.ErrCred
	}

	if !domain.IsValidPassword(user.Passwd) {
		clear(user.Passwd)
		return fmt.Errorf("%w: invalid password", domain.ErrInval)
	}

	username, err := domain.NormalizeUsername(user.Username)
	if err != nil {
		clear(user.Passwd)
		return err
	}
	user.Username = username

	user.Email, err = domain.ParseEmail(user.Email)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword(user.Passwd, 12)
	clear(user.Passwd)
	if err != nil {
		logger.Error("User.Service", "Failed to hash the user password", err)
		return err
	}

	user.Passwd = hash
	user.Permissions = domain.RoleEditor
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
		if !errors.Is(err, domain.ErrNotFound) {
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

func (s *Service) AllActive(ctx context.Context) ([]Document, error) {

	users, err := s.repo.GetActive(ctx)
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
	if !auth.HasPermission(ctx, EntityName, "delete") {
		return domain.ErrCred
	}

	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return err
	}

	if err := s.repo.Suspend(ctx, username); err != nil {
		logger.Error("User.Service", "Failed to suspend user", err)
		return err
	}

	return nil
}

func (s *Service) PasswordMatches(ctx context.Context, username string, passwd Password) error {
	username, err := domain.NormalizeUsername(username)
	if err != nil {
		return domain.ErrCred
	}

	hash, err := s.repo.GetPasswdHash(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			hash = s.dummyHash
		} else {
			logger.Error("User.Service", "Failed to retrieve user's password", err)
			return err
		}
	}

	err = bcrypt.CompareHashAndPassword(hash, passwd)
	clear(hash)
	if err != nil {
		return domain.ErrCred
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
		return domain.ErrCred
	}

	if initiator != username && !auth.HasPermission(ctx, EntityName, "edit") {
		return domain.ErrCred
	}

	email, err = domain.ParseEmail(email)
	if err != nil {
		return err
	}

	diff := map[string]interface{}{
		"email": email,
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
		return domain.ErrCred
	}

	if initiator != username && !auth.HasPermission(ctx, EntityName, "edit") {
		return domain.ErrCred
	}

	if !domain.IsValidPassword(passwd) {
		clear(passwd)
		return fmt.Errorf("%w: invalid password", domain.ErrInval)
	}

	hash, err := bcrypt.GenerateFromPassword(passwd, s.bcryptCost)
	clear(passwd)
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

	initiator, ok := ctx.Value(auth.UserNameKey).(string)
	if !ok {
		return domain.ErrCred
	}

	if initiator == username || !auth.HasPermission(ctx, EntityName, "edit") {
		return domain.ErrCred
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
