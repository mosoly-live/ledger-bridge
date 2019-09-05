package middleware

import (
	"net/http"

	"gitlab.com/p-invent/mosoly-ledger-bridge/metrics"
	"gitlab.com/p-invent/mosoly-ledger-bridge/web"
)

// MetricsHandler is a handler that creates counters for client and server http errors
func MetricsHandler(h http.Handler, r *metrics.Registry) http.Handler {
	clientErrorRate := metrics.NewRate()
	serverErrorRate := metrics.NewRate()

	if r != nil {
		r.RegisterRate("client_errors", clientErrorRate)
		r.RegisterRate("server_errors", serverErrorRate)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				serverErrorRate.Mark(1)
				panic(p) // re-throw panic
			}
		}()

		lrw := web.NewLogStatusReponseWriter(w)
		h.ServeHTTP(lrw, r)

		httpStatus := lrw.Status()
		if httpStatus >= 500 {
			serverErrorRate.Mark(1)
		} else if httpStatus >= 400 {
			clientErrorRate.Mark(1)
		}
	})
}
