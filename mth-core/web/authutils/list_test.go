package authutils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInvalidationList(t *testing.T) {
	r := require.New(t)

	ctx, cancelCtx := context.WithCancel(context.Background())

	il := newInvalidationList(ctx, time.Millisecond)

	testToken := "test_token"
	testNewToken := "test_new_token"

	// No token should exist yet.
	newToken, ok := il.GetNewTokenForOld(testToken)
	r.False(ok)
	r.Empty(newToken)

	testInvalidateFunc := func(ctx context.Context, token string) error {
		return nil
	}

	// Add old token with new.
	il.InvalidateTokenAfter(testToken, testNewToken, -time.Hour, testInvalidateFunc)

	// Old token and new token pair should exist.
	newToken, ok = il.GetNewTokenForOld(testToken)
	r.True(ok)
	r.Equal(testNewToken, newToken)

	// Delete tokens (part of async loop).
	il.deleteTokens(ctx)

	// No token should exist anymore.
	newToken, ok = il.GetNewTokenForOld(testToken)
	r.False(ok)
	r.Empty(newToken)

	cancelCtx()
}
