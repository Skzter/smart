# S.M.A.R.T Architecture Documentation

This directory contains the C4 model architecture documentation for the S.M.A.R.T (Software for Mockserver and Automated Resource Testing) system.

## About C4 Model

The C4 model provides a hierarchical set of software architecture diagrams at different levels of abstraction:

1. **System Context** - Shows the system in its environment with users and external systems
2. **Container** - Shows high-level technology choices and how responsibilities are distributed
3. **Component** - Shows how containers are made up of components and their interactions
4. **Code** - Shows implementation details for selected complex areas

## Documentation Structure

### Level 1: System Context
- [System Context Diagram](01-context/README.md) - Overall system and its environment

### Level 2: Container
- [Container Diagram](02-container/README.md) - Services, databases, and infrastructure

### Level 3: Component
- [Autotester Components](03-component/autotester.md) - Autotester service internal structure
- [Suproxy Components](03-component/suproxy.md) - Suproxy service internal structure
- [Frontend Components](03-component/frontend.md) - Web frontend internal structure
- [MCP Components](03-component/mcp.md) - MCP service internal structure

### Level 4: Code
- [LLM Integration Flow](04-code/llm-flow.md) - Test generation via LLM
- [Authentication Flow](04-code/auth-flow.md) - Authentication and authorization

## Viewing Diagrams

Diagrams are Mermaid source (`.mmd`) with rendered SVGs (`.mmd.svg`) next to each file. The docs reference the SVG files so they display in GitLab, GitHub, and IDEs.

### Viewing in GitLab/GitHub
Rendered `.mmd.svg` images show in markdown; Mermaid also renders inline in some viewers.

### Viewing in IDE
- **VS Code**: Mermaid preview built-in or install Mermaid Editor extension
- **IntelliJ/GoLand**: Mermaid plugin
- **Cursor**: Built-in Mermaid support

### Regenerating SVGs
Requires [Deno](https://deno.land). Run from the `scripts/` directory (script uses `../docs/architecture`):
```bash
cd scripts && ./render-mermaid.sh
```
Writes `${file}.svg` next to each `.mmd` under `docs/architecture`.

### Live Preview
Open any `.mmd` file in your IDE or use the [Mermaid Live Editor](https://mermaid.live) to preview and edit diagrams.

## Key Architectural Decisions

1. **Microservices Architecture** - Separate services for autotesting, proxying, and MCP
2. **LLM Integration via MCP** - Model Context Protocol for LLM communication
3. **Multi-layered Security** - Nginx + application middleware for endpoint protection
4. **Docker-based Deployment** - All services containerized
5. **Configuration as Code** - Pkl for type-safe configuration
6. **Observability** - Datadog integration for monitoring and tracing

## Maintenance

This documentation should be updated when:
- New services or containers are added
- Major architectural changes occur
- External system integrations change
- Key component responsibilities shift

Last updated: 2026-01-31
