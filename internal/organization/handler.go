package organization

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
	var req CreateOrganizationRequest
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
		writeOrganizationError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusCreated, ToOrganizationResponse(created))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.service.List(r.Context())
	if err != nil {
		writeOrganizationError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToOrganizationResponses(orgs))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	organizationID := chi.URLParam(r, "organization_id")

	org, err := h.service.GetByID(r.Context(), organizationID)
	if err != nil {
		writeOrganizationError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToOrganizationResponse(org))
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	organizationID := chi.URLParam(r, "organization_id")

	var req UpdateOrganizationStatusRequest
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
		organizationID,
		req,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writeOrganizationError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToOrganizationResponse(updated))
}

func (h *Handler) AddCapability(w http.ResponseWriter, r *http.Request) {
	organizationID := chi.URLParam(r, "organization_id")

	var req AddOrganizationCapabilityRequest
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

	created, err := h.service.AddCapability(
		r.Context(),
		organizationID,
		req,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writeOrganizationError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusCreated, ToCapabilityResponse(created))
}

func (h *Handler) ListCapabilities(w http.ResponseWriter, r *http.Request) {
	organizationID := chi.URLParam(r, "organization_id")

	capabilities, err := h.service.ListCapabilities(r.Context(), organizationID)
	if err != nil {
		writeOrganizationError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToCapabilityResponses(capabilities))
}

func (h *Handler) RemoveCapability(w http.ResponseWriter, r *http.Request) {
	organizationID := chi.URLParam(r, "organization_id")
	capability := chi.URLParam(r, "capability")

	removed, err := h.service.RemoveCapability(
		r.Context(),
		organizationID,
		capability,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writeOrganizationError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToCapabilityResponse(removed))
}

func writeOrganizationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		httpserver.WriteError(
			w,
			r,
			http.StatusBadRequest,
			httpserver.ErrorCodeValidation,
			"Organization request is invalid.",
		)

	case errors.Is(err, ErrNotFound):
		httpserver.WriteError(
			w,
			r,
			http.StatusNotFound,
			httpserver.ErrorCodeNotFound,
			"Organization was not found.",
		)

	case errors.Is(err, ErrCapabilityNotFound):
		httpserver.WriteError(
			w,
			r,
			http.StatusNotFound,
			httpserver.ErrorCodeNotFound,
			"Organization capability was not found.",
		)

	case errors.Is(err, ErrDuplicateCapability):
		httpserver.WriteError(
			w,
			r,
			http.StatusConflict,
			httpserver.ErrorCodeConflict,
			"Organization capability already exists.",
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
