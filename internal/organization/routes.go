package organization

import "github.com/go-chi/chi/v5"

func RegisterRoutes(router chi.Router, handler *Handler) {
	router.Route("/api/v1/organizations", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/", handler.List)
		r.Get("/{organization_id}", handler.GetByID)
		r.Patch("/{organization_id}/status", handler.UpdateStatus)

		r.Post("/{organization_id}/capabilities", handler.AddCapability)
		r.Get("/{organization_id}/capabilities", handler.ListCapabilities)
		r.Delete("/{organization_id}/capabilities/{capability}", handler.RemoveCapability)
	})
}
