package internal

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func Heartbeat(cli *http.Client, leaseID string, period int) func() {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Duration(period) * time.Second)
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				if err := heartbeatHandler(cli, leaseID); err != nil {
					log.Printf("Heartbeat error: %v", err)
				}
			}
		}
	}()

	var once sync.Once
	return func() { once.Do(func() { close(stop) }) }
}

func heartbeatHandler(cli *http.Client, leaseID string) error {

	for _, registryUrl := range GetRegistryURL() {
		url := fmt.Sprintf("%s/leases/%s/heartbeat", registryUrl, leaseID)
		req, _ := http.NewRequest(http.MethodPut, url, nil)
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
