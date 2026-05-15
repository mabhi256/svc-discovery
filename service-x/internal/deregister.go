package internal

import (
	"fmt"
	"net/http"
)

func Deregister(cli *http.Client, leaseID string) error {

	for _, registryUrl := range GetRegistryURL() {
		url := fmt.Sprintf("%s/services/%s/%s", registryUrl, GetServiceName(), leaseID)
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		resp, err := cli.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusNoContent {
			resp.Body.Close()
			continue
		}

		return nil
	}

	return fmt.Errorf("DELETE /leases/%s: all registries unreachable", leaseID)
}
