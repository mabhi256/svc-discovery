package internal

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	resp := map[string]string{
		"status":   "ok",
		"hostname": hostname,
		"time":     time.Now().UTC().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}
