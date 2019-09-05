package metrics

import (
	"expvar"
	"math"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	metrics "github.com/rcrowley/go-metrics"
)

var (
	startTime           = time.Now().UTC()
	threadCreateProfile = pprof.Lookup("threadcreate")
)

func init() {
	runtimeMap := expvar.NewMap("runtime")
	runtimeMap.Set("NumGoroutine", expvar.Func(func() interface{} {
		return int64(runtime.NumGoroutine())
	}))
	runtimeMap.Set("NumThread", expvar.Func(func() interface{} {
		return int64(threadCreateProfile.Count())
	}))
	runtimeMap.Set("NumCPU", expvar.Func(func() interface{} {
		return int64(runtime.NumCPU())
	}))
	expvar.Publish("Uptime", expvar.Func(func() interface{} {
		uptime := time.Since(startTime)
		return int64(uptime)
	}))
	expvar.Publish("TimestampMs", expvar.Func(func() interface{} {
		timestampMs := systemUTCUnixMs()
		return int64(timestampMs)
	}))
}

// systemUTCUnixMs returns a Unix time, the number of milliseconds elapsed
// since January 1, 1970 UTC.
func systemUTCUnixMs() int64 {
	return time.Now().UTC().UnixNano() / 1e6
}

type markCallback func(int64)

// Rate tracks the rate of values per second
type Rate struct {
	m  metrics.Meter
	cs []markCallback
}

// NewRate creates a new rate metric
func NewRate() *Rate {
	meter := metrics.NewMeter()
	return &Rate{meter,
		[]markCallback{func(n int64) {
			meter.Mark(n)
		}}}
}

// Mark records the occurance of n events.
func (r *Rate) Mark(n int64) {
	for _, markCallback := range r.cs {
		markCallback(n)
	}
}

func (r *Rate) registerMarkCallback(c markCallback) {
	if c == nil {
		panic("nil rate mark callback")
	}
	r.cs = append(r.cs, c)
}

type timeCallback func(time.Duration)

// Timer capture the duration and rate of events.
type Timer struct {
	t  metrics.Timer
	cs []timeCallback
}

// NewTimer constructs a new Timer using an exponentially-decaying
// sample with the same reservoir size and alpha as UNIX load averages.
func NewTimer() *Timer {
	timer := metrics.NewTimer()
	return &Timer{timer,
		[]timeCallback{func(duration time.Duration) {
			timer.Update(duration)
		}}}
}

// Time record the duration of the execution of the given function.
func (t *Timer) Time(f func()) {
	ts := time.Now()
	f()
	duration := time.Since(ts)
	for _, timeCallback := range t.cs {
		timeCallback(duration)
	}
}

func (t *Timer) registerTimeCallback(c timeCallback) {
	if c == nil {
		panic("nil timer time callback")
	}
	t.cs = append(t.cs, c)
}

type updateCallback func(int64)

// Gauge hold an int64 value that can be set arbitrarily
type Gauge struct {
	g  metrics.Gauge
	cs []updateCallback
}

// NewGauge constructs a new Gauge
func NewGauge() *Gauge {
	gauge := metrics.NewGauge()
	return &Gauge{gauge,
		[]updateCallback{func(v int64) {
			gauge.Update(v)
		}}}
}

// Update updates the gauge's value
func (g *Gauge) Update(v int64) {
	for _, updateCallback := range g.cs {
		updateCallback(v)
	}
}

func (g *Gauge) registerUpdateCallback(c updateCallback) {
	if c == nil {
		panic("nil gauge update callback")
	}
	g.cs = append(g.cs, c)
}

// Value returns the gauge's current value
func (g *Gauge) Value() int64 {
	return g.g.Value()
}

// Histogram calculate distribution statistics from a series of int64 values.
type Histogram struct {
	h  metrics.Histogram
	cs []updateCallback
}

// NewHistogram constructs a new Histogram.
func NewHistogram() *Histogram {
	histogram := metrics.NewHistogram(metrics.NewExpDecaySample(1028, 0.015))
	return &Histogram{histogram,
		[]updateCallback{func(v int64) {
			histogram.Update(v)
		}}}
}

// Update samples a new value.
func (h *Histogram) Update(v int64) {
	for _, updateCallback := range h.cs {
		updateCallback(v)
	}
}

func (h *Histogram) registerUpdateCallback(c updateCallback) {
	if c == nil {
		panic("nil histogram update callback")
	}
	h.cs = append(h.cs, c)
}

// Registry is a registry of all metrics
type Registry struct {
	m     *expvar.Map
	r     metrics.Registry
	pName string // full register name in Prometheus format (includes parent registry name)
}

// NewRegistry creates a new Registry.
// name parameter is used as a root node name
func NewRegistry(name string) *Registry {
	return &Registry{expvar.NewMap(name), metrics.NewPrefixedRegistry(name), toPrometheusName(name)}
}

// CreateSubRegistry creates inner Registry with the give name.
func (r *Registry) CreateSubRegistry(name string) *Registry {
	v := new(expvar.Map).Init()
	r.m.Set(name, v)
	return &Registry{v, metrics.NewPrefixedChildRegistry(r.r, "."+name), r.pName + "_" + toPrometheusName(name)}
}

// RegisterRate registers a new rate metric or returns an error
func (r *Registry) RegisterRate(name string, rate *Rate) error {
	count := new(expvar.Int)
	oneMinute := new(expvar.Float)
	fiveMinute := new(expvar.Float)
	fifteenMinute := new(expvar.Float)
	mean := new(expvar.Float)

	m := new(expvar.Map).Init()
	m.Set("count", count)
	m.Set("one-minute", oneMinute)
	m.Set("five-minute", fiveMinute)
	m.Set("fifteen-minute", fifteenMinute)
	m.Set("mean", mean)

	r.m.Set(name, varFunc(func() expvar.Var {
		s := rate.m.Snapshot()
		count.Set(s.Count())
		oneMinute.Set(s.Rate1())
		fiveMinute.Set(s.Rate5())
		fifteenMinute.Set(s.Rate15())
		mean.Set(s.RateMean())
		return m
	}))

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: r.pName,
		Name:      toPrometheusName(name + "_total"),
		Help:      name,
	})
	prometheus.MustRegister(counter)

	rate.registerMarkCallback(func(v int64) {
		counter.Add(float64(v))
	})

	return r.r.Register("."+name, rate.m)
}

var (
	percentileNames = []struct {
		name string
		perc float64
	}{
		{"percentile-50", 0.5},
		{"percentile-75", 0.75},
		{"percentile-95", 0.95},
		{"percentile-99", 0.99},
		{"percentile-999", 0.999},
	}
	// percentiles will be initialized in init
	percentiles = make([]float64, len(percentileNames))
)

func init() {
	// initializing percentiles
	for i, tp := range percentileNames {
		percentiles[i] = tp.perc
	}
}

// RegisterTimer registers a new timer under the given name or returns an error
func (r *Registry) RegisterTimer(name string, timer *Timer) error {
	count := new(expvar.Int)
	min := new(expvar.Int)
	max := new(expvar.Int)
	mean := new(expvar.Int)
	stdDev := new(expvar.Int)
	oneMinute := new(expvar.Float)
	fiveMinute := new(expvar.Float)
	fifteenMinute := new(expvar.Float)
	meanRate := new(expvar.Float)
	percentileVars := make([]*expvar.Int, len(percentileNames))

	m := new(expvar.Map).Init()
	m.Set("count", count)
	m.Set("min", min)
	m.Set("max", max)
	m.Set("mean", mean)
	m.Set("std-dev", stdDev)
	m.Set("one-minute", oneMinute)
	m.Set("five-minute", fiveMinute)
	m.Set("fifteen-minute", fifteenMinute)
	m.Set("mean-rate", meanRate)
	for i, tp := range percentileNames {
		v := new(expvar.Int)
		percentileVars[i] = v
		m.Set(tp.name, v)
	}

	r.m.Set(name, varFunc(func() expvar.Var {
		s := timer.t.Snapshot()
		count.Set(s.Count())
		min.Set(s.Min())
		max.Set(s.Max())
		mean.Set(ceilTime(s.Mean()))
		stdDev.Set(ceilTime(s.StdDev()))
		oneMinute.Set(s.Rate1())
		fiveMinute.Set(s.Rate5())
		fifteenMinute.Set(s.Rate15())
		meanRate.Set(s.RateMean())
		ps := s.Percentiles(percentiles)
		for i, pv := range ps {
			percentileVars[i].Set(ceilTime(pv))
		}
		return m
	}))

	summary := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: r.pName,
		Name:      toPrometheusName(name + "_seconds"),
		Help:      name,
	})
	prometheus.MustRegister(summary)

	timer.registerTimeCallback(func(duration time.Duration) {
		summary.Observe(duration.Seconds())
	})

	return r.r.Register("."+name, timer.t)
}

// RegisterGauge registers a new gauge under the given name or returns an error
func (r *Registry) RegisterGauge(name string, gauge *Gauge, units string) error {
	if len(units) == 0 {
		panic("expected non empty units")
	}

	value := new(expvar.Int)

	r.m.Set(name, varFunc(func() expvar.Var {
		s := gauge.g.Snapshot()
		value.Set(s.Value())
		return value
	}))

	pgauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: r.pName,
		Name:      toPrometheusName(name + "_" + units),
		Help:      name,
	})
	prometheus.MustRegister(pgauge)

	gauge.registerUpdateCallback(func(v int64) {
		pgauge.Set(float64(v))
	})

	return r.r.Register("."+name, gauge.g)
}

func ceilTime(v float64) int64 {
	return int64(math.Ceil(v))
}

type varFunc func() expvar.Var

func (f varFunc) String() string {
	return f().String()
}

// GetMetricsRegistry returns global registry of metrics used in the solution
func (r *Registry) GetMetricsRegistry() metrics.Registry {
	return r.r
}

// RegisterHistogram registers a new histogram under the given name or returns an error
func (r *Registry) RegisterHistogram(name string, histogram *Histogram, units string) error {
	if len(units) == 0 {
		panic("expected non empty units")
	}

	count := new(expvar.Int)
	min := new(expvar.Int)
	max := new(expvar.Int)
	mean := new(expvar.Float)
	stdDev := new(expvar.Float)
	percentileVars := make([]*expvar.Float, len(percentileNames))

	m := new(expvar.Map).Init()
	m.Set("count", count)
	m.Set("min", min)
	m.Set("max", max)
	m.Set("mean", mean)
	m.Set("std-dev", stdDev)
	for i, tp := range percentileNames {
		v := new(expvar.Float)
		percentileVars[i] = v
		m.Set(tp.name, v)
	}

	r.m.Set(name, varFunc(func() expvar.Var {
		s := histogram.h.Snapshot()
		count.Set(s.Count())
		min.Set(s.Min())
		max.Set(s.Max())
		mean.Set(s.Mean())
		stdDev.Set(s.StdDev())
		ps := s.Percentiles(percentiles)
		for i, pv := range ps {
			percentileVars[i].Set(pv)
		}
		return m
	}))

	summary := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: r.pName,
		Name:      toPrometheusName(name + "_" + units),
		Help:      name,
	})
	prometheus.MustRegister(summary)

	histogram.registerUpdateCallback(func(v int64) {
		summary.Observe(float64(v))
	})

	return r.r.Register("."+name, histogram.h)
}

func toPrometheusName(name string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ':
			return '_'
		case '.':
			return '_'
		default:
			return r
		}
	}, name)
}
