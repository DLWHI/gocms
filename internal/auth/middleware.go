package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	BearerPrefix = "Bearer "
)

type JWTMiddleware struct {
	jwtSecret *rsa.PublicKey
}

type SessionMiddleware struct {
	provider SessionRepository
}

type TestMiddleware struct {
	username string
	level    uint32
}

func NewJWTMiddleware(jwtSecret *rsa.PublicKey) *JWTMiddleware {
	return &JWTMiddleware{
		jwtSecret: jwtSecret,
	}
}

func NewSessionMiddleware(provider SessionRepository) *SessionMiddleware {
	return &SessionMiddleware{
		provider: provider,
	}
}

func NewTestMiddleware(username string, level uint32) *TestMiddleware {
	return &TestMiddleware{
		username: username,
		level:    level,
	}
}

func (m *JWTMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			http.Error(w, "Unauthorized: unknown bearer", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, prefix)

		var claims JWTClaims
		token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
			}
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Unauthorized: token has expired", http.StatusUnauthorized)
			} else {
				http.Error(w, "Unauthorized: invalid token signature", http.StatusUnauthorized)
			}
			logger.Error("Middleware.JWTAuth", "Failed to authenticate request token", err)
			return
		}

		ctx := context.WithValue(r.Context(), domain.UsernameKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *SessionMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			http.Error(w, "Unauthorized: unknown bearer", http.StatusUnauthorized)
			return
		}
		sessionID := strings.TrimPrefix(authHeader, BearerPrefix)
		ctx := r.Context()

		username, session, err := m.provider.Validate(ctx, sessionID)
		if err != nil {
			if errors.Is(err, domain.ErrExpired) {
				http.Error(w, "Unauthorized: Expired or invalid session", http.StatusUnauthorized)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				logger.Error("Middleware.SessionAuth", "Failed to validate session", err)
			}
			return
		}

		ctx = context.WithValue(ctx, domain.UsernameKey, username)
		ctx = context.WithValue(ctx, domain.SessionIDKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *TestMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, domain.UsernameKey, m.username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func HasPermission(ctx context.Context, entity string, action string) bool {
	return false
}

func IsSubsetOf(ctx context.Context, permissions uint32) bool {
	return false
}
