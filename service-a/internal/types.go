package internal

import "net/http"

type EventType string

const (
	PutEvent    EventType = "put"
	DeleteEvent EventType = "delete"
)

type Service struct {
	Cli  *http.Client
	Pool *PoolMap
}
