package response

import (
	"encoding/json"
	"errors"
	"net/http"

	domain "github.com/bandrosh/boilerplate-api/internal/domain/user"
)

type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

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
