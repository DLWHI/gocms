package advertisement

import (
	"context"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/search"
	"net/url"
	"strings"
	"unicode"
)

type Service struct {
	repo Repository
	se   SearchEngine
	auth auth.Service
}

func NewService(repo Repository, se SearchEngine) *Service {
	return &Service{
		repo: repo,
		se:   se,
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

func (s *Service) sanitize(ad *domain.Advertisement) error {
	ad.Group = strings.TrimSpace(strings.ToLower(ad.Group))
	if strings.IndexFunc(ad.Group, unicode.IsSpace) >= 0 {
		return fmt.Errorf("%w: no spaces allowed in group", domain.ErrInval)
	}

	if !isValidURL(ad.URL) {
		return fmt.Errorf("%w: URL is not valid", domain.ErrInval)
	}

	return nil
}

func (s *Service) Create(ctx context.Context, ad *domain.Advertisement) error {
	if !s.auth.Authorize(ctx, domain.EntityName(&domain.Advertisement{}), domain.ActionCreate) {
		return domain.ErrCred
	}

	name, err := domain.NormalizeName(ad.Name)
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("%w: name cannot be empty", err)
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
	ad.URL = url

	banner, err := domain.NormalizeURL(ad.Banner)
	if err != nil {
		return err
	}
	ad.Banner = banner

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
		diff.URL = &url
	}

	if diff.Banner != nil {
		banner, err := domain.NormalizeURL(*diff.Banner)
		if err != nil {
			return err
		}
		diff.Banner = &banner
	}
	if err := s.repo.Update(ctx, id, diff); err != nil {
		logger.Error("Ads.Service", "Failed to update advertisement", err)
		return err
	}

	return nil
}

func (s *Service) Get(ctx context.Context, id uint) (*domain.Advertisement, error) {
	ad, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error("Ads.Service", "Failed to retrieve advertisement", err)
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
		logger.Error("Ads.Service", "Failed to search advertisements", err)
		return nil, err
	}

	return ads, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if !s.auth.Authorize(ctx, domain.EntityName(&domain.Advertisement{}), domain.ActionDelete) {
		return domain.ErrCred
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error("Ads.Service", "Failed to delete advertisement", err)
		return err
	}

	return nil
}
