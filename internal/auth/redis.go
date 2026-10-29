package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"gocms/internal/cache"
	"gocms/internal/domain"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	createScript = redis.NewScript(`
		redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2])
		for _, k in ipairs(redis.call("SMEMBERS", KEYS[2])) do
			if redis.call("EXISTS", k) == 0 then redis.call("SREM", KEYS[2], k) end
		end
		redis.call("SADD", KEYS[2], KEYS[1])
		redis.call("EXPIRE", KEYS[2], ARGV[2])
		return 1
	`)

	revokeUserScript = redis.NewScript(`
		local ids = redis.call("SMEMBERS", KEYS[1])
		for _, id in ipairs(ids) do redis.call("DEL", id) end
		redis.call("DEL", KEYS[1])
		return #ids
	`)

	revokeIdScript = redis.NewScript(`
		redis.call("DEL", KEYS[1])
		redis.call("SREM", KEYS[2], KEYS[1])
		return 1
	`)
)

type RedisSessionRepository struct {
	client *redis.Client
	keys   cache.Keyspace
	ttl    time.Duration
}

func NewRedisSessionRepository(client *redis.Client, keys cache.Keyspace, ttl time.Duration) *RedisSessionRepository {
	return &RedisSessionRepository{
		ttl:    ttl,
		keys:   keys,
		client: client,
	}
}

func (s *RedisSessionRepository) GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand error: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *RedisSessionRepository) GenerateSecret() ([]byte, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("crypto/rand error: %w", err)
	}
	return b, nil
}

func (s *RedisSessionRepository) Create(ctx context.Context, username string) (string, error) {
	suid, err := s.GenerateSessionID()
	if err != nil {
		return "", err
	}

	secret, err := s.GenerateSecret()
	if err != nil {

		return "", err
	}

	key := s.keys.Key(username + "." + suid)
	list := s.keys.List(username)
	hash := sha256.Sum256(secret)

	if _, err := createScript.Run(ctx, s.client, []string{key, list}, hash[:], int64(s.ttl.Seconds())).Result(); err != nil {
		return "", err
	}

	b64 := base64.RawURLEncoding.EncodeToString(secret)
	clear(secret)

	return username + "." + suid + "." + b64, nil
}

func (s *RedisSessionRepository) Revoke(ctx context.Context, username string, suid string) error {
	key := s.keys.Key(username + "." + suid)
	list := s.keys.List(username)

	if _, err := revokeIdScript.Run(ctx, s.client, []string{key, list}).Result(); err != nil {
		return err
	}
	return nil
}

func (s *RedisSessionRepository) RevokeUser(ctx context.Context, username string) error {
	list := s.keys.List(username)

	if _, err := revokeUserScript.Run(ctx, s.client, []string{list}).Result(); err != nil {
		return err
	}
	return nil
}

func (s *RedisSessionRepository) Validate(ctx context.Context, token string) (string, string, error) {
	parts := strings.Split(token, ".")

	if len(parts) != 3 {
		return "", "", fmt.Errorf("%w: invalid token", domain.ErrInval)
	}

	username := strings.Clone(parts[0])
	suid := strings.Clone(parts[1])

	key := s.keys.Key(username + "." + suid)

	store, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", "", domain.ErrExpired
		}
		return "", "", err
	}

	secret, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", "", fmt.Errorf("%w: invalid token", domain.ErrInval)
	}

	hash := sha256.Sum256(secret)
	clear(secret)

	match := subtle.ConstantTimeCompare(store, hash[:]) == 1
	clear(hash[:])
	clear(secret)
	if !match {
		return "", "", domain.ErrExpired
	}

	return username, suid, nil
}
