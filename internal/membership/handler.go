package membership

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
	var req CreateMembershipRequest
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
		writeMembershipError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusCreated, ToMembershipResponse(created))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	membershipID := chi.URLParam(r, "membership_id")

	membership, err := h.service.GetByID(r.Context(), membershipID)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToMembershipResponse(membership))
}

func (h *Handler) ListByPerson(w http.ResponseWriter, r *http.Request) {
	personID := chi.URLParam(r, "person_id")

	memberships, err := h.service.ListByPerson(r.Context(), personID)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToMembershipResponses(memberships))
}

func (h *Handler) ListByOrganization(w http.ResponseWriter, r *http.Request) {
	organizationID := chi.URLParam(r, "organization_id")

	memberships, err := h.service.ListByOrganization(r.Context(), organizationID)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToMembershipResponses(memberships))
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	membershipID := chi.URLParam(r, "membership_id")

	var req UpdateMembershipStatusRequest
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
		membershipID,
		req,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToMembershipResponse(updated))
}

func (h *Handler) AddRole(w http.ResponseWriter, r *http.Request) {
	membershipID := chi.URLParam(r, "membership_id")

	var req AddMembershipRoleRequest
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

	created, err := h.service.AddRole(
		r.Context(),
		membershipID,
		req,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusCreated, ToMembershipRoleResponse(created))
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	membershipID := chi.URLParam(r, "membership_id")

	roles, err := h.service.ListRoles(r.Context(), membershipID)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToMembershipRoleResponses(roles))
}

func (h *Handler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	membershipID := chi.URLParam(r, "membership_id")
	role := chi.URLParam(r, "role")

	removed, err := h.service.RemoveRole(
		r.Context(),
		membershipID,
		role,
		httpserver.CorrelationIDFromContext(r.Context()),
	)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}

	httpserver.WriteData(w, r, http.StatusOK, ToMembershipRoleResponse(removed))
}

func writeMembershipError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		httpserver.WriteError(
			w,
			r,
			http.StatusBadRequest,
			httpserver.ErrorCodeValidation,
			"Membership request is invalid.",
		)

	case errors.Is(err, ErrNotFound):
		httpserver.WriteError(
			w,
			r,
			http.StatusNotFound,
			httpserver.ErrorCodeNotFound,
			"Membership was not found.",
		)

	case errors.Is(err, ErrRoleNotFound):
		httpserver.WriteError(
			w,
			r,
			http.StatusNotFound,
			httpserver.ErrorCodeNotFound,
			"Membership role was not found.",
		)

	case errors.Is(err, ErrReferenceNotFound):
		httpserver.WriteError(
			w,
			r,
			http.StatusNotFound,
			httpserver.ErrorCodeNotFound,
			"Referenced person or organization was not found.",
		)

	case errors.Is(err, ErrPersonNotActive):
		httpserver.WriteError(
			w,
			r,
			http.StatusConflict,
			httpserver.ErrorCodeConflict,
			"Person is not active.",
		)

	case errors.Is(err, ErrOrganizationNotActive):
		httpserver.WriteError(
			w,
			r,
			http.StatusConflict,
			httpserver.ErrorCodeConflict,
			"Organization is not active.",
		)

	case errors.Is(err, ErrMembershipNotActive):
		httpserver.WriteError(
			w,
			r,
			http.StatusConflict,
			httpserver.ErrorCodeConflict,
			"Membership is not active.",
		)

	case errors.Is(err, ErrDuplicateActiveMembership):
		httpserver.WriteError(
			w,
			r,
			http.StatusConflict,
			httpserver.ErrorCodeConflict,
			"An active or pending membership already exists for this person, organization, and membership type.",
		)

	case errors.Is(err, ErrDuplicateRole):
		httpserver.WriteError(
			w,
			r,
			http.StatusConflict,
			httpserver.ErrorCodeConflict,
			"Membership role already exists.",
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
