package lang

import (
	"context"
	"gocms/internal/domain"
	"net/http"
	"strings"

	"golang.org/x/text/language"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			header := r.Header.Get("Content-Language")
			tags := strings.Split(header, ",")
			if len(tags) == 1 {
				base, err := domain.ToISO639_1Code(tags[0])
				if err == nil {
					ctx = context.WithValue(ctx, domain.LangKey, strings.ToLower(base))
				}
			}
		case http.MethodGet:
			header := r.Header.Get("Accept-Language")
			tags, _, err := language.ParseAcceptLanguage(header)
			if err == nil && len(tags) > 0 {
				base, conf := tags[0].Base()
				if base.String() == "*" {
					http.Error(w, "Fallback language is not accepted", http.StatusNotAcceptable)
				}
				if conf != language.No {
					ctx = context.WithValue(ctx, domain.LangKey, base.String())
				}
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
