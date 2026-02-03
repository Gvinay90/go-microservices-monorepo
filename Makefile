.PHONY: help build build-docker run stop clean test gazelle deps

# Default target
help:
	@echo "Available targets:"
	@echo "  make build         - Build all services with Bazel"
	@echo "  make build-docker  - Build Docker images using Bazel"
	@echo "  make run           - Run services with Docker Compose"
	@echo "  make stop          - Stop all services"
	@echo "  make clean         - Clean Bazel cache and Docker containers"
	@echo "  make test          - Run tests (when available)"
	@echo "  make gazelle       - Update BUILD files"
	@echo "  make deps          - Update Go dependencies"

# Build with Bazel
build:
	@echo "Building with Bazel..."
	bazel build //...

# Build specific services
build-user:
	bazel build //services/user:user_service

build-expense:
	bazel build //services/expense:expense_service

build-client:
	bazel build //client:client

# Build Docker images using Bazel-built binaries
build-docker:
	@echo "Building Docker images with Bazel..."
	docker-compose build

# Run services
run: build-docker
	@echo "Starting services..."
	docker-compose up -d user-service expense-service keycloak
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Services are running!"
	@echo "  User Service: localhost:50051"
	@echo "  Expense Service: localhost:50052"
	@echo "  Keycloak: http://localhost:8080"

# Run client
run-client:
	docker-compose up client

# Stop services
stop:
	@echo "Stopping services..."
	docker-compose down

# Clean everything
clean:
	@echo "Cleaning Bazel cache..."
	bazel clean --expunge
	@echo "Cleaning Docker..."
	docker-compose down --rmi all --volumes
	@echo "Clean complete!"

# Run tests
test:
	bazel test //...

# Update BUILD files with Gazelle
gazelle:
	bazel run //:gazelle

# Update Go dependencies
deps:
	go get -u ./...
	go mod tidy
	bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%go_dependencies

# Run locally (without Docker)
run-local-user:
	bazel run //services/user:user_service

run-local-expense:
	APP_SERVER_PORT=:50052 bazel run //services/expense:expense_service

run-local-client:
	bazel run //client:client --action_env=USER_SERVICE_ADDR=localhost:50051 --action_env=EXPENSE_SERVICE_ADDR=localhost:50052

# View logs
logs:
	docker-compose logs -f

# Health check
health:
	@echo "Checking service health..."
	@grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check || echo "User service not responding"
	@grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check || echo "Expense service not responding"
