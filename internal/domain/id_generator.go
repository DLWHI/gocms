package domain

import (
	"sync/atomic"
)

type IDGenerator struct {
	last atomic.Int64
}

func (g *IDGenerator) Next() int64 {
	return g.last.Add(1)
}
