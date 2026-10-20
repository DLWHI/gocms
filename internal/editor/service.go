package editor

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

func (s *Service) Create(ctx context.Context, editor *Document) error {
	if !auth.HasPermission(ctx, "editors", "create") {
		return errs.ErrCred
	}

	editor.Locale = strings.TrimSpace(strings.ToLower(editor.Locale))
	if editor.Locale == "" {
		locale := ctx.Value("locale")
		if locale == nil {
			return fmt.Errorf("%w: empty locale", errs.ErrInval)
		}
		editor.Locale = locale.(string)
	}

	model := editor.ToModel()
	if err := s.repo.Create(ctx, &model); err != nil {
		logger.Error("Editor.Service", "Failed to create editor profile", err)
		return err
	}

	return nil
}

func (s *Service) Edit(ctx context.Context, editor *Document) error {
	initiator := ctx.Value("username")

	if initiator == nil || initiator.(string) != editor.User {
		return errs.ErrCred
	}

	editor.Locale = strings.TrimSpace(strings.ToLower(editor.Locale))
	if editor.Locale == "" {
		locale := ctx.Value("locale")
		if locale == nil {
			return fmt.Errorf("%w: empty locale", errs.ErrInval)
		}
		editor.Locale = locale.(string)
	}

	model := editor.ToModel()
	if err := s.repo.Update(ctx, &model); err != nil {
		logger.Error("Editor.Service", "Failed to update editor profile", err)
		return err
	}

	return nil
}

func (s *Service) Get(ctx context.Context, username string) (*Document, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%w: empty locale", errs.ErrInval)
	}

	editor, err := s.repo.Get(ctx, username, locale.(string))
	if err != nil {
		logger.Error("Editor.Service", "Failed to retrieve editor profile", err)
		return nil, err
	}

	doc := editor.ToDoc()
	return &doc, nil
}

func (s *Service) All(ctx context.Context, suspended *bool) ([]Document, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%w: empty locale", errs.ErrInval)
	}

	model, err := s.repo.GetAll(ctx, locale.(string), suspended)
	if err != nil {
		logger.Error("Editor.Service", "Failed to list editor profiles", err)
		return nil, err
	}

	editors := make([]Document, 0, len(model))

	for _, editor := range model {
		if !editor.UserData.SuspendedAt.Valid {
			editors = append(editors, editor.ToDoc())
		}
	}

	return editors, nil
}

func (s *Service) Search(ctx context.Context, filters *SearchParams) (*SearchResult, error) {
	search.ClampParams(&filters.Params)

	editors, err := s.se.Search(ctx, filters)
	if err != nil {
		logger.Error("Editor.Service", "Failed to search editor profiles", err)
		return nil, err
	}

	return editors, nil
}

func (s *Service) Suspend(ctx context.Context, username string) error {
	if !auth.HasPermission(ctx, "editors", "delete") {
		return errs.ErrCred
	}

	if err := s.repo.Delete(ctx, username); err != nil {
		logger.Error("Editor.Service", "Failed to suspend editor profile", err)
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
		editor, err := s.repo.Get(ctx, payload.Username, payload.Locale)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return nil
			}
			return err
		}
		doc := editor.ToDoc()

		if err := s.se.Index(ctx, &doc); err != nil {
			return err
		}
		return nil
	case "delete":
		if err := s.se.Delete(ctx, payload.Username, *payload.SuspendedAt); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: unkown action", errs.ErrInval)
}
