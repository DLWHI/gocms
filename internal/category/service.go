package category

import (
	"context"
	"fmt"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/permissions"
	"gocms/internal/search"
	"strings"
	"unicode"
)

type Service struct {
	repo   Repository
	lookup SearchEngine
}

func NewService(db Repository, se SearchEngine) *Service {
	return &Service{
		repo:   db,
		lookup: se,
	}
}

func (s *Service) sanitize(ctx context.Context, category *Doc) error {
	category.Slug = strings.TrimSpace(strings.ToLower(category.Slug))
	if category.Slug == "" {
		return fmt.Errorf("%w: slug cannot be empty", errs.ErrInval)
	} else if strings.IndexFunc(category.Slug, unicode.IsSpace) >= 0 {
		return fmt.Errorf("%w: no spaces allowed in slug", errs.ErrInval)
	}

	category.Locale = strings.TrimSpace(strings.ToLower(category.Locale))
	if category.Locale == "" {
		locale := ctx.Value("locale")
		if locale == nil {
			return fmt.Errorf("%w: empty locale", errs.ErrInval)
		}
		category.Locale = locale.(string)
	}

	return nil
}

func (s *Service) Create(ctx context.Context, category *Doc) error {
	level := ctx.Value("userperm")

	if level == nil || !permissions.Has("categories", "create", level.(uint32)) {
		return errs.ErrCred
	}

	if err := s.sanitize(ctx, category); err != nil {
		return err
	}

	if category.Slug == "" {
		return fmt.Errorf("%w: slug cannot be empty", errs.ErrInval)
	}

	model := category.ToModel()
	if err := s.repo.Create(ctx, &model); err != nil {
		logger.Error("Category.Service", "Failed to create category", err)
		return err
	}

	if err := s.lookup.Index(ctx, category); err != nil {
		logger.Error("Category.Service", "Failed to add category to the search index", err)
		return err
	}
	return nil
}

func (s *Service) Edit(ctx context.Context, category *Doc) (*Doc, error) {
	level := ctx.Value("userperm")

	if level == nil || !permissions.Has("categories", "edit", level.(uint32)) {
		return nil, errs.ErrCred
	}

	if err := s.sanitize(ctx, category); err != nil {
		return nil, err
	}

	model := category.ToModel()
	err := s.repo.Update(ctx, &model)
	if err != nil {
		logger.Error("Category.Service", "Failed to update category", err)
		return nil, err
	}

	if err := s.lookup.Index(ctx, category); err != nil {
		logger.Error("Category.Service", "Failed to update category in the search index", err)
		return nil, err
	}
	return category, nil
}

func (s *Service) Get(ctx context.Context, slug string) (*Doc, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%w: empty locale", errs.ErrInval)
	}

	category, err := s.repo.Get(ctx, slug, locale.(string))
	if err != nil {
		logger.Error("Category.Service", "Failed to retrieve category", err)
		return nil, err
	}

	doc := category.ToDoc()
	return &doc, nil
}

func (s *Service) All(ctx context.Context) ([]Doc, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%w: empty locale", errs.ErrInval)
	}

	model, err := s.repo.GetAll(ctx, locale.(string))
	if err != nil {
		logger.Error("Category.Service", "Failed to list categories", err)
		return nil, err
	}

	categories := make([]Doc, 0, len(model))

	for _, category := range model {
		categories = append(categories, category.ToDoc())
	}

	return categories, nil
}

func (s *Service) AllNotHidden(ctx context.Context) ([]Doc, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%w: empty locale", errs.ErrInval)
	}

	model, err := s.repo.GetAll(ctx, locale.(string))
	if err != nil {
		logger.Error("Category.Service", "Failed to list visible categories", err)
		return nil, err
	}

	categories := make([]Doc, 0, len(model))

	for _, category := range model {
		categories = append(categories, category.ToDoc())
	}

	return categories, nil
}

func (s *Service) Search(ctx context.Context, filters *SearchParams) (*SearchResult, error) {
	search.ClampParams(&filters.Params)

	categories, err := s.lookup.Search(ctx, filters)
	if err != nil {
		logger.Error("Category.Service", "Failed to search categories", err)
		return nil, err
	}
	return categories, nil
}

func (s *Service) Delete(ctx context.Context, slug string) error {
	level := ctx.Value("userperm")

	if level == nil || !permissions.Has("categories", "delete", level.(uint32)) {
		return errs.ErrCred
	}

	if err := s.repo.Delete(ctx, slug); err != nil {
		logger.Error("Category.Service", "Failed to delete category", err)
		return err
	}

	if err := s.lookup.Delete(ctx, slug); err != nil {
		logger.Error("Category.Service", "Failed to remove category from the search index", err)
		return err
	}

	return nil
}
