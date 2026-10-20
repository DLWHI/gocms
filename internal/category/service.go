package category

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/search"
	"strings"
	"unicode"
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

func (s *Service) sanitize(ctx context.Context, category *Document) error {
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

func (s *Service) Create(ctx context.Context, category *Document) error {
	if !auth.HasPermission(ctx, "categories", "create") {
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

	return nil
}

func (s *Service) Edit(ctx context.Context, category *Document) (*Document, error) {
	if !auth.HasPermission(ctx, "categories", "edit") {
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

	return category, nil
}

func (s *Service) Get(ctx context.Context, slug string) (*Document, error) {
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

func (s *Service) All(ctx context.Context, hidden *bool) ([]Document, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%w: empty locale", errs.ErrInval)
	}

	model, err := s.repo.GetAll(ctx, locale.(string), hidden)
	if err != nil {
		logger.Error("Category.Service", "Failed to list categories", err)
		return nil, err
	}

	categories := make([]Document, 0, len(model))

	for _, category := range model {
		categories = append(categories, category.ToDoc())
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
		return errs.ErrCred
	}

	if err := s.repo.Delete(ctx, slug); err != nil {
		logger.Error("Category.Service", "Failed to delete category", err)
		return err
	}

	return nil
}

func (s *Service) SyncSearch(ctx context.Context, data []byte) error {
	var payload MessagePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch payload.Action {
	case "create", "update":
		category, err := s.repo.Get(ctx, payload.Slug, payload.Locale)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return nil
			}
			return err
		}
		doc := category.ToDoc()

		if err := s.se.Index(ctx, &doc); err != nil {
			return err
		}
		return nil
	case "delete":
		if err := s.se.Delete(ctx, payload.Slug); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: unkown action", errs.ErrInval)
}
