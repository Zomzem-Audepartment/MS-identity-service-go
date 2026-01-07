package http

import (
	"net/http"
)

func InternalAPIKeyMiddleware(apiKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for health check
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-Internal-API-Key")
			if key != apiKey {
				http.Error(w, "Forbidden: Invalid Internal API Key", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
