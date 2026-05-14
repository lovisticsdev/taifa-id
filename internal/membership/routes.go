package membership

import "github.com/go-chi/chi/v5"

func RegisterRoutes(router chi.Router, handler *Handler) {
	router.Route("/api/v1/memberships", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/{membership_id}", handler.GetByID)
		r.Patch("/{membership_id}/status", handler.UpdateStatus)

		r.Post("/{membership_id}/roles", handler.AddRole)
		r.Get("/{membership_id}/roles", handler.ListRoles)
		r.Delete("/{membership_id}/roles/{role}", handler.RemoveRole)
	})

	router.Get("/api/v1/persons/{person_id}/memberships", handler.ListByPerson)
	router.Get("/api/v1/organizations/{organization_id}/memberships", handler.ListByOrganization)
}
