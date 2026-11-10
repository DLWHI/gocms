package advertisement

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/search"
	"net/url"
	"strings"
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

func isValidURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}

	if u.Host == "" {
		return false
	}

	return true
}

func (s *Service) Create(ctx context.Context, ad *domain.Advertisement) error {
	if !s.auth.Authorize(ctx, domain.EntityName(&domain.Advertisement{}), domain.ActionCreate) {
		return domain.ErrPerm
	}

	name, err := domain.NormalizeName(ad.Name)
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("%w: name cannot be empty", domain.ErrInval)
	}
	ad.Name = name

	group := strings.TrimSpace(ad.Group)
	if group != "" {
		group, err := domain.NormalizeSlug(group)
		if err != nil {
			return err
		}
		ad.Group = group
	}

	url, err := domain.NormalizeURL(ad.URL)
	if err != nil {
		return err
	}
	if url == "" {
		return fmt.Errorf("%w: Missing url", domain.ErrInval)
	}
	ad.URL = url

	banner, err := domain.NormalizeURL(ad.Banner)
	if err != nil {
		return err
	}
	if banner == "" {
		return fmt.Errorf("%w: Missing banner image", domain.ErrInval)
	}
	ad.Banner = banner

	if ad.ExpiresAt != nil && *ad.ExpiresAt < 0 {
		return fmt.Errorf("%w: ExpiresAt must be non-negative", domain.ErrInval)
	}

	if err := s.repo.Create(ctx, ad); err != nil {
		logger.Error("Ads.Service", "Failed to create advertisement", err)
		return err
	}

	return nil
}

func (s *Service) Update(ctx context.Context, id uint, diff *domain.AdvertisementDiff) error {
	if !s.auth.Authorize(ctx, domain.EntityName(&domain.Advertisement{}), domain.ActionEdit) {
		return domain.ErrPerm
	}

	if diff.Name != nil {
		name, err := domain.NormalizeName(*diff.Name)
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("%w: name cannot be empty", err)
		}
		diff.Name = &name
	}

	if diff.Group != nil {
		group := strings.TrimSpace(*diff.Group)
		if group != "" {
			group, err := domain.NormalizeSlug(group)
			if err != nil {
				return err
			}
			diff.Group = &group
		}
	}

	if diff.URL != nil {
		url, err := domain.NormalizeURL(*diff.URL)
		if err != nil {
			return err
		}
		if url == "" {
			return fmt.Errorf("%w: Missing url", domain.ErrInval)
		}
		diff.URL = &url
	}

	if diff.Banner != nil {
		banner, err := domain.NormalizeURL(*diff.Banner)
		if err != nil {
			return err
		}
		if banner == "" {
			return fmt.Errorf("%w: Missing banner image", domain.ErrInval)
		}
		diff.Banner = &banner
	}

	if diff.ExpiresAt != nil && *diff.ExpiresAt < 0 {
		return fmt.Errorf("%w: ExpiresAt must be non-negative", domain.ErrInval)
	}

	if err := s.repo.Update(ctx, id, diff); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Ads.Service", "Failed to update advertisement", err)
		}
		return err
	}

	return nil
}

func (s *Service) Get(ctx context.Context, id uint) (*domain.Advertisement, error) {
	ad, err := s.repo.Get(ctx, id)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Ads.Service", "Failed to retrieve advertisement", err)
		}
		return nil, err
	}

	return ad, nil
}

func (s *Service) GetGroup(ctx context.Context, group string) ([]domain.Advertisement, error) {
	ads, err := s.repo.GetGroup(ctx, group)
	if err != nil {
		logger.Error("Ads.Service", "Failed to retrieve advertisements", err)
		return nil, err
	}

	return ads, nil
}

func (s *Service) AllGroups(ctx context.Context) ([]string, error) {
	groups, err := s.repo.GetGroups(ctx)
	if err != nil {
		logger.Error("Ads.Service", "Failed to retrieve advertisement groups", err)
		return nil, err
	}

	return groups, nil
}

func (s *Service) Search(ctx context.Context, filters *SearchParams) (*SearchResult, error) {
	search.ClampParams(&filters.Params)

	ads, err := s.se.Search(ctx, filters)
	if err != nil {
		if !errors.Is(err, domain.ErrInval) {
			logger.Error("Ads.Service", "Failed to search advertisements", err)
		}
		return nil, err
	}

	return ads, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if !s.auth.Authorize(ctx, domain.EntityName(&domain.Advertisement{}), domain.ActionDelete) {
		return domain.ErrPerm
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Ads.Service", "Failed to delete advertisement", err)
		}
		return err
	}

	return nil
}
