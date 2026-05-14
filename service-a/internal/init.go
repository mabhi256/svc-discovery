package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func (service *Service) Init() error {
	for _, registry := range GetRegistryURL() {
		url := fmt.Sprintf("%s/services/%s", registry, SVC_B)
		resp, err := service.Cli.Get(url)
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		var result []string
		decodeErr := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if decodeErr != nil {
			continue
		}

		service.Pool.Add(SVC_B, result...)

		revision, _ := strconv.ParseInt(resp.Header.Get("X-Etcd-Revision"), 10, 64)
		service.Pool.SetRevision(revision)
		return nil
	}

	return fmt.Errorf("GET /services/%s: all registries unreachable", SVC_B)
}
