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

# Run the application locally
run: build
	@echo "Starting LiteFaaS..."
	./$(BINARY_NAME)

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
	@echo "  run           - Build and run locally"
	@echo "  dev           - Build and run in development mode"
	@echo "  test-api      - Test API endpoints (basic)"
	@echo "  test-complete - Test API endpoints (complete)"
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
