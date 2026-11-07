package post

import (
	"context"
	"gocms/internal/cache"
	"gocms/internal/domain"
	"gocms/internal/flow"
	"strconv"
	"time"
)

type FlowRepository struct {
	Repository
	keys    cache.Keyspace
	group   flow.Group
	timeout time.Duration
}

func NewFlowRepository(db Repository, timeout time.Duration, config cache.Keyspace) *FlowRepository {
	return &FlowRepository{
		Repository: db,
		keys:       config,
		timeout:    timeout,
	}
}

func (r *FlowRepository) Update(ctx context.Context, id uint, language string, diff *domain.PostDiff) error {
	if err := r.Repository.Update(ctx, id, language, diff); err != nil {
		return err
	}
	r.group.ForgetPattern(r.keys.KeyWithSuffix(strconv.FormatUint(uint64(id), 10), "translation:*"))
	return nil
}

func (r *FlowRepository) Get(ctx context.Context, id uint, language string) (*domain.Post, error) {
	key := r.keys.KeyWithSuffix(strconv.FormatUint(uint64(id), 10), "translation:"+language)
	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.timeout)
		defer cancel()

		retrieved, err := r.Repository.Get(bg, id, language)
		if err != nil {
			return nil, err
		}
		return retrieved, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.(*domain.Post), nil
	}
}

func (r *FlowRepository) CreatorOf(ctx context.Context, id uint) (string, error) {
	key := r.keys.KeyWithSuffix(strconv.FormatUint(uint64(id), 10), "created_by")
	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.timeout)
		defer cancel()

		retrieved, err := r.Repository.GetCreator(bg, id)
		if err != nil {
			return nil, err
		}
		return retrieved, err
	})

	select {
	case <-ctx.Done():
		return "", ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return "", res.Err
		}
		return res.Val.(string), nil
	}
}

func (r *FlowRepository) Delete(ctx context.Context, id uint) error {
	if err := r.Repository.Delete(ctx, id); err != nil {
		return err
	}
	r.group.ForgetPattern(r.keys.KeyWithSuffix(strconv.FormatUint(uint64(id), 10), "translation:*"))
	return nil
}
