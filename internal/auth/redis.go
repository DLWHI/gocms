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

var revoke = redis.NewScript(`
	local userKey = KEYS[1]
	local newSessionKey = KEYS[2]
	local newSessionData = ARGV[1]
	local newSuid = ARGV[2]
	local ttl = tonumber(ARGV[3])

	local oldSuid = redis.call("GET", userKey)

	if oldSuid and oldSuid ~= "" then
			redis.call("DEL", "cms:sessions:" .. oldSuid)
	end

	redis.call("SET", newSessionKey, newSessionData, "EX", ttl)
	redis.call("SET", userKey, newSuid, "EX", ttl)

	return "OK"
`)

type RedisSessionRepository struct {
	cache *cache.RedisClient
	users UserService
	keys  cache.Keyspace
	ttl   time.Duration
}

func NewRedisRepository(client *redis.Client, users UserService, keys cache.Keyspace, ttl time.Duration) *RedisSessionRepository {
	return &RedisSessionRepository{
		users: users,
		ttl:   ttl,
		keys:  keys,
		cache: &cache.RedisClient{
			Client:   client,
			LeaseTTL: 5 * time.Second,
		},
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
	hash := sha256.Sum256(secret)

	if _, err := s.cache.Client.Set(ctx, key, hash[:], s.ttl).Result(); err != nil {
		return "", err
	}

	b64 := base64.RawURLEncoding.EncodeToString(secret)
	clear(secret)

	return username + "." + suid + "." + b64, nil
}

func (s *RedisSessionRepository) Revoke(ctx context.Context, username string, suid string) error {
	key := s.keys.Key(username + "." + suid)

	if _, err := s.cache.Client.Del(ctx, key).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}
	return nil
}

func (s *RedisSessionRepository) RevokeUser(ctx context.Context, username string) error {
	pattern := s.keys.Key(username + ".*")

	if err := s.cache.InvalidatePattern(ctx, pattern); err != nil {
		return err
	}
	return nil
}

func (s *RedisSessionRepository) Validate(ctx context.Context, token string) (string, error) {
	parts := strings.Split(token, ".")

	if len(parts) != 3 {
		return "", fmt.Errorf("%w: invalid token", domain.ErrInval)
	}

	username := parts[0]
	suid := parts[1]

	key := s.keys.Key(username + "." + suid)

	store, err := s.cache.Client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", domain.ErrExpired
		}
		return "", err
	}

	secret, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("%w: invalid token", domain.ErrInval)
	}

	hash := sha256.Sum256(secret)
	clear(secret)

	match := subtle.ConstantTimeCompare(store, hash[:]) == 1
	clear(hash[:])
	clear(secret)
	if !match {
		return "", domain.ErrExpired
	}

	return username, nil
}
