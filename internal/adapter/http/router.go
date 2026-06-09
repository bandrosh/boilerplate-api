// Package http is the inbound HTTP adapter. It builds the chi router, wires
// middlewares and mounts the feature handlers.
package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/bandrosh/boilerplate-api/internal/adapter/http/handler"
	mw "github.com/bandrosh/boilerplate-api/internal/adapter/http/middleware"
)

// Router bundles the handlers the router needs to mount.
type Router struct {
	Log    *slog.Logger
	User   *handler.UserHandler
	Health *handler.HealthHandler
}

// Build assembles the chi router with the standard middleware stack.
func (rt Router) Build() http.Handler {
	r := chi.NewRouter()

	// Standard production middleware stack.
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(mw.RequestLogger(rt.Log))

	// Health probes (kept out of the traced/business routes).
	r.Get("/healthz", rt.Health.Live)
	r.Get("/readyz", rt.Health.Ready)

	// Business API, instrumented with OpenTelemetry HTTP tracing.
	r.Group(func(r chi.Router) {
		r.Use(otelMiddleware)
		r.Mount("/api/v1/users", rt.User.Routes())
	})

	return r
}

// otelMiddleware wraps handlers with OpenTelemetry HTTP instrumentation.
func otelMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http.server")
}
