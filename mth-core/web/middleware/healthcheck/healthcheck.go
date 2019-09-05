package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/log"
)

var (
	// Timeout is check timeout.
	Timeout = time.Minute
	// MaxFailureInARow is the number for when a dependency is considered broken/down.
	MaxFailureInARow = 3
	// AsyncLoggingInterval is the interval for async logging of failing dependencies.
	AsyncLoggingInterval = time.Minute
	// dependencies are each of the dependencies which are needed to be checked in order to
	// be able to say that service is completely healthy.
	dependencies []*dependency
	// enabledDependencies are assigned to dependencies on start.
	enabledDependencies = []*dependency{}
)

// HealthChecker checks health.
type HealthChecker interface {
	CheckHealth() error
}

// dependency is a microservice dependency, which is registered and health checked.
type dependency struct {
	Name     string
	Checker  HealthChecker
	LastErr  error
	Interval time.Duration

	FailureInARow int

	sync.RWMutex
}

// AddDependency adds a health checked dependency.
func AddDependency(name string, checker HealthChecker, interval time.Duration) {
	enabledDependencies = append(enabledDependencies, &dependency{
		Name:     name,
		Checker:  checker,
		Interval: interval,
	})
}

// Handler is simple handler for /health endpoint that returns 200 OK status.
func Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" && (r.Method == "GET" || r.Method == "HEAD") {
			w.WriteHeader(http.StatusOK)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

// GetStatuses returns statuses of service dependencies.
func GetStatuses() map[string]bool {
	results := make(map[string]bool)

	for _, dep := range dependencies {
		dep.RLock()
		consideredHealthy := dep.failuresAreNegligible()
		results[dep.Name] = consideredHealthy
		dep.RUnlock()
	}

	return results
}

func (dep *dependency) failuresAreNegligible() bool {
	return dep.FailureInARow < MaxFailureInARow
}

func (dep *dependency) applyHealthCheckResult(healthyNow bool) {
	if healthyNow {
		dep.FailureInARow = 0
		return
	}

	if dep.failuresAreNegligible() {
		dep.FailureInARow++ // Increment it so maybe it becomes non-negligible soon
	}
}

// Start starts async health check.
func Start(ctx context.Context) {
	dependencies = enabledDependencies

	for _, dep := range dependencies {
		dep.check()
	}

	logFailingDeps()

	for _, dep := range dependencies {
		go dep.runAsync(ctx)
	}

	go infinitelyLogFailingDeps()
}

func (dep *dependency) check() {
	err := dep.Checker.CheckHealth()
	dep.Lock()
	dep.LastErr = err
	dep.applyHealthCheckResult(err == nil)
	dep.Unlock()
}

func (dep *dependency) checkAndNotify(c chan struct{}) {
	dep.check()
	close(c)
}

func (dep *dependency) runAsync(ctx context.Context) {
	ticker := time.NewTicker(dep.Interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		c := make(chan struct{})
		go dep.checkAndNotify(c)

		select {
		case <-ctx.Done():
			return
		case <-c:
			continue
		case <-time.After(Timeout):
			continue
		}
	}
}

func logFailingDeps() {
	for _, dep := range dependencies {
		dep.RLock()
		consideredHealthy := dep.failuresAreNegligible()
		if !consideredHealthy {
			log.With().Error(fmt.Sprintf("healthcheck: dependency '%s' is failing: %v", dep.Name, dep.LastErr))
		}
		dep.RUnlock()
	}
}

func infinitelyLogFailingDeps() {
	ticker := time.NewTicker(AsyncLoggingInterval)
	for {
		<-ticker.C
		logFailingDeps()
	}
}
