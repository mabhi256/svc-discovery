package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mabhi256/svc-discovery/service-a/internal"
)

// Service A is the consumer:
//   1. Fetch all service B endpoints via GET /services/{svc}.
//   2. Opens a persistent SSE watch via GET /watch/{svc}.
//   3. Routes outbound RPCs through a local round-robin pool.
//   4. Exposes GET /call-b so you can trigger a real proxied call.

func main() {
	cli := &http.Client{Timeout: 5 * time.Second}
	watchCli := &http.Client{}
	pool := internal.NewPool()
	service := &internal.Service{Cli: cli, Pool: pool}
	if err := service.Init(); err != nil {
		log.Println("init:", err)
	}

	for _, svc := range internal.GetUpstreamServices() {
		go internal.Watch(watchCli, svc, pool)
	}

	// 1. Setup ServeMux and Server
	mux := http.NewServeMux()
	mux.HandleFunc("GET /pool", service.PoolHandler)
	mux.HandleFunc("GET /proxy/{svc}", service.ProxyHandler)
	mux.HandleFunc("GET /health", service.HealthCheckHandler)

	addr := internal.GetAdvertiseAddr()
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
	log.Println("...shutting down server")

	// 4. Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 5. Gracefully shut down
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server shutdown done")
}
