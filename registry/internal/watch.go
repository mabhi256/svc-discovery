package internal

import (
	"encoding/json"
	"fmt"
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
	prefix := fmt.Sprintf("services/%s/", svc)
	var watchCh clientv3.WatchChan

	revisionStr := r.URL.Query().Get("revision")
	if revisionStr != "" {
		revision, err := strconv.ParseInt(revisionStr, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		watchCh = registry.Etcd.Watch(ctx, prefix, clientv3.WithPrefix(), clientv3.WithRev(revision))
	} else {
		watchCh = registry.Etcd.Watch(ctx, prefix, clientv3.WithPrefix())
	}

	for {
		select {
		case <-ctx.Done():
			// Client disconnected.
			return

		case resp := <-watchCh:
			for _, ev := range resp.Events {
				sendSSE(ev, w)
				flusher.Flush()
			}
		}
	}
}

func sendSSE(ev *clientv3.Event, w http.ResponseWriter) {
	var eventType EventType

	switch ev.Type {
	case mvccpb.PUT:
		eventType = PutEvent
	case mvccpb.DELETE:
		eventType = DeleteEvent
	}

	payload := WatchEvent{Type: eventType, Address: string(ev.Kv.Value)}

	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
}
