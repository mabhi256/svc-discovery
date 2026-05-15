package internal

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// service-A ---[network partition]--- service-X
//      \                                   /
//       \________ registry (etcd) ________/

func (service *Service) PingUpstream() func() {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				service.sendPings()
			}
		}
	}()

	var once sync.Once
	return func() { once.Do(func() { close(stop) }) }
}

func (service *Service) sendPings() {
	for svc, endpoints := range service.Snapshot() {
		for _, addr := range endpoints {
			url := fmt.Sprintf("http://%s/health", addr)
			resp, err := service.Cli.Get(url)
			if err != nil {
				service.Pool.SetActive(svc, addr, false)
				continue
			}
			// Drain the response body without saving its content. If you just
			// Close() without draining, the http client can't reuse that connection
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				service.Pool.SetActive(svc, addr, false)
			} else {
				service.Pool.SetActive(svc, addr, true)
			}
		}
	}
}
