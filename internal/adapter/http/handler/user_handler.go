package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/bandrosh/boilerplate-api/internal/adapter/http/dto"
	"github.com/bandrosh/boilerplate-api/internal/adapter/http/response"
	app "github.com/bandrosh/boilerplate-api/internal/application/user"
)

type UserHandler struct {
	svc *app.Service
}

func NewUserHandler(svc *app.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.ErrorBody{
			Error: "bad_request", Message: "invalid JSON body",
		})
		return
	}

	u, err := h.svc.Create(r.Context(), app.CreateInput{Name: req.Name, Email: req.Email})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, dto.FromDomain(u))
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	u, err := h.svc.Get(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, dto.FromDomain(u))
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	cursor := r.URL.Query().Get("cursor")

	page, err := h.svc.List(r.Context(), limit, cursor)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, dto.FromDomainPage(page))
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.ErrorBody{
			Error: "bad_request", Message: "invalid JSON body",
		})
		return
	}

	u, err := h.svc.Update(r.Context(), app.UpdateInput{ID: id, Name: req.Name})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, dto.FromDomain(u))
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.ErrorBody{
			Error: "bad_request", Message: "invalid id",
		})
		return uuid.Nil, false
	}
	return id, true
}

func parseInt(raw string, def int32) int32 {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return int32(v)
}
