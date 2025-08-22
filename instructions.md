# LiteFaaS - Lightweight Function-as-a-Service Platform

Use the following instructions to build a lightweight, resource-efficient Function-as-a-Service platform in Go.

## Technical Foundation

### 1. Core Technologies

-   **Go**: Latest stable version (1.21+)
-   **Container Runtime**: containerd
-   **Database**: SQLite (for minimal resource usage)
-   **Web Interface**: Embedded HTML/CSS/JavaScript
-   **Container Images**: Alpine Linux (minimal footprint) with mirror.gcr.io/ prefix
-   **Supported Languages**: Python and Node.js
-   **Application Containerization**: LiteFaaS runs in a container with containerd access

### 2. Project Architecture

```
/
├── cmd/                    # Application entry points
│   └── litefaas/          # Main application binary
├── internal/              # Private application code
│   ├── api/              # HTTP API handlers
│   ├── container/        # Container management
│   ├── database/         # SQLite operations
│   ├── proxy/            # Function proxy logic
│   └── web/              # Embedded web interface
├── pkg/                   # Public packages
├── scripts/              # Installation and utility scripts
├── templates/            # Container templates
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── Dockerfile            # Container build definition
├── install.sh            # Installation script
└── README.md             # Project documentation
```

### 3. Key Technical Considerations

#### Resource Optimization

-   Minimal memory footprint (< 50MB base)
-   Efficient container lifecycle management
-   Connection pooling for database operations
-   Graceful shutdown handling
-   Resource limits per function container

#### Security Considerations

-   Container isolation using containerd
-   No authentication required (as specified)
-   Localhost-only function binding
-   Input validation and sanitization
-   Resource limits enforcement

#### Performance Requirements

-   Fast function startup (< 2 seconds)
-   Efficient proxy routing
-   Minimal overhead per function call
-   Optimized database queries
-   Connection reuse where possible

## Core Components Architecture

### 1. Main Application (cmd/litefaas)

Purpose: Entry point for the LiteFaaS application with configuration management.

Requirements:

-   Parse command-line flags and environment variables
-   Initialize database connection
-   Start containerd client
-   Launch web server and proxy
-   Handle graceful shutdown

Key Features:

-   Configuration validation
-   Health check endpoints
-   Metrics collection
-   Logging setup

### 2. Container Manager (internal/container)

Purpose: Manage function containers using containerd from within the application container.

Requirements:

-   Create and manage function containers via containerd socket
-   Handle container lifecycle (start, stop, remove)
-   Monitor container health
-   Allocate unique ports for functions
-   Resource limit enforcement
-   **Container-to-container communication**: Gestion des communications entre le conteneur principal et les conteneurs de fonctions

Key Features:

-   Container template system with mirror.gcr.io/ prefix
-   Port allocation management
-   Health monitoring
-   Resource usage tracking
-   Cleanup on function deletion
-   **Containerd client integration**: Utilisation directe du client containerd depuis le conteneur

### 3. Database Layer (internal/database)

Purpose: Manage function metadata and configuration using SQLite.

Requirements:

-   Function registration and metadata storage
-   Port allocation tracking
-   Function status management
-   Configuration persistence

Key Features:

-   Connection pooling
-   Prepared statements for performance
-   Transaction support
-   Migration system
-   Backup and restore capabilities

### 4. API Handlers (internal/api)

Purpose: Handle HTTP requests for function management.

Requirements:

-   Function CRUD operations
-   Function execution status
-   Health check endpoints
-   Metrics endpoints

Key Features:

-   RESTful API design
-   JSON request/response handling
-   Error handling and logging
-   Request validation
-   Rate limiting (optional)

### 5. Proxy Layer (internal/proxy)

Purpose: Route requests to appropriate function containers.

Requirements:

-   Route incoming requests to function containers
-   Load balancing (round-robin)
-   Health check integration
-   Request/response modification
-   Error handling and fallback

Key Features:

-   Dynamic routing based on function name
-   Connection pooling to containers
-   Request timeout handling
-   Circuit breaker pattern
-   Metrics collection

### 6. Web Interface (internal/web)

Purpose: Provide web-based management interface.

Requirements:

-   Function management dashboard
-   Real-time function status
-   Function creation and editing
-   Log viewing capabilities
-   Simple and intuitive UI

Key Features:

-   Single-page application
-   Real-time updates
-   Code editor integration
-   Function testing interface
-   Responsive design

## Container Templates

### 1. Python Function Template

```dockerfile
FROM mirror.gcr.io/library/python:3.11-alpine
WORKDIR /app
COPY function.py .
COPY requirements.txt .
RUN pip install -r requirements.txt
EXPOSE 8080
CMD ["python", "function.py"]
```

### 2. Node.js Function Template

```dockerfile
FROM mirror.gcr.io/library/node:18-alpine
WORKDIR /app
COPY function.js .
COPY package.json .
RUN npm install --production
EXPOSE 8080
CMD ["node", "function.js"]
```

## Installation Script

### Requirements

-   Idempotent installation process
-   OS detection and compatibility check
-   Automatic updates from GitHub releases
-   Dependency installation
-   Service configuration
-   **Single entry point**: Only `install.sh` script exists, no variations

### Script Features

-   Detect OS (Linux, macOS, Windows)
-   Install Go if not present
-   Install containerd if not present
-   Download latest release from GitHub
-   Configure systemd service (Linux)
-   Set up auto-update mechanism
-   Validate installation
-   **Container deployment**: Deploy LiteFaaS as a container with containerd access

## Development Workflow

### 1. Local Development

```bash
# Clone repository
git clone https://github.com/your-org/litefaas.git
cd litefaas

# Install dependencies
go mod download

# Run locally (with containerd access)
go run cmd/litefaas/main.go

# Run in container (development)
docker run --privileged -v /run/containerd/containerd.sock:/run/containerd/containerd.sock -p 8080:8080 litefaas:dev

# Run tests
go test ./...

# Build binary
go build -o bin/litefaas cmd/litefaas/main.go

# Build container
docker build -t litefaas:latest .
```

### 2. Testing Strategy

-   Unit tests for all packages
-   Integration tests for container operations
-   End-to-end tests for complete workflows
-   Performance benchmarks
-   Security testing

### 3. Code Quality

-   Go linting with golangci-lint
-   Code formatting with gofmt
-   Documentation with godoc
-   Error handling best practices
-   Minimal external dependencies

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
name: Development Workflow

on:
    push:
        branches: [main, develop]
    pull_request:
        branches: [main]

jobs:
    test:
        runs-on: ubuntu-latest
        steps:
            - uses: actions/checkout@v3
            - uses: actions/setup-go@v3
              with:
                  go-version: "1.21"
            - run: go mod download
            - run: golangci-lint run

    build:
        runs-on: ubuntu-latest
        needs: test
        steps:
            - uses: actions/checkout@v3
            - uses: actions/setup-go@v3
              with:
                  go-version: "1.21"
            - run: go build -o bin/litefaas cmd/litefaas/main.go
            - run: docker build -t litefaas .
```

## Configuration

### Environment Variables

```bash
# Application configuration
LITEFAAS_PORT=8080
LITEFAAS_DB_PATH=/var/lib/litefaas/functions.db
LITEFAAS_CONTAINERD_SOCKET=/run/containerd/containerd.sock
LITEFAAS_FUNCTION_PORT_START=9000
LITEFAAS_FUNCTION_PORT_END=9999

# Container configuration
LITEFAAS_CONTAINER_IMAGE_PREFIX=mirror.gcr.io/library/
LITEFAAS_CONTAINER_PRIVILEGED=true
LITEFAAS_CONTAINER_NETWORK_MODE=host

# Logging
LITEFAAS_LOG_LEVEL=info
LITEFAAS_LOG_FORMAT=json

# Resource limits
LITEFAAS_MAX_MEMORY_PER_FUNCTION=512MB
LITEFAAS_MAX_CPU_PER_FUNCTION=0.5
```

### Configuration File

```yaml
# config.yaml
server:
    port: 8080
    host: "0.0.0.0"

database:
    path: "/var/lib/litefaas/functions.db"
    max_connections: 10

containerd:
    socket: "/run/containerd/containerd.sock"
    namespace: "litefaas"

container:
    image_prefix: "mirror.gcr.io/library/"
    privileged: true
    network_mode: "host"

functions:
    port_range:
        start: 9000
        end: 9999
    resource_limits:
        memory: "512MB"
        cpu: "0.5"

logging:
    level: "info"
    format: "json"
```

## Best Practices

### 1. Code Organization

-   Follow Go project layout conventions
-   Use meaningful package names
-   Implement proper error handling
-   Write self-documenting code
-   Minimize external dependencies

### 2. Performance Optimization

-   Use connection pooling
-   Implement caching where appropriate
-   Optimize database queries
-   Use efficient data structures
-   Profile and benchmark critical paths

### 3. Security

-   Validate all inputs
-   Sanitize user-provided code
-   Implement resource limits
-   Use secure defaults
-   Regular security updates

### 4. Monitoring

-   Health check endpoints
-   Metrics collection
-   Structured logging
-   Error tracking
-   Performance monitoring

## Deployment

### Container Architecture

LiteFaaS est déployé dans un conteneur avec accès à containerd pour gérer les fonctions utilisateur.

#### Container Requirements

-   **Privileged container**: Accès au socket containerd (`/run/containerd/containerd.sock`)
-   **Volume mounts**: Accès aux volumes nécessaires pour containerd
-   **Network access**: Accès au réseau pour les communications inter-conteneurs
-   **Resource limits**: Limitation des ressources du conteneur principal

### Docker Deployment

```dockerfile
FROM mirror.gcr.io/golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o litefaas cmd/litefaas/main.go

FROM mirror.gcr.io/library/alpine:latest
RUN apk add --no-cache ca-certificates containerd
COPY --from=builder /app/litefaas /usr/local/bin/
EXPOSE 8080
CMD ["litefaas"]
```

### Container Runtime Configuration

```yaml
# docker-compose.yml ou équivalent
version: "3.8"
services:
    litefaas:
        image: litefaas:latest
        privileged: true
        volumes:
            - /run/containerd/containerd.sock:/run/containerd/containerd.sock
            - /var/lib/containerd:/var/lib/containerd
            - /var/lib/litefaas:/var/lib/litefaas
        ports:
            - "8080:8080"
        environment:
            - LITEFAAS_CONTAINERD_SOCKET=/run/containerd/containerd.sock
        restart: unless-stopped
```

### Systemd Service

```ini
[Unit]
Description=LiteFaaS Function Server
After=network.target containerd.service

[Service]
Type=simple
User=litefaas
Group=litefaas
ExecStart=/usr/local/bin/litefaas
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## Future Enhancements

-   Multi-language support (Rust, Go functions)
-   Function versioning
-   Advanced monitoring and metrics
-   Function scaling based on demand
-   Integration with external services
-   Plugin system for custom runtimes
