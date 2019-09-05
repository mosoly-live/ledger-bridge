package web

import (
	"context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type requestBodyContextKey int

const (
	// bodyBytesContextKey is the context key name of the body bytes id
	bodyBytesContextKey requestBodyContextKey = iota
)

// RequestBody returns correlation-id from request context
func RequestBody(ctx context.Context) []byte {
	return ctx.Value(bodyBytesContextKey).([]byte)
}

// WithBody returns a new Context carrying body
func WithBody(ctx context.Context, body []byte) context.Context {
	return context.WithValue(ctx, bodyBytesContextKey, body)
}
