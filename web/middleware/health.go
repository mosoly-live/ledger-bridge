package middleware

import (
	"net/http"
)

// HealthHandler is simple handler for /health endpoint that returns 200 OK status
func HealthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" && (r.Method == "GET" || r.Method == "HEAD") {
			w.WriteHeader(http.StatusOK)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
