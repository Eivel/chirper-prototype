package routing

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// NewBaseRouter creates a new Chi router with basic middlewares.
func NewBaseRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON), // TODO: extend in the future to handle different formats.
		middleware.Logger, // TODO: Integrate with external logging system, eg. Datadog.
		middleware.DefaultCompress,
		middleware.RedirectSlashes,
		middleware.Recoverer,
		Authorize,
	)

	return router
}
