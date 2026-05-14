package actorcontext

import (
	"errors"
	"net/http"

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

func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	var req ResolveActorContextRequest
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

	resolved, err := h.service.Resolve(
		r.Context(),
		req,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writeActorContextError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToResponse(resolved))
}

func writeActorContextError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		httpserver.WriteError(
			w,
			r,
			http.StatusBadRequest,
			httpserver.ErrorCodeValidation,
			"Actor context request is invalid.",
		)

	case errors.Is(err, ErrInvalidToken):
		httpserver.WriteError(
			w,
			r,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Token is invalid.",
		)

	case errors.Is(err, ErrOrganizationNotFound):
		httpserver.WriteError(
			w,
			r,
			http.StatusNotFound,
			httpserver.ErrorCodeNotFound,
			"Organization was not found.",
		)

	case errors.Is(err, ErrCredentialInactive),
		errors.Is(err, ErrPersonInactive),
		errors.Is(err, ErrOrganizationInactive),
		errors.Is(err, ErrNoActiveMembership):
		httpserver.WriteError(
			w,
			r,
			http.StatusForbidden,
			"FORBIDDEN",
			"Actor context was denied.",
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
