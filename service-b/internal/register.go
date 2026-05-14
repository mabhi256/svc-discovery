package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (client *RegistryClient) Register() (*RegisterResponse, error) {
	body, err := json.Marshal(RegisterRequest{Service: SVC, Address: GetAdvertiseAddr()})
	if err != nil {
		return nil, fmt.Errorf("marshal register request: %w", err)
	}

	for _, url := range GetRegistryURL() {
		resp, err := client.HttpCli.Post(url+"/services", "application/json", bytes.NewReader(body))
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		var result RegisterResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if decodeErr != nil {
			continue
		}

		return &result, nil
	}

	return nil, fmt.Errorf("POST /services: all registries unreachable")
}
