// Package that creates service and runs it.

package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bbt-t/lets-go-shortener/internal/adapter/storage"
	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/controller/handlers"
)

// Run service.
func Run(cfg config.Config) {
	s, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := handlers.NewRouter(cfg.ServerAddress, s)
	// Start server:
	go func() {
		log.Println(server.UP())
	}()
	// Graceful shutdown:
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-gracefulStop

	ctxGrace, cancelGrace := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelGrace()
	if err := server.Stop(ctxGrace); err != nil {
		log.Printf("! Error shutting down server: !\n%v", err)
	} else {
		log.Println("! SERVER STOPPED !")
	}
}
