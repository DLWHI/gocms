package lang

import (
	"context"
	"net/http"
)

const (
	ContextKey = "lang"
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
		lang := r.PathValue("lang")
		if lang != "" {
			lang, err := NormalizeLangCode(lang)
			if err != nil {
				http.Error(w, "Invalid language code", http.StatusBadRequest)
				return
			}
			if !m.svc.IsSupported(lang) {
				http.Error(w, "Language is not supported", http.StatusNotAcceptable)
				return
			}
			ctx = context.WithValue(ctx, ContextKey, lang)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
