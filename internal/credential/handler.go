package credential

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
	var req CreateCredentialRequest
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
		writeCredentialError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusCreated, ToCredentialResponse(created))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	credentialID := chi.URLParam(r, "credential_id")

	credential, err := h.service.GetByID(r.Context(), credentialID)
	if err != nil {
		writeCredentialError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToCredentialResponse(credential))
}

func (h *Handler) ListByPerson(w http.ResponseWriter, r *http.Request) {
	personID := chi.URLParam(r, "person_id")

	credentials, err := h.service.ListByPerson(r.Context(), personID)
	if err != nil {
		writeCredentialError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToCredentialResponses(credentials))
}

func writeCredentialError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		httpserver.WriteError(
			w,
			r,
			http.StatusBadRequest,
			httpserver.ErrorCodeValidation,
			"Credential request is invalid.",
		)

	case errors.Is(err, ErrNotFound):
		httpserver.WriteError(
			w,
			r,
			http.StatusNotFound,
			httpserver.ErrorCodeNotFound,
			"Credential was not found.",
		)

	case errors.Is(err, ErrReferenceNotFound):
		httpserver.WriteError(
			w,
			r,
			http.StatusNotFound,
			httpserver.ErrorCodeNotFound,
			"Referenced person was not found.",
		)

	case errors.Is(err, ErrPersonNotActive):
		httpserver.WriteError(
			w,
			r,
			http.StatusConflict,
			httpserver.ErrorCodeConflict,
			"Person is not active.",
		)

	case errors.Is(err, ErrDuplicateUsername):
		httpserver.WriteError(
			w,
			r,
			http.StatusConflict,
			httpserver.ErrorCodeConflict,
			"Username already exists.",
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
