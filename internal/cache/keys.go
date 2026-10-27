package cache

import (
	"fmt"
	"gocms/internal/domain"
	"strings"
)

type KeyConfig struct {
	Prefix  string
	Version string
	HashTag bool
}

type Keyspace struct{ base string }

func NewKeyspace(cfg KeyConfig, entity string) (*Keyspace, error) {
	if cfg.Prefix == "" || entity == "" {
		return nil, fmt.Errorf("%w: prefix and entity are required to create keyspace", domain.ErrInval)
	}
	for _, s := range []string{cfg.Prefix, cfg.Version, entity} {
		if strings.ContainsAny(s, " \t\r\n{}:") {
			return nil, fmt.Errorf("%w: key parts must not contain whitespace, braces or ':'", domain.ErrInval)
		}
	}
	base := cfg.Prefix + ":" + entity
	if cfg.Version != "" {
		base += ":" + cfg.Version
	}
	if cfg.HashTag {
		base = "{" + base + "}"
	}
	return &Keyspace{base: base}, nil
}

func (k Keyspace) Item(id string) string   { return k.base + ":i:" + id }
func (k Keyspace) Gen(id string) string    { return k.base + ":gen:" + id }
func (k Keyspace) List(name string) string { return k.base + ":l:" + name }
