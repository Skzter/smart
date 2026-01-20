# Level 2: Container

The Container diagram zooms into the S.M.A.R.T system to show the high-level technology choices and how responsibilities are distributed across containers (applications, services, databases).

## Diagram

See [containers.mmd](diagrams/containers.mmd) for the Mermaid source (diagram embedded below renders in GitLab/GitHub).

## Containers Overview

### 1. Autotester Service
**Technology:** Go, Gin Framework  
**Port:** 8081  
**Repository:** `cmd/autotester/`

**Responsibilities:**
- Provides REST API for test automation
- Chat interface for LLM-based test creation
- Test execution via Playwright integration
- User and chat session management
- Test result storage and retrieval

**Key Endpoints:**
- `POST /api/v1/chat` - Chat with LLM for test creation
- `POST /api/v1/validate` - Validate test prompts
- `POST /api/v1/auth/generate` - Internal auth token generation (restricted)
- Test execution endpoints

**Dependencies:**
- PostgreSQL (persistent storage)
- Redis (caching, sessions)
- LLM Service (via HTTP)
- Playwright (embedded)

**Configuration:** `configs/autotester.pkl`

---

### 2. Autotester MCP
**Technology:** Go, Model Context Protocol  
**Port:** 8084  
**Repository:** `cmd/mcp/`

**Responsibilities:**
- Implements Model Context Protocol server
- Bridges between LLM and Autotester service
- Exposes tools/functions for LLM to call
- Manages context and state for LLM interactions

**Key Features:**
- Tool registry for LLM-callable functions
- Internal authentication with Autotester
- Protocol adapter for MCP specification

**Dependencies:**
- Autotester Service (via HTTP)

**Configuration:** `configs/mcp.pkl`

---

### 3. Suproxy Service
**Technology:** Go, Gin Framework  
**Port:** 8080  
**Repository:** `cmd/suproxy/`

**Responsibilities:**
- Proxy between supplier-d and supplier systems
- Request/response transformation
- Header and body manipulation
- Tag processing
- Destination routing

**Key Endpoints:**
- `POST /api/v1/Offerlist` - Proxy offerlist requests

**Dependencies:**
- Supplier Systems (external)
- Redis (optional caching)

**Configuration:** `configs/suproxy.pkl`

---

### 4. Web Frontend
**Technology:** Svelte, TypeScript  
**Port:** (served by application servers)  
**Repository:** `web/`

**Responsibilities:**
- User interface for test creation and management
- Chat interface for LLM interaction
- Test result visualization
- User authentication flow

**Key Components:**
- Chat interface
- Test manager
- Result viewer
- Authentication module

**Dependencies:**
- Autotester Service (via REST API)
- Auth0 (authentication)

---

### 5. Nginx Reverse Proxy
**Technology:** Nginx  
**Configuration:** `deployments/nginx/default.conf`

**Responsibilities:**
- Request routing to backend services
- TLS termination (in production)
- Security layer for sensitive endpoints
- Load balancing (if scaled)

**Security Rules:**
- `/api/v1/auth/generate` restricted to internal networks only
- CIDR-based access control (127.0.0.1, 172.16.0.0/12, 192.168.0.0/16)

---

### 6. PostgreSQL Database
**Technology:** PostgreSQL 17 Alpine  
**Port:** 5432  
**Image:** `postgres:17-alpine`

**Responsibilities:**
- Persistent storage for:
  - User accounts
  - Chat sessions and history
  - Test definitions
  - Test execution results
  - System configuration

**Schema Management:** SQLC (`sqlc.yaml`)

---

### 7. Redis/Valkey
**Technology:** Valkey (Redis-compatible)  
**Image:** `valkey/valkey:alpine`

**Responsibilities:**
- Session management
- Caching layer
- Temporary data storage
- Rate limiting data

---

### 8. Mock Services

#### Frontend Mock
**Technology:** Node.js/Svelte  
**Port:** 8082  
**Image:** Custom GitLab registry image

**Purpose:** Mock frontend application for testing Suproxy integration

#### Supplier Mock
**Technology:** Node.js  
**Port:** 8083  
**Image:** Custom GitLab registry image

**Purpose:** Mock supplier API for testing without external dependencies

---

### 9. Datadog Agent
**Technology:** Datadog Agent  
**Ports:** 8126 (APM), 8125 (StatsD)  
**Image:** `datadog/agent:latest`

**Responsibilities:**
- Collect APM traces from services
- Aggregate metrics and logs
- Forward to Datadog cloud
- Container monitoring

---

## Container Communication

### Internal Network Communication
All services communicate via Docker internal network:
- Service discovery by container name
- No external exposure except via Nginx

### Authentication Flow
1. User → Frontend → Auth0 (OAuth)
2. Frontend → Autotester (with Auth0 token)
3. MCP → Autotester (internal token from `/auth/generate`)

### Test Creation Flow
1. User → Frontend (chat message)
2. Frontend → Autotester (`/chat`)
3. Autotester → MCP (context provision)
4. MCP → LLM Service (prompt)
5. LLM → MCP (generated code)
6. MCP → Autotester (test code)
7. Autotester → Frontend (response)

### Test Execution Flow
1. Frontend → Autotester (execute test request)
2. Autotester → Playwright (browser automation)
3. Playwright → Web Application Under Test
4. Playwright → Autotester (results)
5. Autotester → PostgreSQL (store results)
6. Autotester → Frontend (test results)

### Proxy Flow
1. Frontend Under Test → Nginx
2. Nginx → Suproxy
3. Suproxy → Supplier System
4. Supplier System → Suproxy → Nginx → Frontend

## Technology Stack Summary

| Container | Language | Framework | Port | Database |
|-----------|----------|-----------|------|----------|
| Autotester | Go | Gin | 8081 | PostgreSQL, Redis |
| Autotester MCP | Go | Custom MCP | 8084 | - |
| Suproxy | Go | Gin | 8080 | Redis (optional) |
| Frontend | TypeScript | Svelte | - | - |
| Nginx | - | Nginx | 80/443 | - |

## Deployment

### Development
- Docker Compose: `deployments/compose.dev.yml`
- Live reload with Air
- Direct port mappings for debugging

### Production
- Docker Compose: `deployments/compose.prod.yml`
- All services behind Nginx
- No direct external port exposure
- Environment-specific configuration via Doppler

## Configuration Management

All services use **Pkl** (Apple's configuration language) for type-safe configuration:
- `configs/autotester.pkl`
- `configs/mcp.pkl`
- `configs/suproxy.pkl`
- `configs/shared/taglist.pkl`

## Security Considerations

1. **Network Isolation:** Services communicate only via internal Docker network
2. **Endpoint Protection:** `/auth/generate` restricted by Nginx and application middleware
3. **Secret Management:** All secrets stored in Doppler, never in code
4. **Authentication:** Auth0 for frontend, internal tokens for service-to-service
5. **Observability:** All services traced via Datadog for security monitoring
