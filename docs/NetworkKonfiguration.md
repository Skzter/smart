## Security & Network Configuration

### Endpoint Access Restrictions

The `/api/v1/auth/generate` endpoint is restricted to **internal networks only** to prevent unauthorized token generation. Access is controlled through two layers:

#### 1. Network Layer (nginx)
Located in `deployments/nginx/default.conf`, the endpoint is only accessible from:
- `127.0.0.1` / `::1` - Localhost (IPv4/IPv6)
- `172.16.0.0/12` - Docker default bridge networks
- `192.168.0.0/16` - Docker Compose internal networks

External requests receive a `403 Forbidden` response.

#### 2. Application Layer (Gin Middleware)
The `internalOnlyMiddleware` in `internal/autotester/application/router.go` validates client IP addresses against allowed CIDR ranges. This provides defense-in-depth even if the nginx proxy is bypassed.

### Deployment Requirements

**Production:**
- The autotester service should **only** be exposed through the nginx reverse proxy
- Direct port mappings to the host (e.g., `8081:8081`) should be avoided
- Only services within the Docker network (e.g., `autotester-mcp`) can call the auth endpoint directly

**Development:**
- Port mappings are allowed for local testing
- The middleware still enforces localhost/Docker network restrictions

### Auth0 Integration

Auth0 is used **only for frontend authentication** at HTWK. The `/api/v1/auth/generate` endpoint:
- Does **not** validate Auth0 tokens
- Is intended for internal service-to-service communication (e.g., MCP ↔ Autotester)
- Will be replaced or removed when transitioning away from HTWK-specific Auth0 setup

### Testing Network Restrictions

#### Test 1: Direct Access (Application Layer) - Should Work
When running locally without nginx:
```bash
# Start autotester directly
go run cmd/autotester/main.go

# Test from localhost - should succeed (200 OK)
curl -X POST http://localhost:8081/api/v1/auth/generate \
  -H "Content-Type: application/json" \
  -d '{"userId": "test-user"}'
```

#### Test 2: Through Docker Network - Should Work
When running via Docker Compose:
```bash
# Start services
docker-compose -f deployments/compose.dev.yml up -d

# Test from another container in same network - should succeed
docker exec -it <mcp-container-name> curl -X POST http://autotester:8081/api/v1/auth/generate \
  -H "Content-Type: application/json" \
  -d '{"userId": "test-user"}'

# Or test from host via nginx (localhost mapped) - should succeed
curl -X POST http://localhost:8081/api/v1/auth/generate \
  -H "Content-Type: application/json" \
  -d '{"userId": "test-user"}'
```

#### Test 3: External IP Simulation - Should Fail
Test the middleware blocking external IPs:
```bash
# Modify request to simulate external IP
# This requires setting X-Forwarded-For header (if nginx passes it through)
curl -X POST http://localhost:8081/api/v1/auth/generate \
  -H "Content-Type: application/json" \
  -H "X-Forwarded-For: 8.8.8.8" \
  -d '{"userId": "test-user"}'

# Expected response: 403 Forbidden
# {"error":"Forbidden: Endpoint only accessible from internal networks"}
```

#### Test 4: Nginx Layer Blocking - Should Fail (Production)
When deployed with nginx in production (external access):
```bash
# Try to access from outside Docker network
curl -X POST https://your-production-domain.com/api/v1/auth/generate \
  -H "Content-Type: application/json" \
  -d '{"userId": "test-user"}'

# Expected response: 403 Forbidden (from nginx)
```

#### Verify Logs
Check that blocked attempts are logged:
```bash
# Application logs
docker logs <autotester-container-name> | grep "Blocked external access"

# Nginx logs (in production)
docker logs <nginx-container-name> | grep "403"
```

#### Integration Test Script
Create a test script to verify all scenarios:
```bash
#!/bin/bash
# test-auth-endpoint.sh

echo "Testing localhost access..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8081/api/v1/auth/generate \
  -H "Content-Type: application/json" \
  -d '{"userId": "test-user"}')

if [ "$RESPONSE" = "200" ]; then
  echo "✅ Localhost access: PASSED"
else
  echo "❌ Localhost access: FAILED (got $RESPONSE)"
fi

# Add more test cases as needed
```

