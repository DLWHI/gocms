package locale

import (
	"context"
	"slices"
	"sync"

	"gocms/internal/domain"
	"gocms/internal/logger"
)

type LocalCacheRepository struct {
	Repository
	locales []domain.Locale
	lock    sync.RWMutex
}

func NewLocalCacheRepository(ctx context.Context, origin Repository) *LocalCacheRepository {
	locales, err := origin.GetAll(ctx)
	if err != nil {
		logger.Warn("Locale.LocalCache", "Failed to retreive locales for caching, requests will fetch directly from db", err)
	}
	return &LocalCacheRepository{
		Repository: origin,
		locales:    locales,
	}
}

func (r *LocalCacheRepository) Create(ctx context.Context, locale *domain.Locale) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if err := r.Repository.Create(ctx, locale); err != nil {
		return err
	}
	if r.locales != nil {
		r.locales = append(r.locales, *locale)
	}
	return nil
}

func (r *LocalCacheRepository) GetAll(ctx context.Context) ([]domain.Locale, error) {
	var locales []domain.Locale

	r.lock.RLock()
	if r.locales != nil {
		locales = slices.Clone(r.locales)
		r.lock.RUnlock()
		return locales, nil
	}
	r.lock.RUnlock()

	locales, err := r.Repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return locales, nil
}

func (r *LocalCacheRepository) Delete(ctx context.Context, code string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if err := r.Repository.Delete(ctx, code); err != nil {
		return err
	}
	r.locales = slices.DeleteFunc(r.locales, func(locale domain.Locale) bool {
		return locale.Code == code
	})
	return nil
}
