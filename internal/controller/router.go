// Package that creates Router and interface.

package controller

import (
	"context"
	"net/http"

	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/controller/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/acme/autocert"
)

// server HTTP-server struct.
type server struct {
	httpServer *http.Server
}

// newHTTPServer Initializing a new router.
func newHTTPServer(cfg config.Config, s *handlers.ShortenerHandler, manager *autocert.Manager) *server {
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
	router.Post("/api/shorten", handlers.RecoverOriginalURLPost(s))
	router.Post("/api/shorten/batch", handlers.URLBatch(s))

	router.Group(func(r chi.Router) {
		r.Use(handlers.NewIPPermissionsChecker(cfg))
		r.Get("/api/internal/stats", handlers.StatisticHandler(s))
	})

	return &server{
		httpServer: &http.Server{
			Addr:      cfg.ServerAddress,
			Handler:   router,
			TLSConfig: manager.TLSConfig(),
		},
	}
}

// Start http-server start.
func (s *server) Start() error {
	return s.httpServer.ListenAndServe()
}

// StartTLS http-server start with TLS.
func (s *server) StartTLS(certFile, keyFile string) error {
	return s.httpServer.ListenAndServeTLS(certFile, keyFile)
}

// Stop http-server stop.
func (s *server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// HTTPServer - Router interface.
type HTTPServer interface {
	Start() error
	StartTLS(certFile, keyFile string) error
	Stop(ctx context.Context) error
}

// NewRouter - make router.
func NewRouter(cfg config.Config, s *handlers.ShortenerHandler, manager *autocert.Manager) HTTPServer {
	return newHTTPServer(cfg, s, manager)
}
