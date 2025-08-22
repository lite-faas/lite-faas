.PHONY: build clean test run docker-build docker-run install help

# Variables
BINARY_NAME=litefaas
MAIN_PATH=cmd/litefaas/main.go
DOCKER_IMAGE=litefaas:latest

# Build the application
build:
	@echo "Building LiteFaaS..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f *.db
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test ./...
	@echo "Tests complete"

# Run the application in container (production mode)
run: docker-build
	@echo "Starting LiteFaaS in production container..."
	@echo "Building and running LiteFaaS with containerd access..."
	docker run --rm -d \
		--name litefaas-prod \
		--privileged \
		-p 8080:8080 \
		-v /run/containerd/containerd.sock:/run/containerd/containerd.sock \
		-v /var/lib/containerd:/var/lib/containerd \
		-v litefaas-data:/var/lib/litefaas \
		-e LITEFAAS_CONTAINERD_SOCKET=/run/containerd/containerd.sock \
		-e LITEFAAS_DB_PATH=/var/lib/litefaas/functions.db \
		-e LITEFAAS_PORT=8080 \
		-e LITEFAAS_LOG_LEVEL=info \
		-e LITEFAAS_LOG_FORMAT=json \
		$(DOCKER_IMAGE)
	@echo "LiteFaaS is running in production mode on http://localhost:8080"
	@echo "Container name: litefaas-prod"
	@echo "Use 'make stop' to stop the container"

# Run in development mode
dev: build
	@echo "Starting LiteFaaS in development mode..."
	./$(BINARY_NAME) --dev --db-path ./dev.db --port 8080 --log-format text --log-level debug

# Test API endpoints
test-api:
	@echo "Testing API endpoints..."
	@chmod +x test_diagnostic.sh
	./test_diagnostic.sh

# Test complet de l'API
test-complete:
	@echo "Running complete API tests..."
	@chmod +x test_complete.sh
	./test_complete.sh

# Test de l'interface web
test-web:
	@echo "Testing web interface..."
	@chmod +x test_web.sh
	./test_web.sh

# Test complet (API + Web)
test-all:
	@echo "Running all tests..."
	@chmod +x test_all.sh
	./test_all.sh

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker build complete"

# Run with Docker Compose
docker-run:
	@echo "Starting LiteFaaS with Docker Compose..."
	docker-compose up -d
	@echo "LiteFaaS is running on http://localhost:8080"

# Stop production container
stop:
	@echo "Stopping LiteFaaS production container..."
	docker stop litefaas-prod 2>/dev/null || echo "Container not running"
	docker rm litefaas-prod 2>/dev/null || echo "Container not found"
	@echo "LiteFaaS production container stopped"

# Show production container logs
logs:
	@echo "Showing LiteFaaS production container logs..."
	docker logs -f litefaas-prod

# Stop Docker Compose
docker-stop:
	@echo "Stopping LiteFaaS..."
	docker-compose down
	@echo "LiteFaaS stopped"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run
	@echo "Linting complete"

# Install the application
install:
	@echo "Installing LiteFaaS..."
	chmod +x install.sh
	./install.sh
	@echo "Installation complete"

# Show help
help:
	@echo "LiteFaaS Makefile Commands:"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  run           - Build and run in production container"
	@echo "  dev           - Build and run in development mode"
	@echo "  stop          - Stop production container"
	@echo "  logs          - Show production container logs"
	@echo "  test-api      - Test API endpoints (basic)"
	@echo "  test-complete - Test API endpoints (complete)"
	@echo "  test-web      - Test web interface"
	@echo "  test-all      - Test complet (API + Web)"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-stop   - Stop Docker Compose"
	@echo "  deps          - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  install       - Install the application"
	@echo "  help          - Show this help"

# Default target
.DEFAULT_GOAL := help
