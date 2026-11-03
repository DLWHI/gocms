package lang

import (
	"context"
	"gocms/internal/domain"
	"net/http"
	"strings"

	"golang.org/x/text/language"
)

// Determines request content language (what it got)
type RequestMiddleware struct {
	svc Service
}

// Determines target response language (what it should answer)
type ResponseMiddleware struct {
	svc Service
}

func NewRequestMiddleware(svc Service) *RequestMiddleware {
	return &RequestMiddleware{
		svc: svc,
	}
}

func NewResponseMiddleware(svc Service) *ResponseMiddleware {
	return &ResponseMiddleware{
		svc: svc,
	}
}

func (m *RequestMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		header := r.Header.Get("Content-Language")
		tags := strings.Split(header, ",")
		if len(tags) != 1 {
			http.Error(w, "Ambiguous content language", http.StatusBadRequest)
			return
		}
		base, err := ToISO639_1Code(tags[0])
		if err == nil && m.svc.IsSupported(base) {
			ctx = context.WithValue(ctx, domain.LangContextKey, strings.ToLower(base))
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *ResponseMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
