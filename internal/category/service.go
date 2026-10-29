package category

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/search"
	"strings"
)

type Service struct {
	repo Repository
	se   SearchEngine
}

func NewService(db Repository, se SearchEngine) *Service {
	return &Service{
		repo: db,
		se:   se,
	}
}

func (s *Service) resolveLocale(ctx context.Context, value string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		locale, ok := ctx.Value("locale").(string)
		if !ok {
			return "", fmt.Errorf("%w: empty locale", domain.ErrInval)
		}
		return locale, nil
	}
	return normalized, nil
}

func (s *Service) Create(ctx context.Context, category *domain.Category) (string, error) {
	if !auth.HasPermission(ctx, "categories", "create") {
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
		locale, ok := ctx.Value("locale").(string)
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
	if !auth.HasPermission(ctx, "categories", "edit") {
		return "", domain.ErrCred
	}

	slug, err := domain.NormalizeSlug(slug)
	if err != nil {
		return "", err
	}

	locale, ok := ctx.Value("locale").(string)
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
	locale, ok := ctx.Value("locale").(string)
	if !ok {
		return nil, fmt.Errorf("%w: empty locale", domain.ErrInval)
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

func (s *Service) AllNotHidden(ctx context.Context) ([]domain.Category, error) {
	locale, ok := ctx.Value("locale").(string)
	if !ok {
		return nil, fmt.Errorf("%w: empty locale", domain.ErrInval)
	}

	categories, err := s.repo.GetNotHidden(ctx, locale)
	if err != nil {
		logger.Error("Category.Service", "Failed to list categories", err)
		return nil, err
	}

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
	if !auth.HasPermission(ctx, "categories", "delete") {
		return domain.ErrCred
	}

	if err := s.repo.Delete(ctx, slug); err != nil {
		logger.Error("Category.Service", "Failed to delete category", err)
		return err
	}

	return nil
}
