# Level 3: Suproxy Components

The Suproxy service acts as an intelligent proxy between frontend applications (supplier-d) and external supplier systems, providing request/response transformation, caching, and validation.

## Diagram

See [suproxy.mmd](diagrams/suproxy.mmd) for the Mermaid source (diagram embedded below renders in GitLab/GitHub).

## Architecture Overview

Suproxy follows a simplified layered architecture optimized for request proxying and transformation:
- **Application Layer**: Router, middleware
- **Domain Layer**: Handler, services, repositories
- **Infrastructure Layer**: External API clients

## Components

### Application Layer

#### Router (`application/router.go`)
**Responsibilities:**
- HTTP server initialization
- Route registration
- CORS configuration
- Middleware setup (recovery, Datadog tracing)

**Key Routes:**
- `POST /api/v1/Offerlist` - Proxy offerlist requests to suppliers

**CORS Policy:**
- Allow all origins (`*`)
- Support credentials
- Methods: POST, GET, PUT, DELETE, OPTIONS
- Headers: Content-Type, Authorization, etc.

#### CORS Middleware
**Responsibilities:**
- Cross-Origin Resource Sharing headers
- Preflight request handling
- Enable frontend access

---

### Domain Layer - Handler

#### SuproxyController (`domain/handler/handler.go`)
**Responsibilities:**
- HTTP request handling
- Request validation and parsing
- Response transformation
- Error handling

**Handler Method:**
- `PostOfferlist` - Main proxy endpoint handler

**Processing Flow:**
1. Parse incoming request
2. Validate request structure
3. Check cache (if enabled)
4. Transform request for supplier
5. Forward to supplier system
6. Transform response
7. Cache response (if enabled)
8. Return to client

---

### Domain Layer - Services

#### Validator Service (`domain/service/validation.go`)
**Responsibilities:**
- Request validation logic
- Header validation
- Body structure validation
- Destination validation

**Validation Checks:**
- Required fields present
- Valid destination format
- Valid header structure
- Body format correctness

#### Tag Search Service (`domain/service/tagSearchService.go`)
**Responsibilities:**
- Tag extraction from requests
- Tag-based routing logic
- Tag list synchronization

**Features:**
- Parse tags from request body
- Match tags to suppliers
- Route based on tag patterns

#### Taglist Sync Service (`domain/service/taglistsync.go`)
**Responsibilities:**
- Synchronize tag lists from configuration
- Update tag mappings
- Maintain tag consistency

**Configuration Source:** `configs/shared/taglist.pkl`

#### Database Service (`domain/service/database.go`)
**Responsibilities:**
- Business logic for data persistence
- Query orchestration
- Transaction management

**Operations:**
- Store request/response pairs
- Retrieve historical data
- Analytics data collection

#### Cache Service (`domain/service/cache.go`)
**Responsibilities:**
- Caching strategy implementation
- TTL policy management
- Cache key generation
- Cache invalidation

**Caching Strategy:**
- Request-based cache keys
- Configurable TTL policies
- Destination-aware caching
- Header-based cache control

---

### Domain Layer - Repositories

#### Cache Repository (`domain/repository/cache.go`)
**Responsibilities:**
- Redis operations for caching
- Cache read/write operations
- TTL management

**Technology:** Go-Redis client

**Operations:**
- `Get(key)` - Retrieve cached response
- `Set(key, value, ttl)` - Store response
- `Delete(key)` - Invalidate cache entry
- `Exists(key)` - Check cache presence

#### Database Repository (`domain/repository/database.go`)
**Responsibilities:**
- PostgreSQL operations
- Request/response logging
- Analytics data storage

**Schema:**
- Request headers
- Request body
- Response body
- Timestamp
- Destination
- Processing time

---

### Domain Layer - Entities

#### Request (`domain/entity/request.go`)
**Structure:**
- `Header` - HTTP headers map
- `Tags` - Optional tag string
- `Destination` - Target supplier URL/IP
- `Body` - JSON request body string

#### Response (`domain/entity/response.go`)
**Structure:**
- Raw response string from supplier
- Can be JSON or other formats

#### ValidationResponse (`domain/entity/validation_response.go`)
**Structure:**
- Validation result
- Error messages
- Field-level errors

#### CacheEntry (`domain/entity/cacheEntry.go`)
**Structure:**
- Cached response data
- Metadata (timestamp, TTL)
- Cache key

#### CacheTTLPolicy (`domain/entity/cacheTTLPolicy.go`)
**Structure:**
- TTL duration per destination
- Cache control policies

#### DBEntry (`domain/entity/dbEntry.go`)
**Structure:**
- Request data
- Response data
- Metadata for analytics

---

### Configuration

#### Pkl Config (`domain/config/`)
**Configuration Files:**
- `Config.pkl.go` - Main configuration
- `RedisConfig.pkl.go` - Redis settings
- `Prompts.pkl.go` - Transformation templates

**Configuration Includes:**
- Server port (8080)
- Redis connection
- Database connection
- Supplier endpoints
- Cache TTL policies
- Tag list configuration

---

## Component Interactions

### Request Proxy Flow
1. **Router** receives `/api/v1/Offerlist` request
2. **CORS Middleware** adds headers
3. **SuproxyController** receives request
4. **Validator** validates request structure
5. **Cache Service** checks for cached response
   - If hit: return cached response
   - If miss: continue
6. **Tag Search Service** extracts tags and determines destination
7. **Controller** transforms request headers/body
8. **HTTP Client** forwards to supplier system
9. **Supplier** processes and returns response
10. **Cache Service** stores response (if cacheable)
11. **Database Service** logs request/response (optional)
12. **Controller** transforms response
13. **Response** returned to client

### Cache Hit Flow
1. Request arrives
2. **Cache Service** generates cache key
3. **Cache Repository** queries Redis
4. Cache hit found
5. Response returned immediately (no supplier call)

### Tag-based Routing Flow
1. Request contains tags in body or query
2. **Tag Search Service** parses tags
3. **Taglist Sync** provides tag-to-supplier mapping
4. **Controller** selects appropriate destination
5. Request forwarded to correct supplier

---

## Key Design Patterns

1. **Proxy Pattern**: Central pattern for request forwarding
2. **Cache-Aside Pattern**: Check cache, then load from source
3. **Strategy Pattern**: Different transformation strategies per supplier
4. **Repository Pattern**: Abstract data access
5. **Middleware Pattern**: Cross-cutting concerns

## Technology Stack

- **Framework**: Gin (HTTP router)
- **Caching**: Redis/Valkey with TTL policies
- **Database**: PostgreSQL (optional logging)
- **Monitoring**: Datadog APM
- **Configuration**: Pkl
- **HTTP Client**: Go standard library

## Performance Considerations

1. **Caching Strategy**: Redis-based caching reduces supplier load
2. **Connection Pooling**: Reuse HTTP connections to suppliers
3. **Async Logging**: Non-blocking database writes
4. **TTL Policies**: Configurable per-destination cache duration

## Security Considerations

1. **CORS Configuration**: Controlled cross-origin access
2. **Request Validation**: Prevent malformed requests
3. **Destination Whitelist**: Only allow configured suppliers
4. **Header Sanitization**: Remove sensitive headers before forwarding
5. **Rate Limiting**: (Future) Prevent abuse
