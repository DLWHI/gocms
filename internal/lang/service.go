package lang

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"strings"
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
	langs := *s.snap.Load()
	res := make([]Language, 0, len(langs))

	for _, lang := range langs {
		res = append(res, lang)
	}
	return res
}

func (s *LocalService) IsSupported(code string) bool {
	langs := *s.snap.Load()
	_, exists := langs[code]
	return exists
}

func (s *LocalService) ResolveLang(ctx context.Context, code string) (string, error) {
	normalized := strings.TrimSpace(code)
	if normalized != "" && !IsISO639_1Code(normalized) {
		return "", fmt.Errorf("%w: invalid language code", domain.ErrInval)
	}
	if !s.IsSupported(normalized) {
		code, ok := ctx.Value(domain.LangContextKey).(string)
		if !ok {
			return "", fmt.Errorf("%w: invalid language code", domain.ErrInval)
		} else if code == "" {
			return "", fmt.Errorf("%w: missing language code", domain.ErrInval)
		}
		return code, nil
	}
	return normalized, nil
}

func (s *LocalService) Add(ctx context.Context, lang *Language) error {
	if !s.auth.Authorize(ctx, domain.LangEntityName, domain.ActionCreateName) {
		return domain.ErrCred
	}

	created := *lang
	name, err := domain.NormalizeName(lang.Name)
	if err != nil {
		return err
	}
	created.Name = name

	code, err := NormalizeLangCode(lang.Code)
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
	if !s.auth.Authorize(ctx, domain.LangEntityName, domain.ActionDeleteName) {
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
	langs, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	langmap := make(map[string]Language, 0)
	for _, lang := range langs {
		langmap[lang.Code] = lang
	}

	s.snap.Store(&langmap)
	return nil
}
