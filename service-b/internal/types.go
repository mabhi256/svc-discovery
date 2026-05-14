package internal

const SVC = "svc-b"

type RegisterRequest struct {
	Service string `json:"service" validate:"required,min=1,max=64"`
	Address string `json:"address" validate:"required,hostname_port"`
}

type RegisterResponse struct {
	LeaseID   string `json:"lease_id"`
	Revision  int64  `json:"revision"`
	Heartbeat int    `json:"heartbeat"`
}
