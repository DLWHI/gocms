package post

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/permissions"
	"gocms/internal/search"
	"strings"
	"time"
	"unicode"
)

type Service struct {
	repo Repository
	se   SearchEngine
}

func NewService(repo Repository, se SearchEngine) *Service {
	return &Service{
		repo: repo,
		se:   se,
	}
}

func (s *Service) hasPermissionsOnPost(ctx context.Context, action string, id uint) error {
	initiator := ctx.Value("username")
	level := ctx.Value("userperm")

	if initiator != nil && level != nil {
		if permissions.Has("posts", action, level.(uint32)) {
			return nil
		}
		post, err := s.repo.Get(ctx, id)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				logger.Error("Post.Service", "Failed to verify post ownership", err)
			}
			return err
		}
		if post.Author != initiator.(string) {
			return errs.ErrCred
		}
	} else {
		return errs.ErrCred
	}
	return nil
}

func (s *Service) sanitize(ctx context.Context, post *Document) error {
	post.Slug = strings.TrimSpace(strings.ToLower(post.Slug))
	if post.Slug == "" {
		return fmt.Errorf("%w: slug cannot be empty", errs.ErrInval)
	} else if strings.IndexFunc(post.Slug, unicode.IsSpace) >= 0 {
		return fmt.Errorf("%w: no spaces allowed in slug", errs.ErrInval)
	}

	post.Locale = strings.TrimSpace(strings.ToLower(post.Locale))
	if post.Locale == "" {
		locale := ctx.Value("locale")
		if locale == nil {
			return fmt.Errorf("%w: empty locale", errs.ErrInval)
		}
		post.Locale = locale.(string)
	}

	post.Keywords = strings.TrimSpace(post.Keywords)
	return nil
}

func (s *Service) Create(ctx context.Context, post *Document) error {
	author := ctx.Value("username")
	level := ctx.Value("userperm")

	if author == nil || level == nil || !permissions.Has("posts", "create", level.(uint32)) {
		return errs.ErrPerm
	}

	post.Author = author.(string)

	if err := s.sanitize(ctx, post); err != nil {
		return err
	}

	if post.PublishedAt != nil {
		post.CreatedAt = *post.PublishedAt
	} else {
		post.CreatedAt = time.Now().Unix()
	}

	model := post.ToModel()
	if err := s.repo.Create(ctx, &model); err != nil {
		logger.Error("Post.Service", "Failed to create post", err)
		return err
	}
	post.PostID = model.PostID

	return nil
}

func (s *Service) Edit(ctx context.Context, post *Document) (*Document, error) {
	if err := s.hasPermissionsOnPost(ctx, "edit", post.PostID); err != nil {
		return nil, err
	}

	if err := s.sanitize(ctx, post); err != nil {
		return nil, err
	}

	model := post.ToModel()
	err := s.repo.UpdateTranslation(ctx, &model)
	if err != nil {
		logger.Error("Post.Service", "Failed to update post", err)
		return nil, err
	}

	return post, nil
}

func (s *Service) IncrementViews(ctx context.Context, id uint, value uint) error {
	if err := s.repo.IncrementViews(ctx, id, value); err != nil {
		logger.Error("Post.Service", "Failed to increment post view count", err)
		return err
	}

	return nil
}

func (s *Service) SetViews(ctx context.Context, id uint, value uint) error {
	post, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error("Post.Service", "Failed to retrieve post view count", err)
		return err
	}

	updated := Meta{
		ID:    post.ID,
		Views: value,
	}

	if err := s.repo.Update(ctx, &updated); err != nil {
		logger.Error("Post.Service", "Failed to update post view count", err)
		return err
	}

	return nil
}

func (s *Service) Search(ctx context.Context, filters *SearchParams) (*SearchResult, error) {
	search.ClampParams(&filters.Params)

	posts, err := s.se.Search(ctx, filters)
	if err != nil {
		logger.Error("Post.Service", "Failed to search posts", err)
		return nil, err
	}
	return posts, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*Document, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%w: empty locale", errs.ErrInval)
	}

	post, err := s.repo.GetTranslation(ctx, id, locale.(string))
	if err != nil {
		logger.Error("Post.Service", "Failed to retrieve post", err)
		return nil, err
	}

	doc := post.ToDoc()
	doc.Categories, err = s.repo.GetCategories(ctx, id)
	if err != nil {
		logger.Error("Post.Service", "Failed to retrieve post categories", err)
		return nil, err
	}

	return &doc, nil
}

func (s *Service) Read(ctx context.Context, id uint, slug string) (*Document, error) {
	locale := ctx.Value("locale")
	if locale == nil {
		return nil, fmt.Errorf("%w: empty locale", errs.ErrInval)
	}

	post, err := s.repo.GetTranslation(ctx, id, locale.(string))
	if err != nil {
		logger.Error("Post.Service", "Failed to retrieve post translation", err)
		return nil, err
	}
	if post.Post.Slug != slug {
		return nil, errs.ErrNotFound
	}
	if post.Post.PublishedAt == nil {
		return nil, errs.ErrNotFound
	}

	doc := post.ToDoc()
	doc.Categories, err = s.repo.GetCategories(ctx, id)
	if err != nil {
		logger.Error("Post.Service", "Failed to retrieve post categories", err)
		return nil, err
	}

	return &doc, nil
}

func (s *Service) GetViews(ctx context.Context, id uint) (uint, error) {
	post, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error("Post.Service", "Failed to retrieve post view count", err)
		return 0, err
	}

	return post.Views, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if err := s.hasPermissionsOnPost(ctx, "delete", id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error("Post.Service", "Failed to delete post", err)
		return err
	}

	return nil
}

func (s *Service) DeleteTranslation(ctx context.Context, id uint, locale string) error {
	if locale == "" {
		l := ctx.Value("locale")
		if l == nil {
			return fmt.Errorf("%w: empty locale", errs.ErrInval)
		}
		locale = l.(string)
	}

	if err := s.hasPermissionsOnPost(ctx, "delete", id); err != nil {
		return err
	}

	if err := s.repo.DeleteTranslation(ctx, id, locale); err != nil {
		logger.Error("Post.Service", "Failed to delete post translation", err)
		return err
	}

	return nil
}

func (s *Service) SyncSearch(ctx context.Context, action outbox.Event, data []byte) error {
	var payload MessagePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch payload.Action {
	case "create", "update":
		var doc Document
		if payload.Locale == "" {
			post, err := s.repo.Get(ctx, payload.ID)
			if err != nil {
				if errors.Is(err, errs.ErrNotFound) {
					return nil
				}
				return err
			}
			doc = post.ToDoc()
		} else {
			post, err := s.repo.GetTranslation(ctx, payload.ID, payload.Locale)
			if err != nil {
				if errors.Is(err, errs.ErrNotFound) {
					return nil
				}
				return err
			}
			doc = post.ToDoc()
		}
		categories, err := s.repo.GetCategories(ctx, payload.ID)
		if err != nil {
			return err
		}
		doc.Categories = categories
		if err := s.se.Index(ctx, &doc); err != nil {
			return err
		}
		return nil
	case "delete":
		if err := s.se.Delete(ctx, payload.ID); err != nil {
			return err
		}
		return nil
	case "delete.translation":
		if err := s.se.DeleteTranslation(ctx, payload.ID, payload.Locale); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: unkown action", errs.ErrInval)
}
