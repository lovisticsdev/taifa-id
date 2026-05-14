package actorcontext

import "github.com/go-chi/chi/v5"

func RegisterRoutes(router chi.Router, handler *Handler) {
	router.Route("/api/v1/actor-context", func(r chi.Router) {
		r.Post("/resolve", handler.Resolve)
	})
}
