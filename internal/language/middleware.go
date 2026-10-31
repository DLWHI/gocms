package language

import (
	"context"
	"gocms/internal/domain"
	"net/http"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// acceptHeader := r.Header.Get("Accept-Language")

		// tags, _, err := language.ParseAcceptLanguage(acceptHeader)

		// var tag language.Tag
		// if err != nil || len(tags) == 0 {
		// 	matchedTag = supportedLanguages[0]
		// } else {
		// 	matchedTag, _, _ = matcher.Match(tags...)
		// }

		language, _ := domain.NormalizeLangCode(r.Header.Get("Accept-Language"))
		ctx := context.WithValue(r.Context(), "language", language)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
