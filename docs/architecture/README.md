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

Diagrams are created using Mermaid, which has native support in GitLab, GitHub, and many IDEs.

### Viewing in GitLab/GitHub
Mermaid diagrams render automatically when viewing markdown files in GitLab and GitHub.

### Viewing in IDE
- **VS Code**: Mermaid preview built-in or install Mermaid Editor extension
- **IntelliJ/GoLand**: Mermaid plugin
- **Cursor**: Built-in Mermaid support

### Generating Images (Optional)
```bash
# Install Mermaid CLI
npm install -g @mermaid-js/mermaid-cli

# Generate PNG next to each .mmd (run from repo root)
find docs/architecture -name '*.mmd' -exec sh -c 'mmdc -i "$1" -o "${1%.mmd}.png"' _ {} \;

# Generate SVG into output/
mkdir -p docs/architecture/output
find docs/architecture -name '*.mmd' -exec sh -c 'mmdc -i "$1" -o "docs/architecture/output/$(basename "$1" .mmd).svg" -t svg' _ {} \;
```

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

Last updated: 2026-01-20
