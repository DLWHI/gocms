package category

import (
	"context"
	"errors"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/lang"
	"gocms/internal/logger"
	"gocms/internal/search"
)

type Service struct {
	repo Repository
	se   SearchEngine
	auth auth.Service
}

func NewService(repo Repository, se SearchEngine, auth auth.Service) *Service {
	return &Service{
		repo: repo,
		se:   se,
		auth: auth,
	}
}

func (s *Service) Create(ctx context.Context, category *domain.Category) error {
	if !s.auth.Authorize(ctx, domain.EntityName(&domain.Category{}), domain.ActionCreate) {
		return domain.ErrPerm
	}

	slug, err := domain.NormalizeSlug(category.Slug)
	if err != nil {
		return err
	}
	category.Slug = slug

	language, err := lang.ResolveLang(ctx)
	if err != nil {
		return err
	}
	category.Lang = language

	name, err := domain.NormalizeName(category.Name)
	if err != nil {
		return err
	}
	category.Name = name

	if err := s.repo.Create(ctx, category); err != nil {
		if !errors.Is(err, domain.ErrDuplicate) && !errors.Is(err, domain.ErrInval) {
			logger.Error("Category.Service", "Failed to create category", err)
		}
		return err
	}

	return nil
}

func (s *Service) Edit(ctx context.Context, slug string, diff *domain.CategoryDiff) (string, error) {
	if !s.auth.Authorize(ctx, domain.EntityName(&domain.Category{}), domain.ActionEdit) {
		return "", domain.ErrPerm
	}

	slug, err := domain.NormalizeSlug(slug)
	if err != nil {
		return "", err
	}

	language, err := lang.ResolveLang(ctx)
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

	if err := s.repo.Update(ctx, slug, language, diff); err != nil {
		if !errors.Is(err, domain.ErrNotFound) && !errors.Is(err, domain.ErrInval) && !errors.Is(err, domain.ErrDuplicate) {
			logger.Error("Category.Service", "Failed to update category", err)
		}
		return "", err
	}

	if diff.Slug != nil {
		return *diff.Slug, nil
	} else {
		return slug, nil
	}
}

func (s *Service) Get(ctx context.Context, slug string) (*domain.Category, error) {
	language, err := lang.ResolveLang(ctx)
	if err != nil {
		return nil, err
	}

	slug, err = domain.NormalizeSlug(slug)
	if err != nil {
		return nil, err
	}

	category, err := s.repo.Get(ctx, slug, language)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Category.Service", "Failed to retrieve category", err)
		}
		return nil, err
	}

	return category, nil
}

func (s *Service) All(ctx context.Context) ([]domain.Category, error) {
	language, err := lang.ResolveLang(ctx)
	if err != nil {
		return nil, err
	}

	categories, err := s.repo.GetAll(ctx, language)
	if err != nil {
		logger.Error("Category.Service", "Failed to list categories", err)
		return nil, err
	}

	return categories, nil
}

func (s *Service) AllNotHidden(ctx context.Context) ([]domain.Category, error) {
	language, err := lang.ResolveLang(ctx)
	if err != nil {
		return nil, err
	}

	categories, err := s.repo.GetAll(ctx, language)
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

	language, err := lang.ResolveLang(ctx)
	if err != nil {
		return nil, err
	}

	categories, err := s.se.SearchLang(ctx, language, filters)
	if err != nil {
		if !errors.Is(err, domain.ErrInval) {
			logger.Error("Category.Service", "Failed to search categories", err)
		}
		return nil, err
	}
	return categories, nil
}

func (s *Service) Delete(ctx context.Context, slug string) error {
	if !s.auth.Authorize(ctx, domain.EntityName(&domain.Category{}), domain.ActionDelete) {
		return domain.ErrPerm
	}

	slug, err := domain.NormalizeSlug(slug)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, slug); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Category.Service", "Failed to delete category", err)
		}
		return err
	}

	return nil
}
