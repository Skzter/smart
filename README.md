# S.M.A.R.T - Software for Mockserver and Automated Resource Testing

[![Pipeline Status](https://gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/badges/main/pipeline.svg)](https://gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/pipelines)
[![Coverage](https://gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/badges/main/coverage.svg)](https://gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/commits/main)

## Applications

| Application | Description | C4 Doc |
|-------------|-------------|--------|
| **Autotester** | Core backend for LLM-powered test creation, execution, and management (Playwright, chat, storage). | [Component doc](docs/architecture/03-component/autotester.md) |
| **Suproxy** | Proxy between apps under test and supplier systems; request/response transformation, caching, validation. | [Component doc](docs/architecture/03-component/suproxy.md) |
| **Frontend** | Web UI (Svelte/TypeScript) for chat with the LLM, test runs, and results. | [Component doc](docs/architecture/03-component/frontend.md) |
| **MCP** | Model Context Protocol server so LLM systems (Claude/GPT) can call Autotester via structured tools. | [Component doc](docs/architecture/03-component/mcp.md) |

Full architecture (context, containers, code-level flows): [docs/architecture/README.md](docs/architecture/README.md).

## Bootstrap

**Prerequisites:** Go, [Task](https://taskfile.dev/), [Doppler](https://www.doppler.com/) (for secrets).

1. Clone and enter the repo.
2. Copy env: `cp .env.example .env` and set `DB_URL` if running DB-backed services locally.
3. Run setup:
   ```bash
   task setup
   ```
   This configures githooks, runs `go mod tidy`, and runs `doppler login` / `doppler setup -c dev -p htwk-projekt` for secrets.
4. (Optional) Start Postgres for Autotester: `task postgres-up` from repo root.

## How to run

- **Single services (with hot reload via Air):**
  - Suproxy: `task run-suproxy`
  - Autotester (builds frontend first): `task run-autotester`
  - Autotester headless: `task run-autotester-headless`
- **Full stack (Docker):** from repo root, `task compose-local` (uses `deployments/compose.dev.yml`). Stop with `task compose-local-down`.

Build binaries: `task build`. Frontend dev build: `task build-frontend-dev` (in `web/`).
