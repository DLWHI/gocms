package cache

import (
	"fmt"
	"gocms/internal/domain"
	"hash/maphash"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
)

var whitelist = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
var escaper = strings.NewReplacer("%", "%25", "{", "%7B", "}", "%7D")

type KeyConfig struct {
	Prefix  []string
	Version string
}

type Keyspace interface {
	Key(id string) string
	List(name string) string
}

type SimpleKeyspace struct {
	base string
}

type HashedKeyspace struct {
	prefix *SimpleKeyspace
}

type FlightKeyspace struct {
	seed  maphash.Seed
	lists atomic.Uint64
	items []atomic.Uint64
	slots uint64
}

func NewKeyspace(cfg KeyConfig, entity string) (*SimpleKeyspace, error) {
	if len(cfg.Prefix) == 0 {
		return nil, fmt.Errorf("%w: at least one prefix part is required", domain.ErrInval)
	}
	parts := make([]string, 0, len(cfg.Prefix)+2) // fresh slice: no aliasing
	parts = append(parts, cfg.Prefix...)
	parts = append(parts, entity, cfg.Version)

	for _, part := range parts {
		if !whitelist.MatchString(part) {
			return nil, fmt.Errorf("%w: %s key part must contain letters, digits, '_' or '-')", domain.ErrInval, part)
		}
	}
	return &SimpleKeyspace{base: strings.Join(parts, ":")}, nil
}

func NewHashedKeyspace(cfg KeyConfig, entity string) (*HashedKeyspace, error) {
	keyspace, err := NewKeyspace(cfg, entity)
	if err != nil {
		return nil, err
	}
	return &HashedKeyspace{prefix: keyspace}, nil
}

func NewFlightKeyspace(slots int) (*FlightKeyspace, error) {
	if slots <= 0 {
		return nil, fmt.Errorf("%w: slot count must be positive, non-null integer", domain.ErrInval)
	}
	return &FlightKeyspace{
		items: make([]atomic.Uint64, slots),
		slots: uint64(slots),
	}, nil
}

func (k *SimpleKeyspace) Key(id string) string    { return k.base + ":item:" + id }
func (k *SimpleKeyspace) List(name string) string { return k.base + ":list:" + name }

// '{', '}' and '%' are escaped with their web code
func (k *HashedKeyspace) Key(id string) string {
	return k.prefix.base + ":item:{" + escaper.Replace(id) + "}"
}
func (k *HashedKeyspace) List(name string) string {
	return k.prefix.base + ":list:{" + escaper.Replace(name) + "}"
}

func (f *FlightKeyspace) slot(id string) *atomic.Uint64 {
	return &f.items[maphash.String(f.seed, id)&(f.slots-1)]
}
func (f *FlightKeyspace) Key(id string) string {
	b := make([]byte, 0, len(id)+24)
	b = append(b, 'i', '|')
	b = append(b, string(id)...)
	b = append(b, '|')
	b = strconv.AppendUint(b, f.slot(id).Load(), 36)
	return string(b)
}
func (f *FlightKeyspace) List(name string) string {
	return "l|" + name + "|" + strconv.FormatUint(f.lists.Load(), 36)
}
func (f *FlightKeyspace) Invalidate(keys ...domain.Slug) {
	for _, key := range keys {
		f.slot(key).Add(1)
	}
	f.lists.Add(1)
}
