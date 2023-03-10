package handlers

import (
	"context"
	"net/http"

	"github.com/bbt-t/lets-go-shortener/internal/adapter/storage"
	"github.com/bbt-t/lets-go-shortener/internal/controller"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// server HTTP-server struct.
type server struct {
	httpServer *http.Server
}

// newHTTPServer Initializing a new router.
func newHTTPServer(address string, s storage.Repository) *server {
	router := chi.NewRouter()

	router.Use(
		middleware.RealIP, // <- (!) Only if a reverse proxy is used (e.g. nginx) (!)
		middleware.Logger,
		middleware.Recoverer,
		CookieMiddleware,
		GzipHandle,
		GzipRequest,
	)

	router.Get("/ping", pingDB(s))
	router.Get("/{id}", recoverOriginalURL(s))
	router.Get("/api/user/urls", recoverAllURL(s))

	router.Delete("/api/user/urls", deleteURL(s))

	router.Post("/", recoverOriginalURLPost(s))
	router.Post("/api/shorten/batch", buildURLBatch(s))
	router.Post("/api/shorten", recoverOriginalURLPost(s))

	return &server{
		httpServer: &http.Server{
			Addr:    address,
			Handler: router,
		},
	}
}

// UP http-server start.
func (s *server) UP() error {
	return s.httpServer.ListenAndServe()
}

// Stop http-server stop.
func (s *server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// NewRouter - make router.
func NewRouter(address string, storage storage.Repository) controller.HTTPServer {
	return newHTTPServer(address, storage)
}
