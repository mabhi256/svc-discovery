package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func (registry *Registry) HeartbeatHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Parse the hex lease ID back to clientv3.LeaseID
	var leaseID clientv3.LeaseID
	fmt.Sscanf(id, "%x", &leaseID)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := registry.Etcd.KeepAliveOnce(ctx, leaseID)
	if err != nil {
		http.Error(w, "Lease expired or not found. Re-register the service", http.StatusGone)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
