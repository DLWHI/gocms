package lang

import (
	"context"
	"gocms/internal/domain"
	"net/http"
	"strings"

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
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			header := r.Header.Get("Content-Language")
			tags := strings.Split(header, ",")
			if len(tags) != 1 {
				http.Error(w, "Ambiguous content language", http.StatusNotAcceptable)
				return
			}
			base, err := ToISO639_1Code(tags[0])
			if err == nil && m.svc.IsSupported(base) {
				ctx = context.WithValue(ctx, domain.LangContextKey, strings.ToLower(base))
			}
		case http.MethodGet:
			lang := m.svc.DefaultLang().Code
			header := r.Header.Get("Accept-Language")
			tags, _, err := language.ParseAcceptLanguage(header)
			if err == nil && len(tags) > 0 {
				base, conf := tags[0].Base()
				if base.String() != "mul" {
					if conf != language.No {
						lang = base.String()
					}
				}
			}
			ctx = context.WithValue(ctx, domain.LangContextKey, lang)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
