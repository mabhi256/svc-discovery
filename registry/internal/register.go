package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func (registry *Registry) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var body RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := registry.Validate.Struct(body); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Use etcd to grant a lease. If the service stops heartbeating,
	// etcd will automatically delete the key after the lease expires
	lease, err := registry.Etcd.Grant(ctx, LeaseTTL)
	if err != nil {
		// 503: Dependency is down, try again later
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Use lease ID in hex to identify service instances
	// Convention: Hex IDs also used in - etcdctl lease list
	id := fmt.Sprintf("%x", lease.ID)
	key := fmt.Sprintf("/services/%s/%s", body.Service, id)

	etcdResp, err := registry.Etcd.Put(ctx, key, body.Address, clientv3.WithLease(lease.ID))
	if err != nil {
		// 500: A lease was just granted but Put failed unexpectedly
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	revision := etcdResp.Header.Revision
	log.Printf("Registered key=%s addr=%s lease=%x (rev=%v)\n",
		key, body.Address, lease.ID, revision)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RegisterResponse{
		LeaseID:   id,
		Revision:  revision,
		Heartbeat: Heartbeat,
	})
}
