package locale

import (
	"context"
	"net/http"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := r.Header.Get("Accept-Language")
		ctx := context.WithValue(r.Context(), "locale", locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
