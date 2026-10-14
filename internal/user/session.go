package user

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"gocms/internal/errs"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type RedisSessionStorage struct {
	rdb    *redis.Client
	ttl    time.Duration
	revoke *redis.Script
	group  singleflight.Group
}

func NewRedisSessionStorage(client *redis.Client, ttl time.Duration) *RedisSessionStorage {
	return &RedisSessionStorage{
		rdb: client,
		ttl: ttl,
		revoke: redis.NewScript(`
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
		`),
	}
}

func GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand error: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *RedisSessionStorage) Create(ctx context.Context, user *Model) (*Session, error) {
	ch := s.group.DoChan(user.Username, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		suid, err := GenerateSessionID()
		if err != nil {
			s.group.Forget(user.Username)
			return nil, err
		}

		session := Session{
			ID:   suid,
			User: *user,
		}

		json, err := json.Marshal(session)
		if err != nil {
			s.group.Forget(user.Username)
			return nil, err
		}

		sessionKey := fmt.Sprintf("cms:sessions:%s", suid)
		userKey := fmt.Sprintf("cms:users:%s:session", user.Username)

		keys := []string{userKey, sessionKey}
		args := []interface{}{json, suid, int(s.ttl.Seconds())}

		if err := s.revoke.Run(bg, s.rdb, keys, args...).Err(); err != nil {
			s.group.Forget(user.Username)
			return nil, err
		}

		return &session, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.(*Session), nil
	}
}

func (s *RedisSessionStorage) Update(ctx context.Context, user *Model) (*Session, error) {
	userKey := fmt.Sprintf("cms:users:%s:session", user.Username)

	cached, err := s.rdb.Get(ctx, userKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errs.ErrExpired
		}
		return nil, err
	}

	var session Session
	if err := json.Unmarshal([]byte(cached), &session); err != nil {
		return nil, err
	}

	session.User = *user

	json, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	sessionKey := fmt.Sprintf("cms:sessions:%s", session.ID)

	elapsed := time.Since(*session.User.LastLogin)
	if elapsed >= s.ttl {
		expire := time.Now().Add(elapsed)
		pipe := s.rdb.Pipeline()
		pipe.Set(ctx, sessionKey, json, time.Until(expire))
		pipe.Set(ctx, userKey, json, time.Until(expire))
		if _, err := pipe.Exec(ctx); err != nil {
			return nil, err
		}
	}

	return &session, nil
}

func (s *RedisSessionStorage) Revoke(ctx context.Context, username string, suid string) error {
	keys := []string{
		fmt.Sprintf("cms:sessions:%s", suid),
		fmt.Sprintf("cms:users:%s:session", username),
	}
	if _, err := s.rdb.Del(ctx, keys...).Result(); err != nil {
		return err
	}
	return nil
}

func (s *RedisSessionStorage) RevokeID(ctx context.Context, suid string) error {
	sessionKey := fmt.Sprintf("cms:sessions:%s", suid)
	cached, err := s.rdb.Get(ctx, sessionKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}
	var session Session
	if err := json.Unmarshal([]byte(cached), &session); err != nil {
		return err
	}

	userKey := fmt.Sprintf("cms:users:%s:session", session.User.Username)
	if _, err := s.rdb.Del(ctx, sessionKey, userKey).Result(); err != nil {
		return err
	}
	return nil
}

func (s *RedisSessionStorage) RevokeUser(ctx context.Context, username string) error {
	userKey := fmt.Sprintf("cms:users:%s:session", username)
	suid, err := s.rdb.Get(ctx, userKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}

	sessionKey := fmt.Sprintf("cms:sessions:%s", suid)
	if _, err := s.rdb.Del(ctx, sessionKey, userKey).Result(); err != nil {
		return err
	}
	return nil
}

func (s *RedisSessionStorage) Validate(ctx context.Context, suid string) (*Session, error) {
	key := fmt.Sprintf("cms:sessions:%s", suid)
	cached, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errs.ErrExpired
		}
		return nil, err
	}
	var session Session
	if err := json.Unmarshal([]byte(cached), &session); err != nil {
		return nil, err
	}
	return &session, nil
}
