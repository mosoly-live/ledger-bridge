package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// RecoverHandler handles panic (writes panic message to log)
func RecoverHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v\n%v", err, string(debug.Stack()))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		h.ServeHTTP(w, r)
	})
}
