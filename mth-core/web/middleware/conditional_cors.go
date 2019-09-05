package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// ConditionalCors is a middleware that handles CORS headers if they were not already handled on gateway (i.e. Tyk)
type ConditionalCors struct {

	// CorsHandledHeader specifies a http request header name whose existence determines whether header was already handled on gateway
	CorsHandledHeader string
	cors              *cors.Cors
}

const defaultCorsHandledHeader = "cf-ray"

// NewConditionalCors creates new ConditionalCors handler
func NewConditionalCors(c *cors.Cors) *ConditionalCors {
	return &ConditionalCors{
		cors:              c,
		CorsHandledHeader: defaultCorsHandledHeader,
	}
}

// Handler apply the CORS specification on the request, and add relevant CORS headers
// as necessary
func (tc *ConditionalCors) Handler(h http.Handler) http.Handler {
	corsHandlerFunc := tc.cors.Handler(h)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Handle CORS if not already handled
		if !tc.wasCorsAlreadyHandled(r) {
			corsHandlerFunc.ServeHTTP(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (tc *ConditionalCors) wasCorsAlreadyHandled(r *http.Request) bool {
	return (r.Header.Get(tc.CorsHandledHeader) != "")
}
