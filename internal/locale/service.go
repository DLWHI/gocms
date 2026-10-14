package locale

import (
	"context"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/permissions"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, locale *Model) error {
	level := ctx.Value("userperm")

	if level == nil || !permissions.Has("locales", "create", level.(uint32)) {
		return errs.ErrPerm
	}

	if err := s.repo.Create(ctx, locale); err != nil {
		logger.Error("Locale.Service", "Failed to create locale", err)
		return err
	}

	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]Model, error) {
	locales, err := s.repo.GetAll(ctx)
	if err != nil {
		logger.Error("Locale.Service", "Failed to list locales", err)
		return nil, err
	}

	return locales, nil
}

func (s *Service) Delete(ctx context.Context, locale string) error {
	level := ctx.Value("userperm")

	if level == nil || !permissions.Has("locales", "delete", level.(uint32)) {
		return errs.ErrPerm
	}

	if err := s.repo.Delete(ctx, locale); err != nil {
		logger.Error("Locale.Service", "Failed to delete locale", err)
		return err
	}

	return nil
}
