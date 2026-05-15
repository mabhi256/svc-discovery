package internal

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func Watch(cli *http.Client, svc string, pool *PoolMap) {
	delay := time.Second
	for {
		for _, registry := range GetRegistryURL() {
			url := fmt.Sprintf("%s/watch/%s", registry, svc)
			if rev := pool.Revision(); rev > 0 {
				url = fmt.Sprintf("%s?revision=%d", url, rev+1)
			}

			log.Printf("Watch request to %s\n", url)
			if err := streamWatch(cli, url, svc, pool); err != nil {
				log.Printf("watch %s lost: %v", registry, err)
			}
			time.Sleep(delay)
			delay = min(delay*2, 5*time.Second)
		}
	}
}

func streamWatch(cli *http.Client, url string, svc string, pool *PoolMap) error {
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
		case strings.HasPrefix(line, "id:"):
			if rev, err := strconv.ParseInt(extractField(line, "id"), 10, 64); err == nil {
				pool.SetRevision(rev)
			}

		case strings.HasPrefix(line, "event:"):
			evType = EventType(extractField(line, "event"))

		case strings.HasPrefix(line, "data:"):
			addr := extractField(line, "data")

			log.Printf("Watch event=%s addr=%s\n", evType, addr)
			switch evType {
			case PutEvent:
				pool.Add(svc, addr)
			case DeleteEvent:
				pool.Remove(svc, addr)
			}
		}
	}

	return scanner.Err()
}
