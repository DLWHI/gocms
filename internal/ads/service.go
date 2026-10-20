package ads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/search"
	"net/url"
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

func (s *Service) sanitize(ad *Doc) error {
	ad.Group = strings.TrimSpace(strings.ToLower(ad.Group))
	if strings.IndexFunc(ad.Group, unicode.IsSpace) >= 0 {
		return fmt.Errorf("%w: no spaces allowed in group", errs.ErrInval)
	}

	if !isValidURL(ad.URL) {
		return fmt.Errorf("%w: URL is not valid", errs.ErrInval)
	}

	return nil
}

func (s *Service) Create(ctx context.Context, ad *Doc) error {
	if !auth.HasPermission(ctx, "ads", "create") {
		return errs.ErrCred
	}

	if err := s.sanitize(ad); err != nil {
		return err
	}

	model := ad.ToModel()
	if err := s.repo.Create(ctx, &model); err != nil {
		logger.Error("Ads.Service", "Failed to create advertisement", err)
		return err
	}

	return nil
}

func (s *Service) Update(ctx context.Context, ad *Doc) (*Doc, error) {
	if !auth.HasPermission(ctx, "ads", "edit") {
		return nil, errs.ErrCred
	}

	if err := s.sanitize(ad); err != nil {
		return nil, err
	}

	model := ad.ToModel()
	if err := s.repo.Update(ctx, &model); err != nil {
		logger.Error("Ads.Service", "Failed to update advertisement", err)
		return nil, err
	}

	return ad, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*Doc, error) {
	ad, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error("Ads.Service", "Failed to retrieve advertisement", err)
		return nil, err
	}

	doc := ad.ToDoc()
	return &doc, nil
}

func (s *Service) GetGroup(ctx context.Context, group string) ([]Doc, error) {
	model, err := s.repo.GetGroup(ctx, group)
	if err != nil {
		logger.Error("Ads.Service", "Failed to retrieve advertisements", err)
		return nil, err
	}

	ads := make([]Doc, 0, len(model))
	now := time.Now()
	for _, ad := range model {
		if ad.ExpiresAt == nil || ad.ExpiresAt.After(now) {
			ad.ID = 0
			ad.ExpiresAt = nil
			ad.Description = ""
			ads = append(ads, ad.ToDoc())
		}
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
	if !auth.HasPermission(ctx, "ads", "delete") {
		return errs.ErrCred
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error("Ads.Service", "Failed to delete advertisement", err)
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
		category, err := s.repo.Get(ctx, payload.ID)
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
		if err := s.se.Delete(ctx, payload.ID); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: unkown action", errs.ErrInval)
}
