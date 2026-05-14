# Service Discovery

```bash
docker compose up -d --build

# see A's endpoint pool
curl http://localhost:9010/pool  

# call twice and verify round-robin call
curl http://localhost:9010/call-b   

# raw SSE stream
curl -N http://localhost:8500/watch/svc-b  

# simulate a crash
docker compose stop service-b-1

# pool shrinks automatically
curl http://localhost:9010/pool    
```
