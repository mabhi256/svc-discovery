package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func (registry *Registry) ListHandler(w http.ResponseWriter, r *http.Request) {
	svc := chi.URLParam(r, "svc")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("/services/%s", svc)
	resp, err := registry.Etcd.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	endpoints := make([]Endpoint, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		// key is /svc/{name}/{id}
		id := strings.Split(string(kv.Key), "/")[2]
		endpoints = append(endpoints, Endpoint{LeaseID: id, Address: string(kv.Value)})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(endpoints)
}
