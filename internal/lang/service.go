package lang

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"strings"
	"sync"
	"sync/atomic"
)

type snapshot struct {
	langs    map[string]Language
	fallback Language
}

type LocalService struct {
	snap atomic.Pointer[snapshot]
	repo Repository
	auth domain.AuthService
	mu   sync.Mutex
}

func NewLocalService(ctx context.Context, fallback string, repo Repository, auth domain.AuthService) (*LocalService, error) {
	svc := &LocalService{repo: repo, auth: auth}
	svc.snap.Store(&snapshot{fallback: Language{Code: fallback}})
	if err := svc.Reload(ctx, nil); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *LocalService) SetDefaultLang(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sn := *s.snap.Load()

	lang, exists := sn.langs[code]
	if !exists {
		return fmt.Errorf("%w: language is not supported", domain.ErrInval)
	}
	sn.fallback = lang
	s.snap.Store(&sn)
	return nil
}

func (s *LocalService) GetSupported() []Language {
	sn := s.snap.Load()
	langs := make([]Language, 0, len(sn.langs))

	for _, lang := range sn.langs {
		langs = append(langs, lang)
	}
	return langs
}

func (s *LocalService) IsSupported(code string) bool {
	_, exists := s.snap.Load().langs[code]
	return exists
}

func (s *LocalService) DefaultLang() Language {
	return s.snap.Load().fallback
}

func (s *LocalService) ResolveLang(ctx context.Context, code string) (string, error) {
	normalized := strings.TrimSpace(code)
	if normalized == "" {
		code, ok := ctx.Value(domain.LangContextKey).(string)
		if !ok {
			return "", fmt.Errorf("%w: invalid language code", domain.ErrInval)
		} else if code == "" {
			return "", fmt.Errorf("%w: missing language code", domain.ErrInval)
		}

		return code, nil
	}
	if !IsISO639_1Code(normalized) {
		return "", fmt.Errorf("%w: invalid language code", domain.ErrInval)
	}
	if !s.IsSupported(normalized) {
		return "", fmt.Errorf("%w: language is not supported", domain.ErrInval)
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

	sn := *s.snap.Load()

	if value, ok := langmap[sn.fallback.Code]; ok {
		sn.fallback = value
	} else if len(langs) > 0 {
		logger.Warn("Language.Service", "Fallback not supported, using first language", nil, "value", langs[0])
		sn.fallback = langs[0]
	} else {
		logger.Warn("Language.Service", "Deleted last language, fallback will be applied on next add", nil)
		sn.fallback = Language{}
	}

	sn.langs = langmap

	s.snap.Store(&sn)
	return nil
}
