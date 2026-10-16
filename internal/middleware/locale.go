package middleware

import (
	"context"
	"net/http"
)

func Locale(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := r.Header.Get("Accept-Language")
		if locale == "" {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), "locale", locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
