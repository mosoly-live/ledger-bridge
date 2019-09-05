package authutils

import (
	"context"
	"log"
	"sync"
	"time"
)

// InvalidateFunc invalidates token.
type InvalidateFunc func(context.Context, string) error

// TokenInvalidationData contains token invalidation data.
type TokenInvalidationData struct {
	OldToken             string
	NewToken             string
	InvalidateOldTokenAt time.Time
	InvalidateFunc       InvalidateFunc
}

// InvalidationList contains a map which contains mapping from old tokens to new tokens.
// Old tokens expire after a while. This allows us to reuse same new token while old one is being invalidated
// for multiple times.
type InvalidationList struct {
	tokens   map[string]*TokenInvalidationData
	interval time.Duration
	mu       sync.RWMutex
}

// InvalidationListInterface contains invalidation list methods.
type InvalidationListInterface interface {
	GetNewTokenForOld(oldToken string) (newToken string, ok bool)
	InvalidateTokenAfter(token, newToken string, d time.Duration, f InvalidateFunc)
}

// NewInvalidationList creates a new invalidation list.
func NewInvalidationList(ctx context.Context, interval ...time.Duration) *InvalidationList {
	il := newInvalidationList(ctx, interval...)
	go il.runAsync(ctx)
	return il
}

func newInvalidationList(ctx context.Context, interval ...time.Duration) *InvalidationList {
	il := &InvalidationList{
		tokens:   make(map[string]*TokenInvalidationData),
		interval: 3 * time.Second,
	}

	if interval != nil {
		il.interval = interval[0]
	}

	return il
}

func (il *InvalidationList) runAsync(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(il.interval):
			il.deleteTokens(ctx)
		}
	}
}

func (il *InvalidationList) deleteTokens(ctx context.Context) {
	il.mu.Lock()
	defer il.mu.Unlock()

	now := time.Now()

	for key, invalidationData := range il.tokens {
		if now.Before(invalidationData.InvalidateOldTokenAt) {
			continue
		}

		err := invalidationData.InvalidateFunc(ctx, invalidationData.OldToken)
		if err != nil {
			log.Printf("InvalidationList: failed to invalidate old token: %v", err)
		}

		delete(il.tokens, key)
	}
}

// GetNewTokenForOld gets new token for given old token.
func (il *InvalidationList) GetNewTokenForOld(oldToken string) (newToken string, ok bool) {
	il.mu.RLock()
	defer il.mu.RUnlock()

	invalidationData, ok := il.tokens[oldToken]
	if ok {
		return invalidationData.NewToken, true
	}

	return "", false
}

// InvalidateTokenAfter puts token to list. It gets invalidated approximately after given amount of duration.
func (il *InvalidationList) InvalidateTokenAfter(token, newToken string, d time.Duration, f InvalidateFunc) {
	il.mu.Lock()
	defer il.mu.Unlock()

	// Don't put invalidation data again if there exists one for invalidated token already.
	if _, ok := il.tokens[token]; ok {
		return
	}

	il.tokens[token] = &TokenInvalidationData{
		OldToken:             token,
		NewToken:             newToken,
		InvalidateOldTokenAt: time.Now().Add(d),
		InvalidateFunc:       f,
	}
}
