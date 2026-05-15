package internal

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func (registry *Registry) WatchHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Verify svc exists
	svc := chi.URLParam(r, "svc")
	if svc == "" {
		http.Error(w, "missing path param {svc}", http.StatusBadRequest)
		return
	}

	// 2. Assert Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// 3. Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 4. Start etcd watch (returns a channel)
	ctx := r.Context()
	prefix := fmt.Sprintf("/services/%s/", svc)
	var watchCh clientv3.WatchChan

	// DELETE event clears Kv.Value in etcd. WithPrevKV() in
	// watch options tells etcd to include the previous key-value
	revisionStr := r.URL.Query().Get("revision")
	if revisionStr != "" {
		revision, err := strconv.ParseInt(revisionStr, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		watchCh = registry.Etcd.Watch(ctx, prefix,
			clientv3.WithPrefix(), clientv3.WithRev(revision), clientv3.WithPrevKV())
	} else {
		watchCh = registry.Etcd.Watch(ctx, prefix,
			clientv3.WithPrefix(), clientv3.WithPrevKV())
	}
	log.Printf("Watch set up for prefix=%s\n", prefix)

	for {
		select {
		case <-ctx.Done():
			// Client disconnected.
			return

		case resp := <-watchCh:
			for _, ev := range resp.Events {
				log.Printf("Sending SSE event=%s key=%s rev=%d\n", ev.Type, string(ev.Kv.Key), ev.Kv.ModRevision)
				sendSSE(ev, w)
				flusher.Flush()
			}
		}
	}
}

func sendSSE(ev *clientv3.Event, w http.ResponseWriter) {
	var eventType EventType
	var value []byte

	switch ev.Type {
	case mvccpb.PUT:
		eventType = PutEvent
		value = ev.Kv.Value
	case mvccpb.DELETE:
		eventType = DeleteEvent
		value = ev.PrevKv.Value
	}

	fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", ev.Kv.ModRevision, eventType, string(value))
}
