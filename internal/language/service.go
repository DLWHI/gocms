package language

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

func (s *Service) Create(ctx context.Context, lang *domain.Language) error {
	if !s.auth.Authorize(ctx, domain.LangEntityName, domain.ActionCreateName) {
		return domain.ErrCred
	}

	created := *lang

	name, err := domain.NormalizeName(lang.Name)
	if err != nil {
		return err
	}
	created.Name = name

	code, err := domain.NormalizeLangCode(lang.Code)
	if err != nil {
		return err
	}
	created.Code = code

	if err := s.repo.Create(ctx, &created); err != nil {
		if !errors.Is(err, domain.ErrDuplicate) {
			logger.Error("Language.Service", "Failed to create language", err)
		}
		return err
	}

	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]domain.Language, error) {
	langs, err := s.repo.GetAll(ctx)
	if err != nil {
		logger.Error("Language.Service", "Failed to list languages", err)
		return nil, err
	}

	return langs, nil
}

func (s *Service) Delete(ctx context.Context, code string) error {
	if !s.auth.Authorize(ctx, domain.LangEntityName, domain.ActionDeleteName) {
		return domain.ErrCred
	}

	code, err := domain.NormalizeLangCode(code)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, code); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Language.Service", "Failed to delete language", err)
		}
		return err
	}

	return nil
}
