package server

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
	return http.MaxBytesHandler(next, m.value)
}
