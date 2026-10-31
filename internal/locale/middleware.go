package locale

import (
	"context"
	"gocms/internal/domain"
	"net/http"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale, _ := domain.NormalizeLocaleCode(r.Header.Get("Accept-Language"))
		ctx := context.WithValue(r.Context(), "locale", locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
