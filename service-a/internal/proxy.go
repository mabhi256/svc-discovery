package internal

import (
	"fmt"
	"io"
	"net/http"
)

func (service *Service) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	svc := r.PathValue("svc")
	addr, err := service.Pool.Next(svc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	url := fmt.Sprintf("http://%s/health", addr)
	resp, err := service.Cli.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	// tracing header that tells the caller which backend instance actually handled the request
	w.Header().Set("X-Upstream", addr)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
