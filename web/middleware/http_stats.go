package middleware

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPStatsHandler is a handler that handles stats of HTTP requests
func HTTPStatsHandler(h http.Handler) http.Handler {
	const httpNamespace = "http"

	gauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: httpNamespace,
			Name:      "in_flight_requests",
			Help:      "A gauge of HTTP requests currently being served.",
		})

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: httpNamespace,
			Name:      "requests_total",
			Help:      "A counter for HTTP requests.",
		},
		[]string{"code", "method"},
	)

	duration := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: httpNamespace,
			Name:      "requests_duration_seconds",
			Help:      "A summary of latencies for HTTP requests.",
		},
		[]string{"method"},
	)

	writeHeader := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: httpNamespace,
			Name:      "write_header_duration_seconds",
			Help:      "A histogram of time to first write latencies.",
		},
		[]string{"method"},
	)

	requestSize := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: httpNamespace,
			Name:      "request_size_bytes",
			Help:      "A histogram of HTTP request sizes.",
		},
		[]string{},
	)

	responseSize := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: httpNamespace,
			Name:      "response_size_bytes",
			Help:      "A histogram of HTTP response sizes.",
		},
		[]string{},
	)

	prometheus.MustRegister(gauge, counter, duration, writeHeader, requestSize, responseSize)

	return promhttp.InstrumentHandlerInFlight(gauge,
		promhttp.InstrumentHandlerCounter(counter,
			promhttp.InstrumentHandlerDuration(duration,
				promhttp.InstrumentHandlerTimeToWriteHeader(writeHeader,
					promhttp.InstrumentHandlerRequestSize(requestSize,
						promhttp.InstrumentHandlerResponseSize(responseSize,
							h),
					),
				),
			),
		),
	)
}
