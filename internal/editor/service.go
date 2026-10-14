package editor

import (
	"context"
	"fmt"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/permissions"
	"gocms/internal/search"
	"gocms/internal/storage"
	"io"
	"strings"
)

type Service struct {
	repo       Repository
	lookup     SearchEngine
	files      *storage.Manager
	pathFormat string
}

func NewService(db Repository, se SearchEngine, content *storage.Manager) *Service {
	return &Service{
		repo:       db,
		lookup:     se,
		files:      content,
		pathFormat: "/%s/%s.md",
	}
}

func (s *Service) Create(ctx context.Context, editor *Doc) error {
	level := ctx.Value("userperm")

	if level == nil || !permissions.Has("editors", "create", level.(uint32)) {
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

	content := fmt.Sprintf(s.pathFormat, editor.User, editor.Locale)
	if err := s.files.Create(content); err != nil {
		logger.Error("Editor.Service", "Failed to create editor biography storage", err)
		return err
	}

	if err := s.lookup.Index(ctx, editor); err != nil {
		logger.Error("Editor.Service", "Failed to add editor profile to the search index", err)
		return err
	}

	return nil
}

func (s *Service) Edit(ctx context.Context, editor *Doc) error {
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

	content := fmt.Sprintf(s.pathFormat, editor.User, editor.Locale)
	if !s.files.Exists(content) {
		if err := s.files.Create(content); err != nil {
			logger.Error("Editor.Service", "Failed to create editor biography storage", err)
			return err
		}
	}

	if err := s.lookup.Index(ctx, editor); err != nil {
		logger.Error("Editor.Service", "Failed to update editor profile in the search index", err)
		return err
	}

	return nil
}

func (s *Service) Get(ctx context.Context, username string) (*Doc, error) {
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

func (s *Service) Read(ctx context.Context, username string) (*Doc, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%w: empty locale", errs.ErrInval)
	}

	editor, err := s.repo.GetTranslation(ctx, username, locale.(string))
	if err != nil {
		logger.Error("Editor.Service", "Failed to retrieve editor profile translation", err)
		return nil, err
	}

	doc := editor.ToDoc()
	return &doc, nil
}

func (s *Service) All(ctx context.Context) ([]Doc, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%w: empty locale", errs.ErrInval)
	}

	model, err := s.repo.GetAll(ctx, locale.(string))
	if err != nil {
		logger.Error("Editor.Service", "Failed to list editor profiles", err)
		return nil, err
	}

	editors := make([]Doc, 0, len(model))

	for _, editor := range model {
		if !editor.UserData.SuspendedAt.Valid {
			editors = append(editors, editor.ToDoc())
		}
	}

	return editors, nil
}

func (s *Service) Search(ctx context.Context, filters *SearchParams) (*SearchResult, error) {
	search.ClampParams(&filters.Params)

	editors, err := s.lookup.Search(ctx, filters)
	if err != nil {
		logger.Error("Editor.Service", "Failed to search editor profiles", err)
		return nil, err
	}

	return editors, nil
}

func (s *Service) Suspend(ctx context.Context, username string) error {
	level := ctx.Value("userperm")

	if level == nil || !permissions.Has("editors", "delete", level.(uint32)) {
		return errs.ErrCred
	}
	if err := s.repo.Delete(ctx, username); err != nil {
		logger.Error("Editor.Service", "Failed to suspend editor profile", err)
		return err
	}

	if err := s.lookup.Delete(ctx, username); err != nil {
		logger.Error("Editor.Service", "Failed to remove editor profile from the search index", err)
		return err
	}
	return nil
}

func (s *Service) EditBio(ctx context.Context, username string, content io.Reader) error {
	initiator := ctx.Value("username")

	if initiator == nil || initiator.(string) != username {
		return errs.ErrCred
	}

	locale := ctx.Value("locale")
	if locale == nil {
		return fmt.Errorf("%s: missing locale", errs.ErrInval)
	}

	path := fmt.Sprintf(s.pathFormat, username, locale.(string))
	if err := s.files.Save(path, content); err != nil {
		logger.Error("Editor.Service", "Failed to save editor biography", err)
		return err
	}
	return nil
}

func (s *Service) ReadBio(ctx context.Context, username string) (io.ReadCloser, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%s: missing locale", errs.ErrInval)
	}

	bioPath := fmt.Sprintf(s.pathFormat, username, locale)
	content, err := s.files.Read(bioPath)
	if err != nil {
		logger.Error("Editor.Service", "Failed to read editor biography", err)
		return nil, err
	}
	return content, nil
}
