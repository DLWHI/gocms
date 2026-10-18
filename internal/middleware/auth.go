package middleware

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/session"
	"gocms/internal/tokens"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey string = "username"
const UserRoleKey string = "userperm"

type JWTAuth struct {
	jwtSecret *rsa.PublicKey
}

type SessionAuth struct {
	provider session.Repository
}

type TestAuth struct {
	username string
	level    uint32
}

func NewJWTAuth(jwtSecret *rsa.PublicKey) *JWTAuth {
	return &JWTAuth{
		jwtSecret: jwtSecret,
	}
}

func NewSessionAuth(provider session.Repository) *SessionAuth {
	return &SessionAuth{
		provider: provider,
	}
}

func NewTestAuth(username string, level uint32) *TestAuth {
	return &TestAuth{
		username: username,
		level:    level,
	}
}

func (m *JWTAuth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			http.Error(w, "Unauthorized: unkown bearer", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, prefix)

		var claims tokens.JWTClaims
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

		ctx := context.WithValue(r.Context(), UserIDKey, claims.Username)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Permissions)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *SessionAuth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			http.Error(w, "Unauthorized: unkown bearer", http.StatusUnauthorized)
			return
		}
		sessionID := strings.TrimPrefix(authHeader, prefix)
		ctx := r.Context()

		session, err := m.provider.Validate(ctx, sessionID)
		if err != nil {
			if errors.Is(err, errs.ErrExpired) {
				http.Error(w, "Unauthorized: Expired or invalid session", http.StatusUnauthorized)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				logger.Error("Middleware.SessionAuth", "Failed to validate session", err)
			}
			return
		}

		ctx = context.WithValue(ctx, UserIDKey, session.User.Username)
		ctx = context.WithValue(ctx, UserRoleKey, session.User.Permissions)
		ctx = context.WithValue(ctx, "session", session.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *TestAuth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, m.username)
		ctx = context.WithValue(ctx, UserRoleKey, m.level)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
