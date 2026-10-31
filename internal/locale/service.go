package locale

import (
	"context"
	"errors"
	"gocms/internal/domain"
	"gocms/internal/logger"
)

const MaxCodeLength = 2

type Service struct {
	repo Repository
	auth domain.AuthService
}

func NewService(repo Repository, auth domain.AuthService) *Service {
	return &Service{
		repo: repo,
		auth: auth,
	}
}

func (s *Service) Create(ctx context.Context, locale *domain.Locale) error {
	if !s.auth.Authorize(ctx, domain.LocaleEntityName, domain.ActionCreateName) {
		return domain.ErrCred
	}

	created := *locale

	name, err := domain.NormalizeName(locale.Name)
	if err != nil {
		return err
	}
	created.Name = name

	code, err := domain.NormalizeLocaleCode(locale.Code)
	if err != nil {
		return err
	}
	created.Code = code

	if err := s.repo.Create(ctx, &created); err != nil {
		if !errors.Is(err, domain.ErrDuplicate) {
			logger.Error("Locale.Service", "Failed to create locale", err)
		}
		return err
	}

	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]domain.Locale, error) {
	locales, err := s.repo.GetAll(ctx)
	if err != nil {
		logger.Error("Locale.Service", "Failed to list locales", err)
		return nil, err
	}

	return locales, nil
}

func (s *Service) Delete(ctx context.Context, code string) error {
	if !s.auth.Authorize(ctx, domain.LocaleEntityName, domain.ActionDeleteName) {
		return domain.ErrCred
	}

	if err := s.repo.Delete(ctx, code); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Locale.Service", "Failed to delete locale", err)
		}
		return err
	}

	return nil
}
