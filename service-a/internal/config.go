package internal

import (
	"os"
	"strings"
)

const SVC_A = "svc-a"

func GetUpstreamServices() []string {
	val := os.Getenv("UPSTREAM_SERVICES")
	if val == "" {
		return []string{"svc-b"}
	}
	return strings.Split(val, ",")
}

func GetRegistryURL() []string {
	pointsStr := os.Getenv("REGISTRY_URL")
	if pointsStr == "" {
		return []string{"http://localhost:8500"}
	}

	return strings.Split(pointsStr, ",")
}

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "9000"
	}

	return port
}

func GetAdvertiseAddr() string {
	addr := os.Getenv("ADVERTISE_ADDR")
	if addr != "" {
		return addr
	}
	return "service-a:9000"
}
