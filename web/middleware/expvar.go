package middleware

import (
	"expvar"
	"net/http"
)

// ExpVarHandler returns the expvar HTTP Handler
func ExpVarHandler(path string, h http.Handler) http.Handler {
	eh := expvar.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path && r.Method == "GET" {
			eh.ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
