package kinesis

import (
	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/data/metrics"
)

// AddEventPutterMetrics adds metrics by wrapping handler
func AddEventPutterMetrics(p EventPutter, r *metrics.Registry) *MetricsPutter {
	putRate := metrics.NewRate()

	if r != nil {
		r.RegisterRate("put", putRate)
	}

	return &MetricsPutter{
		p:       p,
		putRate: putRate,
	}
}

// MetricsPutter wraps putter and adds metrics
type MetricsPutter struct {
	p       EventPutter
	putRate *metrics.Rate
}

// Put puts event asynchronously to stream. This method is thread-safe.
func (p *MetricsPutter) Put(e *Event) (err error) {
	p.putRate.Mark(1)

	return p.p.Put(e)
}
