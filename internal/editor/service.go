package editor

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/search"
)

type Service struct {
	repo Repository
	se   SearchEngine
	auth domain.AuthService
}

func NewService(db Repository, se SearchEngine, auth domain.AuthService) *Service {
	return &Service{
		repo: db,
		se:   se,
		auth: auth,
	}
}

func (s *Service) Create(ctx context.Context, editor *domain.Editor) (string, error) {
	if !s.auth.Authorize(ctx, domain.EditorEntityName, domain.ActionCreateName) {
		return "", domain.ErrPerm
	}

	created := *editor
	slug, err := domain.NormalizeSlug(editor.Slug)
	if err != nil {
		return "", err
	}
	created.Slug = slug

	locale, err := domain.ResolveLocale(ctx, editor.Locale)
	if err != nil {
		return "", err
	}
	created.Locale = locale

	name, err := domain.NormalizeName(editor.Name)
	if err != nil {
		return "", err
	}
	created.Name = name

	if err := s.repo.Create(ctx, &created); err != nil {
		if !errors.Is(err, domain.ErrDuplicate) {
			logger.Error("Editor.Service", "Failed to create editor profile", err)
		}
		return "", err
	}

	return created.Slug, nil
}

func (s *Service) Edit(ctx context.Context, slug string, diff *domain.EditorDiff) (string, error) {
	if !s.auth.Authorize(ctx, domain.EditorEntityName, domain.ActionEditName) {
		return "", domain.ErrPerm
	}

	slug, err := domain.NormalizeSlug(slug)
	if err != nil {
		return "", err
	}

	locale, err := domain.ResolveLocale(ctx, diff.Locale)
	if err != nil {
		return "", err
	}

	if diff.Slug != nil {
		normalized, err := domain.NormalizeSlug(*diff.Slug)
		if err != nil {
			return "", err
		}
		diff.Slug = &normalized
	}

	if diff.Name != nil {
		normalized, err := domain.NormalizeName(*diff.Name)
		if err != nil {
			return "", err
		}
		diff.Name = &normalized
	}

	if err := s.repo.Update(ctx, slug, locale, diff); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Editor.Service", "Failed to update editor profile", err)
		}
		return "", err
	}

	if diff.Slug != nil {
		return *diff.Slug, nil
	} else {
		return slug, nil
	}
}

func (s *Service) Get(ctx context.Context, slug string) (*domain.Editor, error) {
	locale, ok := ctx.Value(domain.LocaleKey).(string)
	if !ok {
		return nil, fmt.Errorf("%w: empty locale", domain.ErrInval)
	}

	slug, err := domain.NormalizeSlug(slug)
	if err != nil {
		return nil, err
	}

	editor, err := s.repo.GetAny(ctx, slug, locale)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Editor.Service", "Failed to retrieve editor profile", err)
		}
		return nil, err
	}

	return editor, nil
}

func (s *Service) All(ctx context.Context) ([]domain.Editor, error) {
	locale, ok := ctx.Value(domain.LocaleKey).(string)
	if !ok {
		return nil, fmt.Errorf("%w: empty locale", domain.ErrInval)
	}

	editors, err := s.repo.GetAll(ctx, locale)
	if err != nil {
		logger.Error("Editor.Service", "Failed to list editor profiles", err)
		return nil, err
	}

	return editors, nil
}

func (s *Service) Search(ctx context.Context, filters *SearchParams) (*SearchResult, error) {
	search.ClampParams(&filters.Params)

	locale, err := domain.ResolveLocale(ctx, filters.Locale)
	if err != nil {
		return nil, err
	}
	filters.Locale = locale

	editors, err := s.se.Search(ctx, filters)
	if err != nil {
		logger.Error("Editor.Service", "Failed to search editor profiles", err)
		return nil, err
	}

	return editors, nil
}

func (s *Service) Delete(ctx context.Context, slug string) error {
	if !s.auth.Authorize(ctx, domain.EditorEntityName, domain.ActionDeleteName) {
		return domain.ErrPerm
	}

	slug, err := domain.NormalizeSlug(slug)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, slug); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Editor.Service", "Failed to delete editor profile", err)
		}
		return err
	}

	return nil
}
