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

type Router struct {
	Log    *slog.Logger
	User   *handler.UserHandler
	Health *handler.HealthHandler
}

func (rt Router) Build() http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(mw.RequestLogger(rt.Log))

	r.Get("/healthz", rt.Health.Live)
	r.Get("/readyz", rt.Health.Ready)

	r.Group(func(r chi.Router) {
		r.Use(otelMiddleware)
		r.Mount("/api/v1/users", rt.User.Routes())
	})

	return r
}

func otelMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http.server")
}
