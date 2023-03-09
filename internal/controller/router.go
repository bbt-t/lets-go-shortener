package controller

import (
	"context"
	"net/http"

	"github.com/bbt-t/lets-go-shortener/internal/adapter/storage"
	"github.com/bbt-t/lets-go-shortener/internal/controller/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// HTTP-server struct.

type server struct {
	httpServer *http.Server
}

// Initializing a new router.

func newHTTPServer(address string, s storage.Repository) *server {
	router := chi.NewRouter()

	router.Use(
		middleware.RealIP, // <- (!) Only if a reverse proxy is used (e.g. nginx) (!)
		middleware.Logger,
		middleware.Recoverer,
		handlers.CookieMiddleware,
		handlers.GzipHandle,
		handlers.GzipRequest,
	)

	router.Get("/ping", handlers.Ping(s))
	router.Get("/{id}", handlers.RecoverOriginalURL(s))
	router.Get("/api/user/urls", handlers.RecoverAllURL(s))

	router.Delete("/api/user/urls", handlers.DeleteURL(s))

	router.Post("/", handlers.RecoverOriginalURLPost(s))
	router.Post("/api/shorten/batch", handlers.URLBatch(s))
	router.Post("/api/shorten", handlers.RecoverOriginalURLPost(s))

	return &server{
		httpServer: &http.Server{
			Addr:    address,
			Handler: router,
		},
	}
}

func (s *server) UP() error {
	/*
		http-server start.
	*/
	return s.httpServer.ListenAndServe()
}

func (s *server) Stop(ctx context.Context) error {
	/*
		http-server stop.
	*/
	return s.httpServer.Shutdown(ctx)
}

// Router interface.

type HTTPServer interface {
	UP() error
	Stop(ctx context.Context) error
}

// Make router.

func NewRouter(address string, storage storage.Repository) HTTPServer {
	return newHTTPServer(address, storage)
}
