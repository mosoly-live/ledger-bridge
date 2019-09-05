package log

import (
	"net/http"

	corelog "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/log"
	webcontext "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/web/context"
	"go.uber.org/zap"
)

// RequestLogger adds zap fields to log http request
func RequestLogger(r *http.Request) *corelog.Logger {
	l := webcontext.NewLogger(r.Context())
	return l.With(corelog.HTTPHeader(r.Header),
		corelog.RequestURL(*r.URL),
		corelog.RequestMethod(r.Method))
}

// ResponseLogger adds zap fields to log http response
func ResponseLogger(r *http.Response) *corelog.Logger {
	if r != nil {
		return corelog.With(corelog.HTTPHeader(r.Header),
			corelog.StatusCode(r.StatusCode))
	}

	return corelog.With(zap.Any("http_response", nil))
}
