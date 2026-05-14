package person

import "github.com/go-chi/chi/v5"

func RegisterRoutes(router chi.Router, handler *Handler) {
	router.Route("/api/v1/persons", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/", handler.GetBySyntheticNIN)
		r.Get("/{person_id}", handler.GetByID)
		r.Patch("/{person_id}/status", handler.UpdateStatus)
	})
}
