package category

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/search"
	"strings"
)

type Service struct {
	repo Repository
	se   SearchEngine
	auth domain.AuthService
}

func NewService(repo Repository, se SearchEngine, auth domain.AuthService) *Service {
	return &Service{
		repo: repo,
		se:   se,
		auth: auth,
	}
}

func (s *Service) Create(ctx context.Context, category *domain.Category) (string, error) {
	if !s.auth.Authorize(ctx, domain.CategoryEntityName, domain.ActionCreateName) {
		return "", domain.ErrCred
	}

	var created domain.Category
	created = *category
	slug, err := domain.NormalizeSlug(category.Slug)
	if err != nil {
		return "", err
	}
	created.Slug = slug

	normalized := strings.TrimSpace(strings.ToLower(category.Locale))
	if normalized == "" {
		locale, ok := ctx.Value(domain.LocaleKey).(string)
		if !ok {
			return "", fmt.Errorf("%w: empty locale", domain.ErrInval)
		}
		created.Locale = locale
	} else {
		created.Locale = normalized
	}

	if err := s.repo.Create(ctx, &created); err != nil {
		logger.Error("Category.Service", "Failed to create category", err)
		return "", err
	}

	return created.Slug, nil
}

func (s *Service) Edit(ctx context.Context, slug string, diff *domain.CategoryDiff) (string, error) {
	if !s.auth.Authorize(ctx, domain.CategoryEntityName, domain.ActionEditName) {
		return "", domain.ErrCred
	}

	slug, err := domain.NormalizeSlug(slug)
	if err != nil {
		return "", err
	}

	locale, ok := ctx.Value(domain.LocaleKey).(string)
	if !ok {
		return "", fmt.Errorf("%w: empty locale", domain.ErrInval)
	}

	if diff.Slug != nil {
		normalized, err := domain.NormalizeSlug(*diff.Slug)
		if err != nil {
			return "", err
		}
		diff.Slug = &normalized
	}

	if err := s.repo.Update(ctx, slug, locale, diff); err != nil {
		logger.Error("Category.Service", "Failed to update category", err)
		return "", err
	}

	if diff.Slug != nil {
		return *diff.Slug, nil
	} else {
		return slug, nil
	}
}

func (s *Service) Get(ctx context.Context, slug string) (*domain.Category, error) {
	locale, ok := ctx.Value(domain.LocaleKey).(string)
	if !ok {
		return nil, fmt.Errorf("%w: empty locale", domain.ErrInval)
	}

	slug, err := domain.NormalizeSlug(slug)
	if err != nil {
		return nil, err
	}

	category, err := s.repo.Get(ctx, slug, locale)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Category.Service", "Failed to retrieve category", err)
		}
		return nil, err
	}

	return category, nil
}

func (s *Service) All(ctx context.Context) ([]domain.Category, error) {
	locale, ok := ctx.Value(domain.LocaleKey).(string)
	if !ok {
		return nil, fmt.Errorf("%w: empty locale", domain.ErrInval)
	}

	categories, err := s.repo.GetAll(ctx, locale)
	if err != nil {
		logger.Error("Category.Service", "Failed to list categories", err)
		return nil, err
	}

	return categories, nil
}

func (s *Service) AllNotHidden(ctx context.Context) ([]domain.Category, error) {
	locale, ok := ctx.Value(domain.LocaleKey).(string)
	if !ok {
		return nil, fmt.Errorf("%w: empty locale", domain.ErrInval)
	}

	categories, err := s.repo.GetAll(ctx, locale)
	if err != nil {
		logger.Error("Category.Service", "Failed to list categories", err)
		return nil, err
	}

	i := 0
	for _, category := range categories {
		if category.Hidden {
			categories[i] = category
			i++
		}
	}
	categories = categories[:i]

	return categories, nil
}

func (s *Service) Search(ctx context.Context, filters *SearchParams) (*SearchResult, error) {
	search.ClampParams(&filters.Params)

	categories, err := s.se.Search(ctx, filters)
	if err != nil {
		logger.Error("Category.Service", "Failed to search categories", err)
		return nil, err
	}
	return categories, nil
}

func (s *Service) Delete(ctx context.Context, slug string) error {
	if !s.auth.Authorize(ctx, domain.CategoryEntityName, domain.ActionDeleteName) {
		return domain.ErrCred
	}

	slug, err := domain.NormalizeSlug(slug)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, slug); err != nil {
		logger.Error("Category.Service", "Failed to delete category", err)
		return err
	}

	return nil
}
