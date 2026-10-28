package cache

import (
	"fmt"
	"gocms/internal/domain"
	"regexp"
	"strings"
)

var whitelist = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
var escaper = strings.NewReplacer("%", "%25", "{", "%7B", "}", "%7D")

type KeyConfig struct {
	Prefix  []string
	Version string
}

type Keyspace struct {
	base string
}

type HashedKeyspace struct {
	Keyspace
}

func NewKeyspace(cfg KeyConfig, entity string) (*Keyspace, error) {
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
	return &Keyspace{base: strings.Join(parts, ":")}, nil
}

func NewHashedKeyspace(cfg KeyConfig, entity string) (*HashedKeyspace, error) {
	keyspace, err := NewKeyspace(cfg, entity)
	if err != nil {
		return nil, err
	}
	return &HashedKeyspace{Keyspace: *keyspace}, nil
}

func (k *Keyspace) Key(id string) string    { return k.base + ":item:" + id }
func (k *Keyspace) List(name string) string { return k.base + ":list:" + name }

// '{', '}' and '%' are escaped with their web code
func (k *HashedKeyspace) Key(id string) string { return k.base + ":item:{" + escaper.Replace(id) + "}" }
func (k *HashedKeyspace) List(name string) string {
	return k.base + ":list:{" + escaper.Replace(name) + "}"
}
