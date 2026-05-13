package internal

import (
	"net/http"
)

func (registry *Registry) WatchHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("watch"))
}
