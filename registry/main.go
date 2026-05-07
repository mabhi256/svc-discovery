package main

// The Registry interfaces with etcd and exposes HTTP API to services:
//
//   POST   /services                       register a new instance
//   PUT    /services/{svc}/{id}/heartbeat  renew TTL
//   DELETE /services/{svc}/{id}            deregister
//   GET    /services/{svc}                 list healthy endpoints
//   GET    /watch/{svc}                    SSE stream of changes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func getEndpoints() []string {
	pointsStr := os.Getenv("ETCD_ENDPOINTS")
	if pointsStr == "" {
		return []string{"localhost:2379"}
	}

	return strings.Split(pointsStr, ",")
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8500"
	}

	return port
}

func main() {
	endpoints := getEndpoints()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("cannot connect to etcd: %v", err)
	}
	defer cli.Close()
	log.Printf("Connected to etcd cluster: %v", endpoints)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// https://github.com/go-chi/chi/blob/master/_examples/graceful/main.go
	// The HTTP Server
	addr := fmt.Sprintf("0.0.0.0:%s", getPort())
	server := &http.Server{Addr: addr, Handler: r}

	// Create context that listens for the interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run server in the background
	go func() {
		log.Printf("Running registry service on port: %s", getPort())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Listen for the interrupt signal
	<-ctx.Done()

	// Create shutdown context with 30-second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Trigger graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
	log.Println("Shutting down gracefully")
}
