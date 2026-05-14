package person

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"taifa-id/internal/platform/httpserver"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreatePersonRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(
			w,
			r,
			http.StatusBadRequest,
			httpserver.ErrorCodeInvalidJSON,
			"Request body must be valid JSON.",
		)
		return
	}

	created, err := h.service.Create(
		r.Context(),
		req,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writePersonError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusCreated, ToResponse(created))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	personID := chi.URLParam(r, "person_id")

	p, err := h.service.GetByID(r.Context(), personID)
	if err != nil {
		writePersonError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToResponse(p))
}

func (h *Handler) GetBySyntheticNIN(w http.ResponseWriter, r *http.Request) {
	syntheticNIN := r.URL.Query().Get("synthetic_nin")

	p, err := h.service.GetBySyntheticNIN(r.Context(), syntheticNIN)
	if err != nil {
		writePersonError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToResponse(p))
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	personID := chi.URLParam(r, "person_id")

	var req UpdatePersonStatusRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(
			w,
			r,
			http.StatusBadRequest,
			httpserver.ErrorCodeInvalidJSON,
			"Request body must be valid JSON.",
		)
		return
	}

	updated, err := h.service.UpdateStatus(
		r.Context(),
		personID,
		req,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writePersonError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToResponse(updated))
}

func writePersonError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		httpserver.WriteError(
			w,
			r,
			http.StatusBadRequest,
			httpserver.ErrorCodeValidation,
			"Person request is invalid.",
		)

	case errors.Is(err, ErrNotFound):
		httpserver.WriteError(
			w,
			r,
			http.StatusNotFound,
			httpserver.ErrorCodeNotFound,
			"Person was not found.",
		)

	case errors.Is(err, ErrDuplicateSyntheticNIN):
		httpserver.WriteError(
			w,
			r,
			http.StatusConflict,
			httpserver.ErrorCodeConflict,
			"Synthetic NIN already exists.",
		)

	default:
		httpserver.WriteError(
			w,
			r,
			http.StatusInternalServerError,
			httpserver.ErrorCodeInternal,
			"An internal error occurred.",
		)
	}
}
