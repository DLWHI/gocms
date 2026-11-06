package post

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/lang"
	"gocms/internal/logger"
	"gocms/internal/search"
	"gocms/internal/storage"
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

func (s *Service) authorize(ctx context.Context, action string, id uint) error {
	initiator, ok := ctx.Value(domain.UsernameContextKey).(string)
	if !ok {
		return domain.ErrPerm
	}

	if !s.auth.Authorize(ctx, domain.EntityName(&domain.Post{}), action) {
		return domain.ErrPerm
	}
	creator, err := s.repo.GetCreator(ctx, id)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Post.Service", "Failed to verify post ownership", err)
		}
		return err
	}
	if creator != initiator {
		return domain.ErrPerm
	}
	return nil
}

func (s *Service) Create(ctx context.Context, post *domain.Post) error {
	if !s.auth.Authorize(ctx, domain.EntityName(&domain.Post{}), domain.ActionCreate) {
		return domain.ErrCred
	}

	creator, ok := ctx.Value("username").(string)
	if !ok {
		return domain.ErrPerm
	}
	post.CreatedBy = creator

	language, err := lang.ResolveLang(ctx)
	if err != nil {
		return err
	}
	post.Language = language

	slug, err := domain.NormalizeSlug(post.Slug)
	if err != nil {
		return err
	}
	post.Slug = slug

	title, err := domain.NormalizeName(post.Title)
	if err != nil {
		return err
	}
	post.Title = title

	keywords, err := domain.NormalizeKeywords(post.Keywords)
	if err != nil {
		return err
	}
	post.Keywords = keywords

	preview, err := domain.NormalizeURL(post.Preview)
	if err != nil {
		return err
	}
	post.Preview = preview

	media, err := domain.NormalizeURL(post.Media)
	if err != nil {
		return err
	}
	post.Media = media

	source, err := domain.NormalizeName(post.MediaSource)
	if err != nil {
		return err
	}
	post.MediaSource = source

	if post.MediaType != "" && !storage.IsMediaType(post.MediaType) {
		return fmt.Errorf("%w: MediaType is unknown value", err)
	}

	if err := s.repo.Create(ctx, post); err != nil {
		logger.Error("Post.Service", "Failed to create post", err)
		return err
	}

	return nil
}

func (s *Service) Edit(ctx context.Context, id uint, diff *domain.PostDiff) error {
	if err := s.authorize(ctx, domain.ActionEdit, id); err != nil {
		return err
	}

	language, err := lang.ResolveLang(ctx)
	if err != nil {
		return err
	}

	if diff.Slug != nil {
		slug, err := domain.NormalizeSlug(*diff.Slug)
		if err != nil {
			return err
		}
		diff.Slug = &slug
	}

	if diff.Title != nil {
		title, err := domain.NormalizeName(*diff.Title)
		if err != nil {
			return err
		}
		diff.Title = &title
	}

	if diff.Keywords != nil {
		keywords, err := domain.NormalizeKeywords(*diff.Keywords)
		if err != nil {
			return err
		}
		diff.Keywords = &keywords
	}

	if diff.Preview != nil {
		preview, err := domain.NormalizeURL(*diff.Preview)
		if err != nil {
			return err
		}
		diff.Preview = &preview
	}

	if diff.Media != nil {
		media, err := domain.NormalizeURL(*diff.Media)
		if err != nil {
			return err
		}
		diff.Media = &media
	}

	if diff.MediaSource != nil {
		source, err := domain.NormalizeName(*diff.MediaSource)
		if err != nil {
			return err
		}
		diff.MediaSource = &source
	}

	if diff.MediaType != nil && !storage.IsItemType(*diff.MediaType) {
		return fmt.Errorf("%w: MediaType is unknown value", err)
	}

	if err := s.repo.Update(ctx, id, language, diff); err != nil {
		if !errors.Is(err, domain.ErrNotFound) && !errors.Is(err, domain.ErrInval) {
			logger.Error("Post.Service", "Failed to update post", err)
		}
		return err
	}

	return nil
}

func (s *Service) SetViews(ctx context.Context, id uint, value uint) error {
	if !s.auth.Authorize(ctx, "views", domain.ActionEdit) {
		return domain.ErrPerm
	}

	if err := s.repo.SetViews(ctx, id, value); err != nil {
		logger.Error("Post.Service", "Failed to update post view count", err)
		return err
	}

	return nil
}

func (s *Service) CommitViews(ctx context.Context, id uint, value uint) (uint, error) {
	views, err := s.repo.IncrementViews(ctx, id, value)
	if err != nil {
		logger.Error("Post.Service", "Failed to retrieve post view count", err)
		return 0, err
	}

	return views, nil
}

func (s *Service) Search(ctx context.Context, filters *SearchParams) (*SearchResult, error) {
	_, ok := ctx.Value(domain.UsernameContextKey).(string)
	if !ok {
		filters.Select = []string{
			"post_id",
			"title",
			"language",
			"published_at",
			"created_at",
			"views",
			"author",
			"created_by",
			"categories",
		}
	} else {
		filters.Select = []string{
			"post_id",
			"title",
			"description",
			"language",
			"published_at",
			"views",
			"categories",
		}
	}

	search.ClampParams(&filters.Params)

	language, err := lang.ResolveLang(ctx)
	if err != nil {
		return nil, err
	}

	posts, err := s.se.Search(ctx, language, filters)
	if err != nil {
		logger.Error("Post.Service", "Failed to search posts", err)
		return nil, err
	}
	return posts, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*domain.Post, error) {
	language, err := lang.ResolveLang(ctx)
	if err != nil {
		return nil, err
	}

	post, err := s.repo.Get(ctx, id, language)
	if err != nil {
		logger.Error("Post.Service", "Failed to retrieve post", err)
		return nil, err
	}

	return post, nil
}

func (s *Service) GetPublished(ctx context.Context, id uint, slug string) (*domain.Post, error) {
	language, err := lang.ResolveLang(ctx)
	if err != nil {
		return nil, err
	}

	post, err := s.repo.Get(ctx, id, language)
	if err != nil {
		logger.Error("Post.Service", "Failed to retrieve post translation", err)
		return nil, err
	}
	if post.Slug != slug {
		return nil, domain.ErrNotFound
	}
	if post.PublishedAt == nil {
		return nil, domain.ErrNotFound
	}

	return post, nil
}

func (s *Service) GetViews(ctx context.Context, id uint) (uint, error) {
	views, err := s.repo.GetViews(ctx, id)
	if err != nil {
		logger.Error("Post.Service", "Failed to retrieve post view count", err)
		return 0, err
	}

	return views, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if err := s.authorize(ctx, domain.ActionDelete, id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Post.Service", "Failed to delete post", err)
		}
		return err
	}

	return nil
}
