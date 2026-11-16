package lang

import (
	"context"
	"net/http"

	"golang.org/x/text/language"
)

const (
	ContextKey = "lang"
)

type PathMiddleware struct {
	svc Service
}

func NewPathMiddleware(svc Service) *PathMiddleware {
	return &PathMiddleware{
		svc: svc,
	}
}

func (m *PathMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		language := r.PathValue("lang")
		if language != "" {
			language, err := NormalizeLangCode(language)
			if err != nil {
				http.Error(w, "Invalid language code", http.StatusBadRequest)
				return
			}
			if !m.svc.IsSupported(language) {
				http.Error(w, "Language is not supported", http.StatusNotAcceptable)
				return
			}
			ctx = context.WithValue(ctx, ContextKey, language)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type HeaderMiddleware struct {
	svc Service
}

func NewHeaderMiddleware(svc Service) *HeaderMiddleware {
	return &HeaderMiddleware{
		svc: svc,
	}
}

func (m *HeaderMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tags, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
		if err != nil {
			http.Error(w, "Invalid Accept-Language header value", http.StatusBadRequest)
			return
		}
		if len(tags) == 0 {
			http.Error(w, "Please specify Accept-Language", http.StatusBadRequest)
			return
		}
		base, confidence := tags[0].Base()
		if confidence == language.No {
			http.Error(w, "Very ambigous Accept-Language value, please, resolve language", http.StatusNotAcceptable)
			return
		}
		code := base.String()
		if !m.svc.IsSupported(code) {
			http.Error(w, "Please specify Accept-Language", http.StatusNotAcceptable)
			return
		}
		ctx = context.WithValue(ctx, ContextKey, code)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
