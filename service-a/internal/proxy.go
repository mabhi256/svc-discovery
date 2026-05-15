package internal

import (
	"fmt"
	"io"
	"net/http"
)

func (service *Service) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	svc := r.PathValue("svc")
	for range service.Pool.Len(svc) {
		addr, err := service.Pool.Next(svc)
		if err != nil {
			break
		}

		url := fmt.Sprintf("http://%s/hello", addr)
		resp, err := service.Cli.Get(url)
		if err != nil {
			// Len value will drift after SetActive(false) but we don't care if
			// we try an endpoint twice, since we only try N times
			service.Pool.SetActive(svc, addr, false)
			continue
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		// tracing header that tells the caller which backend instance actually handled the request
		w.Header().Set("X-Upstream", addr)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
	http.Error(w, "no upstream available", http.StatusServiceUnavailable)
}
