# Level 3: Autotester Components

The Autotester service is the core of S.M.A.R.T, responsible for LLM-powered test creation, execution, and management. This document describes its internal component structure.

## Diagram

![Autotester](diagrams/autotester.mmd.svg)
See [autotester.mmd](diagrams/autotester.mmd) for the Mermaid source. The SVG is generated from it (see [Regenerating SVGs](../README.md#regenerating-svgs) in the architecture README).

## Architecture Overview

The Autotester follows a layered architecture:
- **Application Layer**: Router, middleware, HTTP configuration
- **Domain Layer**: Handlers, services, repositories, entities
- **Infrastructure Layer**: Database access, external API clients

## Components

### Application Layer

#### Router (`application/router.go`)
**Responsibilities:**
- HTTP server initialization
- Route registration and configuration
- Middleware setup (CORS, recovery, Datadog tracing)
- Static file serving (frontend assets)

**Key Routes:**
- `POST /api/v1/chat` - Chat with LLM
- `POST /api/v1/validate` - Validate prompts
- `POST /api/v1/run` - Execute tests
- `POST /api/v1/auth/generate` - Generate internal tokens
- `GET /api/v1/tests` - List tests
- `GET /api/v1/chats` - List chat sessions
- `GET /api/v1/groups` - List test groups

#### Internal-Only Middleware
**Responsibilities:**
- CIDR-based IP validation
- Protects sensitive endpoints (`/auth/generate`)
- Blocks external access attempts

**Allowed Networks:**
- `127.0.0.0/8` - Localhost
- `::1/128` - IPv6 localhost
- `10.89.0.0/8` - Podman default bridge
- `172.16.0.0/12` - Docker default bridge

#### SSE Middleware
**Responsibilities:**
- Server-Sent Events header configuration
- Real-time log streaming for test execution

---

### Domain Layer - Handlers

#### AutotesterController (`domain/handler/`)
**Responsibilities:**
- HTTP request/response handling
- Request validation and binding
- Coordinates service layer calls
- Error handling and response formatting

**Key Handler Methods:**
- `HandleChatRequest` - Process chat messages
- `HandleChatRequestValidity` - Validate prompts
- `HandleRunContainer` - Execute tests in Docker
- `HandleLogRequest` - Stream test logs (SSE)
- `HandleSaveLocalRequest` - Save tests locally
- `HandleGetRemoteTestcase` - Retrieve tests
- `HandleGenerateToken` - Generate auth tokens
- `HandleGetChats` - List user chats
- `HandleGetGroups` - List groups
- `HandleCreateGroup` - Create test groups
- `HandleAssignChatToGroups` - Assign chats to groups

---

### Domain Layer - Services

#### ChatManager (`domain/service/chatManager.go`)
**Responsibilities:**
- Create, save, and update chat sessions
- Orchestrate loading/updating of chat context
- No LLM service dependency; test generation is handled elsewhere (e.g. GeneratePrompt service, validation before)

**Dependencies:**
- ChatStorageService

#### ChatStorageService (`domain/service/chatStorageService.go`)
**Responsibilities:**
- Chat persistence logic (retrieval, updates)
- Does not perform chat management; ChatManager handles that

**Dependencies:**
- ChatStorageRepository (S3-backed)

#### TestcaseLocalStorageService (`domain/service/testcaseLocalStorageService.go`)
**Responsibilities:**
- Local test storage management
- Does not perform file system operations (repository does); does not perform test template generation

**Dependencies:**
- TestcaseLocalStorageRepository

#### TestcaseStorageService (`domain/service/testcaseStorageService.go`)
**Responsibilities:**
- Remote test storage in S3 (via repository)
- Test versioning, retrieval, and listing

#### Docker Service (`domain/service/docker.go`)
**Responsibilities:**
- Docker container management for test execution
- Playwright container orchestration
- Provides container output for reading (streaming via SSE is done by handler/controller)

**Key Operations:**
- Start Playwright containers
- Execute tests in isolated environments
- Expose container output for handlers to stream
- Cleanup containers

#### LLM Test Suite Service (`domain/service/llmTestSuite.go`)
**Responsibilities:**
- Test generation from natural language
- Prompt engineering and construction
- LLM response parsing
- Test code extraction and validation

#### Validator Service (`domain/service/validator.go`)
**Responsibilities:**
- Prompt validation logic
- Test code validation
- Request validation

#### Auth Service (`domain/service/auth.go`)
**Responsibilities:**
- Token generation for internal services
- Token validation
- JWT handling

**Dependencies:**
- TokenDatabase (PostgreSQL)

#### GroupManager & GroupStorage (`domain/service/group*.go`)
**Responsibilities:**
- Test group management
- Group creation and assignment
- Chat-to-group associations

---

### Domain Layer - Repositories

#### ChatStorageRepository (`domain/repository/chatStorageRepository.go`)
**Responsibilities:**
- Chat data in S3 (Parquet)
- CRUD operations on chat data, user chat associations

**Technology:** S3 API (not PostgreSQL)

#### TestcaseLocalStorageRepository (`domain/repository/testcaseLocalStorageRepository.go`)
**Responsibilities:**
- File system operations for tests
- Local test persistence

#### TokenDatabase (`domain/repository/database.go`)
**Responsibilities:**
- Token storage and retrieval
- Database operations for authentication

#### GroupStorage Repository (`infrastructure/repository/groupStorage.go`)
**Responsibilities:**
- Group data persistence (S3, not PostgreSQL)
- Group-chat relationship management

**Technology:** S3 API

---

### Domain Layer - Entities

Key domain entities:
- `User` - User information
- `Chat` - Chat session data
- `ChatSummary` - Chat overview
- `ChatRequest` - Incoming chat requests
- `LLMResponse` - LLM response structure
- `RunTestRequest` - Test execution request
- `LocalSaveRequest` - Save test request
- `ContainerInfo` - Docker container metadata
- `Group` - Test group entity
- `Token` - Authentication token

---

### Infrastructure Layer

#### Database (`infrastructure/database/`)
**Technology:** PostgreSQL with SQLC
**Responsibilities:**
- Database connection management
- Query execution
- Transaction management

**Tables:**
- Tokens (token management only)

**Generated Code:** SQLC generates type-safe Go code from SQL

---

### Configuration

#### Pkl Config (`domain/config/`)
**Configuration Files:**
- `Config.pkl.go` - Main config structure
- `Prompts.pkl.go` - LLM prompt templates

**Configuration Includes:**
- Server port and host
- Database connection strings
- Redis configuration
- LLM API keys (from Doppler)
- Playwright settings
- Datadog configuration

---

## Component Interactions

### Test Creation Flow
1. **Router** receives chat request
2. **Middleware** validates (e.g. auth)
3. **GeneratePrompt service** and **validation** run (e.g. `/validate`) before generation
4. **ChatManager** creates or orchestrates loading/updating chat; **LLM Test Suite** is not part of this orchestration
5. **ChatStorageService** persists chat history (via ChatStorageRepository to S3)
6. **Response** returned to user

### Test Execution Flow
1. **Router** receives run request
2. **AutotesterController** validates
3. **Docker Service** creates Playwright container and provides container output
4. **Container** executes test
5. **Handler/Controller** streams logs asynchronously via SSE from the output Docker Service provides
6. **TestcaseStorageService** (via repository) saves results
7. **Response** with results returned

### Authentication Flow
1. **Authenticated user** (e.g. via Frontend) calls `/auth/generate` (internal only, from allowed networks)
2. **Internal-Only Middleware** validates IP
3. **AuthController** receives request
4. **Auth Service** generates JWT token
5. **TokenDatabase** stores token
6. **Token** returned to caller

**MCP:** Does not call `/auth/generate`. Sends the bearer token in the `Authorization` header when calling Autotester (token obtained by the authenticated user, e.g. via Frontend).

### Group Management Flow
1. **Router** receives group request
2. **AutotesterController** validates
3. **GroupManager** processes request
4. **GroupStorage Service** persists data
5. **GroupStorage Repository** writes to DB

---

## Key Design Patterns

1. **Layered Architecture**: Clear separation of concerns (Application → Domain → Infrastructure)
2. **Repository Pattern**: Abstract data access through repositories
3. **Service Layer**: Business logic encapsulation
4. **Dependency Injection**: Constructor injection via Wire
5. **CQRS-lite**: Separate read/write operations
6. **Middleware Pattern**: Cross-cutting concerns (auth, logging)

## Technology Stack

- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 17 with SQLC (tokens only)
- **Caching**: Valkey (Redis-compatible)
- **Object Storage**: S3 (chats, testcases, groups)
- **Monitoring**: Datadog APM
- **Configuration**: Pkl
- **Testing**: Go standard testing + Mockery
- **Dependency Injection**: Wire
- **Container Runtime**: Docker API client

## Security Considerations

1. **Network-level Protection**: Internal-only middleware for sensitive endpoints
2. **Token-based Auth**: JWT for service-to-service communication
3. **Input Validation**: All requests validated before processing
4. **Database Security**: Parameterized queries via SQLC
5. **Container Isolation**: Tests run in isolated Docker containers
