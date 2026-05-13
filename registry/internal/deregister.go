package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (registry *Registry) DeregisterHandler(w http.ResponseWriter, r *http.Request) {
	svc := chi.URLParam(r, "svc")
	id := chi.URLParam(r, "id")
	key := fmt.Sprintf("/services/%s/%s", svc, id)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := registry.Etcd.Delete(ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	log.Printf("De-registered key=%s (rev=%v)\n", key, resp.Header.Revision)

	w.WriteHeader(http.StatusNoContent)
}
