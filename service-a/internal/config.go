package internal

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const SVC_A = "svc-a"
const SVC_B = "svc-b"

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

	// Google DNS trick for finding the outbound IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	port := GetPort()
	if err != nil {
		return "127.0.0.1:" + port
	}
	defer conn.Close()

	ip := conn.LocalAddr().(*net.UDPAddr).IP.String()

	return fmt.Sprintf("%s:%s", ip, port)
}
