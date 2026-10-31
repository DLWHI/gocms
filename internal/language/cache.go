package language

import (
	"context"
	"slices"
	"sync"

	"gocms/internal/domain"
	"gocms/internal/logger"
)

type LocalCacheRepository struct {
	Repository
	langs []domain.Language
	lock  sync.RWMutex
}

func NewLocalCacheRepository(ctx context.Context, origin Repository) *LocalCacheRepository {
	langs, err := origin.GetAll(ctx)
	if err != nil {
		logger.Warn("Languages.LocalCache", "Failed to retreive languages for caching, requests will fetch directly from db", err)
	}
	return &LocalCacheRepository{
		Repository: origin,
		langs:      langs,
	}
}

func (r *LocalCacheRepository) Create(ctx context.Context, langs *domain.Language) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if err := r.Repository.Create(ctx, langs); err != nil {
		return err
	}
	if r.langs != nil {
		r.langs = append(r.langs, *langs)
	}
	return nil
}

func (r *LocalCacheRepository) GetAll(ctx context.Context) ([]domain.Language, error) {
	var langs []domain.Language

	r.lock.RLock()
	if r.langs != nil {
		langs = slices.Clone(r.langs)
		r.lock.RUnlock()
		return langs, nil
	}
	r.lock.RUnlock()

	langs, err := r.Repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return langs, nil
}

func (r *LocalCacheRepository) Delete(ctx context.Context, code string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if err := r.Repository.Delete(ctx, code); err != nil {
		return err
	}
	r.langs = slices.DeleteFunc(r.langs, func(lang domain.Language) bool {
		return lang.Code == code
	})
	return nil
}
