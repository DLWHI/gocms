package flow

import (
	"context"
	"sync"
)

type Runner interface {
	Start(ctx context.Context, group *sync.WaitGroup, params ...any)
	Stop(ctx context.Context) error
}
