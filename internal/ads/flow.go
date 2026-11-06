package ads

import (
	"context"
	"errors"
	"strconv"
	"time"

	"gocms/internal/cache"
	"gocms/internal/domain"
	"gocms/internal/flow"
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

func (r *FlowRepository) Create(ctx context.Context, ad *domain.Advertisement) error {
	if err := r.Repository.Create(ctx, ad); err != nil {
		return err
	}
	if ad.Group != "" {
		r.group.Forget(r.keys.List(groupsList))
		r.group.Forget(r.keys.Key(ad.Group))
	}
	return nil
}

func (r *FlowRepository) Update(ctx context.Context, id uint, ad *domain.AdvertisementDiff) error {
	if err := r.Repository.Update(ctx, id, ad); err != nil {
		return err
	}
	r.group.Forget(r.keys.Key(strconv.FormatUint(uint64(id), 10)))
	r.group.Forget(r.keys.List(groupsList))
	r.group.ForgetPattern(r.keys.KeyWithSuffix("*", "group"))
	return nil
}

func (r *FlowRepository) Get(ctx context.Context, id uint) (*domain.Advertisement, error) {
	key := r.keys.Key(strconv.FormatUint(uint64(id), 10))
	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		ad, err := r.Repository.Get(bg, id)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}
		return ad, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.(*domain.Advertisement), nil
	}
}

func (r *FlowRepository) GetGroup(ctx context.Context, group string) ([]domain.Advertisement, error) {
	key := r.keys.KeyWithSuffix(group, "group")

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		ads, err := r.Repository.GetGroup(bg, group)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		return ads, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.([]domain.Advertisement), nil
	}
}

func (r *FlowRepository) GetGroups(ctx context.Context) ([]string, error) {
	key := r.keys.List(groupsList)
	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		groups, err := r.Repository.GetGroups(bg)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		return groups, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.([]string), nil
	}
}

func (r *FlowRepository) Delete(ctx context.Context, id uint) error {
	if err := r.Repository.Delete(ctx, id); err != nil {
		return err
	}
	r.group.Forget(r.keys.Key(strconv.FormatUint(uint64(id), 10)))
	r.group.Forget(r.keys.List(groupsList))
	r.group.ForgetPattern(r.keys.KeyWithSuffix("*", "group"))
	return nil
}
