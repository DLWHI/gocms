package middleware

import (
	"net/http"
)

type RequestLimiter struct {
	value int64
}

func NewRequestLimiter(limit int64) *RequestLimiter {
	return &RequestLimiter{
		value: limit,
	}
}

func (m *RequestLimiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, m.value)
		next.ServeHTTP(w, r)
	})
}
