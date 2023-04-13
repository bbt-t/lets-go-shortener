// Package that creates service and runs it.

package app

import (
	"context"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bbt-t/lets-go-shortener/internal/adapter/storage"
	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/controller"
	"github.com/bbt-t/lets-go-shortener/internal/controller/handlers"
	"github.com/bbt-t/lets-go-shortener/internal/usecase"

	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
)

// Run service.
func Run(cfg config.Config) {
	// Storage
	s, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatal(err)
	}
	// TLS
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		log.Fatalln(err)
	}
	manager := &autocert.Manager{
		Cache:      autocert.DirCache("cache-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(baseURL.Host),
	}
	// New service
	service := usecase.NewShortenerService(cfg, s)

	// New router
	h := handlers.NewShortenerHandler(cfg, service)

	// New server
	server := controller.NewRouter(cfg, h, manager)

	// Init TCP listener for gRPC
	listen, err := net.Listen("tcp", cfg.GrpcPort)
	if err != nil {
		log.Fatal("Error listening -> ", err)
	}
	defer listen.Close()
	// Create gRPC-server without service
	grpcServ := grpc.NewServer()
	// Init gRPC service
	//pb.RegisterShortenerServer(grpcServ, handlers.NewShortenerServer(cfg, service))

	go func() {
		log.Println("-> Start gRPC service <-")
		if err := grpcServ.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}()

	// Start server
	go func() {
		if cfg.EnableHTTPS {
			log.Println(server.StartTLS("", ""))
		} else {
			log.Println(server.Start())
		}
	}()
	// Graceful shutdown
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
