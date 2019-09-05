package web

import "context"

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type correlationIDContextKey int

const (
	// CorrelationIDContextKey is the context key name of the correlation id
	contextKey correlationIDContextKey = iota
)

// CorrelationID returns correlation ID from request context.
func CorrelationID(ctx context.Context) (cID string) {
	cID, _ = ctx.Value(contextKey).(string)
	return
}

// WithCorrelationID returns a new Context carrying correlation ID.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, contextKey, correlationID)
}
