package lang

import (
	"context"
	"gocms/internal/domain"
	"net/http"

	"golang.org/x/text/language"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var header string
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			header = r.Header.Get("Content-Language")
		case http.MethodGet:
			header = r.Header.Get("Accept-Language")
		}

		if header != "" {
			tags, _, err := language.ParseAcceptLanguage(header)
			if err == nil {
				base, conf := tags[0].Base()
				if conf != language.No {
					ctx = context.WithValue(ctx, domain.LangKey, base.String())
				}
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
