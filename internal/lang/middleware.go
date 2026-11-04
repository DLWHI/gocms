package lang

import (
	"context"
	"gocms/internal/domain"
	"net/http"

	"golang.org/x/text/language"
)

type Middleware struct {
	svc Service
}

func NewMiddleware(svc Service) *Middleware {
	return &Middleware{
		svc: svc,
	}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		lang := m.svc.DefaultLang().Code
		header := r.Header.Get("Accept-Language")
		tags, _, err := language.ParseAcceptLanguage(header)
		if err != nil {
			http.Error(w, "Invalid 'Accept-Language'", http.StatusBadRequest)
			return
		}
		if len(tags) > 0 {
			base, conf := tags[0].Base()
			if base.String() != "mul" && conf != language.No && m.svc.IsSupported(base.String()) {
				lang = base.String()
			}
		}
		ctx = context.WithValue(ctx, domain.LangContextKey, lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
