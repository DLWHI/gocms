package lang

import (
	"context"
	"errors"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"sync"
	"sync/atomic"
)

type LocalService struct {
	snap atomic.Pointer[map[string]Language]
	repo Repository
	auth auth.Service
	mu   sync.Mutex
}

func NewLocalService(ctx context.Context, repo Repository, auth auth.Service) (*LocalService, error) {
	svc := &LocalService{repo: repo, auth: auth}
	svc.snap.Store(&map[string]Language{})
	if err := svc.Reload(ctx, nil); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *LocalService) GetSupported() []Language {
	languages := *s.snap.Load()
	res := make([]Language, 0, len(languages))

	for _, language := range languages {
		res = append(res, language)
	}
	return res
}

func (s *LocalService) IsSupported(code string) bool {
	langs := *s.snap.Load()
	_, exists := langs[code]
	return exists
}

func (s *LocalService) Add(ctx context.Context, language *Language) error {
	if !s.auth.Authorize(ctx, EntityName, domain.ActionCreate) {
		return domain.ErrCred
	}

	created := *language
	name, err := domain.NormalizeName(language.Name)
	if err != nil {
		return err
	}
	created.Name = name

	code, err := NormalizeLangCode(language.Code)
	if err != nil {
		return err
	}
	created.Code = code

	if err := s.repo.Create(ctx, &created); err != nil {
		if !errors.Is(err, domain.ErrDuplicate) {
			logger.Error("Language.Service", "Failed to create language", err)
		}
		return err
	}

	return nil
}

func (s *LocalService) Remove(ctx context.Context, code string) error {
	if !s.auth.Authorize(ctx, EntityName, domain.ActionDelete) {
		return domain.ErrCred
	}

	code, err := NormalizeLangCode(code)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, code); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			logger.Error("Language.Service", "Failed to delete language", err)
		}
		return err
	}

	return nil
}

func (s *LocalService) Reload(ctx context.Context, _ []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	languages, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	langmap := make(map[string]Language, 0)
	for _, language := range languages {
		langmap[language.Code] = language
	}

	s.snap.Store(&langmap)
	return nil
}
