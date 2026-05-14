package internal

import (
	"fmt"
	"net/http"
)

func (client *RegistryClient) Deregister(leaseID string) error {

	for _, registryUrl := range GetRegistryURL() {
		url := fmt.Sprintf("%s/services/%s/%s", registryUrl, SVC, leaseID)
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		resp, err := client.HttpCli.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		return nil
	}

	return fmt.Errorf("DELETE /leases/%s: all registries unreachable", leaseID)
}
