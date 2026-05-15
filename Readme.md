# Service Discovery

```bash
docker compose up -d --build

# see A's endpoint pool
curl http://localhost:9010/pool

# raw SSE stream
curl -N http://localhost:8500/watch/svc-b

# scale up upstream services
docker compose up -d --scale service-b=2

# verify updated pool
curl http://localhost:9010/pool
# call multiple times and verify round-robin call
curl http://localhost:9010/proxy/svc-b
curl http://localhost:9010/proxy/svc-c

# simulate a crash
docker compose stop service-b-1

# pool shrinks automatically
curl http://localhost:9010/pool
```
