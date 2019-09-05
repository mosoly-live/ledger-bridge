package middleware

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusHandler adds default prometheus handler middleware
func PrometheusHandler(path string, h http.Handler) http.Handler {
	prometheusHandler := promhttp.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path && r.Method == "GET" {
			prometheusHandler.ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
