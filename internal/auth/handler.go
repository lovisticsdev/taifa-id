package auth

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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
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

	result, err := h.service.Login(
		r.Context(),
		req,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writeAuthError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToLoginResponse(result))
}

func (h *Handler) Introspect(w http.ResponseWriter, r *http.Request) {
	var req IntrospectRequest
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

	result, err := h.service.Introspect(
		r.Context(),
		req,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writeAuthError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToIntrospectResponse(result))
}

func writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		httpserver.WriteError(
			w,
			r,
			http.StatusBadRequest,
			httpserver.ErrorCodeValidation,
			"Authentication request is invalid.",
		)

	case errors.Is(err, ErrInvalidCredential):
		httpserver.WriteError(
			w,
			r,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Invalid username or password.",
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
