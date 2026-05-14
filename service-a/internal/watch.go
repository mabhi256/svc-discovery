package internal

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func Watch(cli *http.Client, pool *PoolMap) {
	registries := GetRegistryURL()
	i := 0
	for {
		registry := registries[i]
		url := fmt.Sprintf("%s/watch/%s", registry, SVC_B)

		// Blocks until the connection drops.
		// So the loop only moves to the next registry after a failure
		if err := streamWatch(cli, url, pool); err != nil {
			log.Printf("watch %s lost: %v", registry, err)
		}
		i = (i + 1) % len(registries)
		time.Sleep(2 * time.Second)
	}
}

func streamWatch(cli *http.Client, url string, pool *PoolMap) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	extractField := func(line string, field string) string {
		return strings.TrimSpace(strings.TrimPrefix(line, field+":"))
	}

	scanner := bufio.NewScanner(resp.Body)
	var evType EventType

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event:"):
			evType = EventType(extractField(line, "event"))

		case strings.HasPrefix(line, "data:"):
			addr := extractField(line, "data")

			switch evType {
			case PutEvent:
				pool.Add(SVC_B, addr)
			case DeleteEvent:
				pool.Remove(SVC_B, addr)
			}
		}
	}

	return scanner.Err()
}
