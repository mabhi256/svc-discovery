package internal

import (
	"os"
	"strings"
)

func GetEndpoints() []string {
	pointsStr := os.Getenv("ETCD_ENDPOINTS")
	if pointsStr == "" {
		return []string{"localhost:2379"}
	}

	return strings.Split(pointsStr, ",")
}

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8500"
	}

	return port
}
