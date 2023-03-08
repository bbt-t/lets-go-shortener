// Package that creates service and runs it.

package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bbt-t/lets-go-shortener/internal/adapter/storage"
	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/controller/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Run service.
func Run(cfg config.Config) {
	route := chi.NewRouter()
	s, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := http.Server{Addr: cfg.ServerAddress, Handler: route}

	route.Use(
		middleware.RealIP, // <- (!) Only if a reverse proxy is used (e.g. nginx) (!)
		middleware.Logger,
		middleware.Recoverer,
		handlers.CookieMiddleware,
		handlers.GzipHandle,
		handlers.GzipRequest,
	)

	route.Get("/ping", handlers.Ping(s))
	route.Get("/{id}", handlers.RecoverOriginalURL(s))
	route.Get("/api/user/urls", handlers.RecoverAllURL(s))
	route.Delete("/api/user/urls", handlers.DeleteURL(s))
	route.Post("/", handlers.RecoverOriginalURLPost(s))
	route.Post("/api/shorten/batch", handlers.URLBatch(s))
	route.Post("/api/shorten", handlers.RecoverOriginalURLPost(s))

	go func() {
		log.Println(server.ListenAndServe())
	}()
	// Graceful shutdown:
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-gracefulStop

	ctxGrace, cancelGrace := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelGrace()
	if err := server.Shutdown(ctxGrace); err != nil {
		log.Printf("! Error shutting down server: !\n%v", err)
	} else {
		log.Println("! SERVER STOPPED !")
	}
}
