package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mabhi256/svc-discovery/service-b/internal"
)

func main() {
	registry := internal.RegistryClient{
		HttpCli: &http.Client{Timeout: 5 * time.Second},
	}

	registration, err := registry.Register()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	addr := internal.GetAdvertiseAddr()
	log.Printf("Registered svc=%s (address=%s) with lease=%s",
		internal.SVC, addr, registration.LeaseID)

	stopHeartbeat := registry.Heartbeat(registration.LeaseID, registration.Heartbeat)

	// 1. Setup ServeMux and Server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", internal.HealthCheckHandler)
	srv := &http.Server{Addr: addr, Handler: mux}

	// 2. Start server in a goroutine
	go func() {
		log.Printf("Server listening on %s", internal.GetAdvertiseAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 3. Listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	// 4. Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 5. Gracefully shut down
	stopHeartbeat()
	registry.Deregister(registration.LeaseID)

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server shutdown done")
}
