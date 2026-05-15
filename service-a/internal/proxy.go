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
			// counter will drift after Remove() but we don't care if
			// we try an endpoint twice, since we only try N times
			service.Pool.Remove(svc, addr)
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
