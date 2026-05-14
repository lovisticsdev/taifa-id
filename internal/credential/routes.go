package credential

import "github.com/go-chi/chi/v5"

func RegisterRoutes(router chi.Router, handler *Handler) {
	router.Route("/api/v1/credentials", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/{credential_id}", handler.GetByID)
	})

	router.Get("/api/v1/persons/{person_id}/credentials", handler.ListByPerson)
}
