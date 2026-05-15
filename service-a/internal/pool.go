package internal

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
)

type endpoint struct {
	addr   string
	active bool
}

type pool struct {
	endpoints []*endpoint
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
	for _, a := range addr {
		pm.pools[svc].endpoints = append(pm.pools[svc].endpoints, &endpoint{addr: a, active: true})
	}
}

func (pm *PoolMap) Remove(svc string, addr string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p := pm.pools[svc]
	if p == nil || len(p.endpoints) == 0 {
		return errors.New("no endpoints")
	}

	for i, e := range p.endpoints {
		if e.addr == addr {
			p.endpoints = append(p.endpoints[:i], p.endpoints[i+1:]...)
			return nil
		}
	}
	return nil
}

func (pm *PoolMap) SetActive(svc string, addr string, active bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p := pm.pools[svc]
	if p == nil {
		return
	}
	for _, e := range p.endpoints {
		if e.addr == addr {
			e.active = active
			return
		}
	}
}

func (pm *PoolMap) Len(svc string) int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p := pm.pools[svc]
	if p == nil {
		return 0
	}

	return len(p.endpoints)
}

// Used for round robin — skips inactive endpoints
func (pm *PoolMap) Next(svc string) (string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p := pm.pools[svc]
	if p == nil || len(p.endpoints) == 0 {
		return "", errors.New("no endpoints")
	}

	n := uint64(len(p.endpoints))
	start := p.counter.Add(1) - 1
	for i := range n {
		e := p.endpoints[(start+i)%n]
		if e.active {
			return e.addr, nil
		}
	}
	return "", errors.New("no active endpoints")
}

func (service *Service) Snapshot() map[string][]string {
	service.Pool.mu.Lock()
	snapshot := make(map[string][]string, len(service.Pool.pools))
	for svc, p := range service.Pool.pools {
		addrs := make([]string, len(p.endpoints))
		for i, e := range p.endpoints {
			addrs[i] = e.addr
		}
		snapshot[svc] = addrs
	}
	service.Pool.mu.Unlock()

	return snapshot
}

func (service *Service) PoolHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service.Snapshot())
}
