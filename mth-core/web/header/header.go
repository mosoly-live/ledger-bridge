package header

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

// Commonly used headers
const (
	HeaderKeyCorrelationID = "mth-correlation-id"
	HeaderKeyClientIP      = "x-forwarded-for"
	HeaderKeyAuthorization = "authorization"
)

var (
	canonicalHeaderKeyCorrelationID = http.CanonicalHeaderKey(HeaderKeyCorrelationID)
	canonicalHeaderKeyClientIP      = http.CanonicalHeaderKey(HeaderKeyClientIP)
	canonicalHeaderKeyAuthorization = http.CanonicalHeaderKey(HeaderKeyAuthorization)
)

// CorrelationID returns correlation ID header value.
func CorrelationID(r *http.Request) *string {
	if val, ok := r.Header[canonicalHeaderKeyCorrelationID]; ok {
		return &val[0]
	}
	return nil
}

// ClientIP returns client IP using request header. Fallback is remote address of request.
func ClientIP(r *http.Request) string {
	if val, ok := r.Header[canonicalHeaderKeyClientIP]; ok {
		return val[0]
	}
	return r.RemoteAddr
}

// AuthClaims returns authorization token claims from request header.
func AuthClaims(r *http.Request) (claimsMap map[string]interface{}) {
	val, ok := r.Header[canonicalHeaderKeyAuthorization]
	if !ok {
		return
	}
	parts := strings.Split(val[0], ".")
	if len(parts) != 3 {
		return
	}
	claims := parts[1]
	b, _ := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(claims)
	json.Unmarshal(b, &claimsMap)
	return
}
