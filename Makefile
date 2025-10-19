.PHONY: help dev test lint build clean docker-up docker-down docker-logs

# Default target
help:
	@echo "Available targets:"
	@echo "  make dev            - Start development environment (Docker services + backend)"
	@echo "  make test           - Run all tests"
	@echo "  make test-unit      - Run unit tests only"
	@echo "  make test-integration - Run integration tests"
	@echo "  make lint           - Run linters"
	@echo "  make build          - Build backend binary"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-up      - Start Docker services"
	@echo "  make docker-down    - Stop Docker services"
	@echo "  make docker-logs    - Show Docker logs"
	@echo "  make run            - Run backend server"
	@echo "  make fmt            - Format code"
	@echo "  make docs           - Generate OpenAPI documentation"

# Start development environment
dev: docker-up
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Starting backend server..."
	cd backend && CONFIG_PATH=../config/local.yaml go run main.go

# Run all tests
test:
	@echo "Running all tests..."
	cd backend && go test -v ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	cd backend && go test -v -short ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	cd backend && go test -v -run Integration ./...

# Run linters
lint:
	@echo "Running linters..."
	cd backend && go vet ./...
	cd backend && go fmt ./...
	@echo "Checking for formatting issues..."
	@cd backend && test -z "$$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)

# Format code
fmt:
	@echo "Formatting code..."
	cd backend && go fmt ./...

# Build backend binary
build:
	@echo "Building backend..."
	cd backend && go build -o bin/server main.go
	@echo "Binary created at backend/bin/server"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf backend/bin
	cd backend && go clean

# Start Docker services
docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d mongodb keycloak

# Start Docker services with observability stack
docker-up-full:
	@echo "Starting all Docker services including observability..."
	docker-compose --profile observability up -d

# Stop Docker services
docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

# Show Docker logs
docker-logs:
	docker-compose logs -f

# Run backend server
run:
	@echo "Running backend server..."
	cd backend && CONFIG_PATH=../config/local.yaml go run main.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	cd backend && go mod download
	cd backend && go mod tidy

# Run database migrations (placeholder for future)
migrate:
	@echo "No migrations implemented yet"

# Seed database with sample data (placeholder for future)
seed:
	@echo "No seed data implemented yet"

# Generate code (placeholder for future)
generate:
	@echo "No code generation configured yet"

# Generate API documentation
docs:
	@echo "Generating OpenAPI documentation..."
	cd backend && swag init --output docs --parseDependency --parseInternal
	@echo "Documentation generated at backend/docs/"
	@echo "View at http://localhost:8080/docs/scalar (Scalar UI)"
	@echo "View at http://localhost:8080/swagger/index.html (Swagger UI)"
