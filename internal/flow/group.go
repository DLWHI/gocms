package flow

import (
	"gocms/internal/domain"
	"maps"
	"sync"
	"unicode/utf8"
)

type Result struct {
	Val    any
	Err    error
	Shared bool
}

type Group struct {
	mu    sync.Mutex
	calls map[string]*call
}

type call struct {
	done  chan Result
	chans []chan<- Result
}

func (g *Group) Do(key string, fn func() (any, error)) (any, error, bool) {
	ch := make(chan Result, 1)
	g.join(key, ch, fn)
	res := <-ch
	return res.Val, res.Err, res.Shared
}

func (g *Group) DoChan(key string, fn func() (any, error)) <-chan Result {
	ch := make(chan Result, 1)
	g.join(key, ch, fn)
	return ch
}

func (g *Group) Forget(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.calls, key)
}

func (g *Group) ForgetPattern(pattern string) {
	g.mu.Lock()
	snap := maps.Clone(g.calls)
	g.mu.Unlock()

	hits := make(map[string]*call, 0)
	for key, c := range snap {
		if g.globMatch(key, pattern) {
			hits[key] = c
		}
	}

	g.mu.Lock()
	for key, c := range hits {
		if g.calls[key] == c {
			delete(g.calls, key)
		}
	}
	g.mu.Unlock()
}

func (g *Group) join(key string, ch chan Result, fn func() (any, error)) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.calls == nil {
		g.calls = make(map[string]*call)
	}
	if c, ok := g.calls[key]; ok {
		c.chans = append(c.chans, ch)
		return
	}
	c := &call{done: make(chan Result, 1)}
	c.chans = append(c.chans, ch)
	g.calls[key] = c

	go g.run(key, c, fn)
}

func (g *Group) finish(key string, c *call) {
	res, ok := <-c.done
	g.mu.Lock()
	if g.calls[key] == c {
		delete(g.calls, key)
	}
	g.mu.Unlock()
	if !ok {
		res.Err = domain.ErrAbnormal
	}

	res.Shared = len(c.chans) > 1

	for _, ch := range c.chans {
		ch <- res
	}
}

func (g *Group) run(key string, c *call, fn func() (any, error)) {
	defer g.finish(key, c)
	defer close(c.done)
	val, err := fn()
	res := Result{Val: val, Err: err}
	c.done <- res
}

func (g *Group) globMatch(key string, pattern string) bool {
	var k, p int
	var wildcard, last int
	wildcard = -1
	for len(key) > k {
		kRune, kSize := utf8.DecodeRuneInString(key[k:])

		if p < len(pattern) {
			pRune, pSize := utf8.DecodeRuneInString(pattern[p:])

			switch {
			case pRune == '*':
				wildcard, last = p, k
				p++
				continue
			case pRune == '?' || pRune == kRune:
				k += kSize
				p += pSize
				continue
			}
		}

		if wildcard < 0 {
			return false
		}
		_, size := utf8.DecodeRuneInString(key[last:])
		last += size
		k = last
		p = wildcard + 1
	}
	for p < len(pattern) && pattern[p] == '*' {
		p++
	}
	return p == len(pattern)
}
