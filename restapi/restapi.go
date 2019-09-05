package restapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/cors"
	mw "gitlab.com/p-invent/mosoly-ledger-bridge/web/middleware"
)

// ServiceConfig is service configuration
type ServiceConfig struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// An origin may contain a wildcard (*) to replace 0 or more characters
	// (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penalty.
	// Only one wildcard can be used per origin.
	// Default value is ["*"]
	AllowedOrigins []string
	// Port is service HTTP port
	Port int
}

// NewService creates an instance of Service
func NewService(cfg *ServiceConfig) *Service {

	// middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	promHandler := alice.Constructor(func(h http.Handler) http.Handler {
		return mw.PrometheusHandler("/metrics", h)
	})
	expVarsHandler := alice.Constructor(func(h http.Handler) http.Handler {
		return mw.ExpVarHandler("/debug/vars", h)
	})

	handler := alice.New(
		mw.RecoverHandler,
		mw.HTTPStatsHandler,
		mw.NoCache,
		corsHandler.Handler,
		mw.HealthHandler,
		promHandler,
		expVarsHandler,
	).Then(http.NotFoundHandler())

	return &Service{
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: handler,
		},
	}
}

// Service is simple REST API service
type Service struct {
	srv *http.Server
}

// Serve the api
func (s *Service) Serve(ctx context.Context) (err error) {
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		if err = s.srv.ListenAndServe(); err == http.ErrServerClosed {
			err = nil // ignore server closed error
		}
	}()

	// waiting for
	select {
	case <-ctx.Done(): // cancellation
		// shutdown HTTP service
		serr := shutdown(s.srv, 10*time.Second)
		// waiting it fully stopped
		<-stopped
		// if no errors assign shutdown error
		if err == nil {
			err = serr
		}
	case <-stopped: // service stop
	}

	return
}

func shutdown(srv *http.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return srv.Shutdown(ctx)
}
