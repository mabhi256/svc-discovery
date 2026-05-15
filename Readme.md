# Service Discovery

```bash
docker compose up -d --build

# see A's endpoint pool
curl http://localhost:9010/pool
# {"svc-b":["172.18.0.12:9000"],"svc-c":["172.18.0.6:9000"]}

# raw SSE stream
curl -N http://localhost:8500/watch/svc-b

# scale up upstream services
docker compose up -d --scale service-b=2

# verify updated pool
curl http://localhost:9010/pool
# {"svc-b":["172.18.0.8:9000","172.18.0.9:9000"],"svc-c":["172.18.0.7:9000"]}

# call multiple times and verify round-robin call
curl http://localhost:9010/proxy/svc-b
# {"hostname":"svc-b-d11940be8449","status":"ok","time":"2026-05-15T07:51:55Z"}
# {"hostname":"svc-b-2f31baef9aa7","status":"ok","time":"2026-05-15T07:51:59Z"}
curl http://localhost:9010/proxy/svc-c

# simulate a crash
docker compose stop service-b-1

# pool shrinks automatically
curl http://localhost:9010/pool
```
