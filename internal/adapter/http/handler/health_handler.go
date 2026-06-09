package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/bandrosh/boilerplate-api/internal/adapter/http/response"
)

// Pinger is anything that can verify a downstream dependency is reachable
// (e.g. the pgx pool). Used by the readiness probe.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler serves liveness and readiness probes for orchestrators.
type HealthHandler struct {
	db Pinger
}

// NewHealthHandler builds the handler with the dependencies to check.
func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Live reports that the process is up. It must not touch dependencies.
func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready reports whether the app can serve traffic (dependencies reachable).
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		response.JSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable", "database": "down",
		})
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "ok", "database": "up",
	})
}
