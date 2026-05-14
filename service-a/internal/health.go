package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func (service *Service) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	resp := map[string]string{
		"status":   "ok",
		"hostname": fmt.Sprintf("%s-%s", SVC_A, hostname),
		"time":     time.Now().UTC().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}
