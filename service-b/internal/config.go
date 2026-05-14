package internal

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func GetRegistryURL() []string {
	pointsStr := os.Getenv("REGISTRY_URL")
	if pointsStr == "" {
		return []string{"localhost:8500"}
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
	// UDP is connectionless, net.Dial does not initiate a "handshake" like TCP.
	// It merely prepares the socket.
	// The OS routing table decides which local IP to bind to the socket based
	// on the destination address (8.8.8.8).
	// Port doesn't matter since no connection is actually established.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	port := GetPort()
	if err != nil {
		return "127.0.0.1:" + port
	}
	defer conn.Close()

	ip := conn.LocalAddr().(*net.UDPAddr).IP.String()

	return fmt.Sprintf("%s:%s", ip, port)
}
