package flow

import (
	"context"
	"errors"
	"gocms/internal/domain"
	"math/rand/v2"
	"time"
)

func DoWithRetry(ctx context.Context, count uint, baseBackoff time.Duration, fn func() error) error {
	maxBackoff := 10 * baseBackoff
	for i := uint(0); i < count; i++ {
		err := fn()
		if !errors.Is(err, domain.ErrRetry) {
			return err
		}
		backoff := baseBackoff * 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}

		jitter := time.Duration(rand.Int64N(int64(backoff)))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(jitter):
			baseBackoff = backoff
		}
	}
	return domain.ErrRetry
}
