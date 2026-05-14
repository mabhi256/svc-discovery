package internal

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
)

type pool struct {
	endpoints []string
	counter   atomic.Uint64
}

type PoolMap struct {
	mu       sync.Mutex
	pools    map[string]*pool
	revision int64
}

func (pm *PoolMap) SetRevision(rev int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.revision = rev
}

func (pm *PoolMap) Revision() int64 {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.revision
}

func NewPool() *PoolMap {
	return &PoolMap{pools: map[string]*pool{}}
}

func (pm *PoolMap) Add(svc string, addr ...string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.pools[svc] == nil {
		pm.pools[svc] = &pool{}
	}
	pm.pools[svc].endpoints = append(pm.pools[svc].endpoints, addr...)
}

func (pm *PoolMap) Remove(svc string, addr string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p := pm.pools[svc]
	if p == nil || len(p.endpoints) == 0 {
		return errors.New("no endpoints")
	}

	for i, e := range p.endpoints {
		if e == addr {
			p.endpoints = append(p.endpoints[:i], p.endpoints[i+1:]...)
			return nil
		}
	}
	return nil
}

func (pm *PoolMap) Get(svc string) []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p := pm.pools[svc]
	if p == nil || len(p.endpoints) == 0 {
		return []string{}
	}

	return slices.Clone(p.endpoints)
}

// Used for round robin
func (pm *PoolMap) Next(svc string) (string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p := pm.pools[svc]
	if p == nil || len(p.endpoints) == 0 {
		return "", errors.New("no endpoints")
	}
	idx := (p.counter.Add(1) - 1) % uint64(len(p.endpoints))
	return p.endpoints[idx], nil
}

func (service *Service) PoolHandler(w http.ResponseWriter, r *http.Request) {
	addrs := service.Pool.Get(SVC_B)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addrs)
}
