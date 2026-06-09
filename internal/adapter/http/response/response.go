// Package response centralises JSON writing and the mapping from domain errors
// to HTTP status codes, keeping handlers thin and consistent.
package response

import (
	"encoding/json"
	"errors"
	"net/http"

	domain "github.com/bandrosh/boilerplate-api/internal/domain/user"
)

// ErrorBody is the canonical error envelope returned to clients.
type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// Error maps an error to an appropriate status code and writes the envelope.
// Domain errors are translated here so the domain stays transport-agnostic.
func Error(w http.ResponseWriter, err error) {
	status, code := classify(err)
	JSON(w, status, ErrorBody{Error: code, Message: err.Error()})
}

func classify(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict, "conflict"
	case errors.Is(err, domain.ErrNameRequired),
		errors.Is(err, domain.ErrEmailRequired),
		errors.Is(err, domain.ErrInvalidEmail):
		return http.StatusUnprocessableEntity, "validation_error"
	default:
		return http.StatusInternalServerError, "internal_error"
	}
}
