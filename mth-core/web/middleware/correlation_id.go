package middleware

import (
	"net/http"

	uuid "github.com/satori/go.uuid"
	webcontext "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/web/context"
	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/web/header"
)

var maxCorrelationIDLength = 200

// CorrelationIDHandler adds or forward mth-correlation-id header
func CorrelationIDHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := getOrCreateNewCorrelationID(r)
		w.Header().Set(header.HeaderKeyCorrelationID, correlationID)
		ctx := webcontext.WithCorrelationID(r.Context(), correlationID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getOrCreateNewCorrelationID(r *http.Request) string {
	if cid := header.CorrelationID(r); cid != nil {
		return truncateString(*cid, maxCorrelationIDLength)
	}
	return uuid.NewV4().String()
}

func truncateString(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return string(runes[:i])
	}
	return s
}
