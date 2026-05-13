package internal

import (
	"github.com/go-playground/validator/v10"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	LeaseTTL  = 10 // seconds
	Heartbeat = 3  // seconds
)

type Registry struct {
	Etcd     *clientv3.Client
	Validate *validator.Validate
}

func NewRegistry(cli *clientv3.Client, validate *validator.Validate) *Registry {
	return &Registry{
		Etcd:     cli,
		Validate: validate,
	}
}

type RegisterRequest struct {
	Service string `json:"service" validate:"required,min=1,max=64"`
	Address string `json:"address" validate:"required,hostname_port"`
}

type RegisterResponse struct {
	LeaseID   string `json:"lease_id"`
	Heartbeat int    `json:"heartbeat"`
}

type Endpoint struct {
	LeaseID string `json:"lease_id"`
	Address string `json:"address"`
}

type EventType string

const (
	PutEvent    EventType = "put"
	DeleteEvent EventType = "delete"
)

type WatchEvent struct {
	Type     EventType `json:"type"` // "put" | "delete"
	Endpoint Endpoint  `json:"endpoint"`
}
