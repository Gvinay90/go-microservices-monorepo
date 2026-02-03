# Quick Start Guide

## Prerequisites

- **Bazel** 7.0+ ([Install](https://bazel.build/install))
- **Docker** & Docker Compose ([Install](https://docs.docker.com/get-docker/))
- **Go** 1.23+ (optional, for development)

## Build & Run Options

### Option 1: Docker (Recommended for Production)

```bash
# Build and run everything
./run.sh build-docker
./run.sh run

# View logs
./run.sh logs

# Stop services
./run.sh stop
```

### Option 2: Local with Bazel (Development)

```bash
# Build with Bazel
./run.sh build

# Run locally (no Docker)
./run.sh run-local

# Or run individual services
bazel run //services/user:user_service
APP_SERVER_PORT=:50052 bazel run //services/expense:expense_service
```

### Option 3: Makefile

```bash
# Build everything
make build

# Run with Docker
make run

# Run locally
make run-local-user    # Terminal 1
make run-local-expense # Terminal 2
make run-local-client  # Terminal 3
```

## Common Commands

```bash
# Build
./run.sh build              # Bazel build
./run.sh build-docker       # Docker build

# Run
./run.sh run                # Docker Compose
./run.sh run-local          # Local Bazel

# Manage
./run.sh stop               # Stop services
./run.sh clean              # Clean everything
./run.sh logs               # View logs
./run.sh health             # Health check

# Development
bazel run //:gazelle        # Update BUILD files
make deps                   # Update dependencies
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| User Service | 50051 | User management & friends |
| Expense Service | 50052 | Expense tracking & splits |
| Keycloak | 8080 | Authentication (admin/admin) |

## Testing

```bash
# With grpcurl
grpcurl -plaintext localhost:50051 user.UserService/ListUsers

# With authentication
TOKEN=$(curl -s -X POST 'http://localhost:8080/realms/expense-sharing/protocol/openid-connect/token' \
  -d 'client_id=client-app' \
  -d 'username=alice' \
  -d 'password=password' \
  -d 'grant_type=password' | jq -r '.access_token')

grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  localhost:50051 user.UserService/ListUsers
```

## Configuration

Edit `config/base.yaml`:

```yaml
server:
  port: ":50051"
  rate_limit: 100

auth:
  enabled: false  # Set to true to enable Keycloak auth
  
database:
  driver: "sqlite"  # or "postgres"
  dsn: "dev.db"
```

## Troubleshooting

**Build fails:**
```bash
bazel clean --expunge
./run.sh build
```

**Services won't start:**
```bash
./run.sh stop
./run.sh clean
./run.sh run
```

**Port already in use:**
```bash
lsof -ti:50051 | xargs kill -9
lsof -ti:50052 | xargs kill -9
```

## Documentation

- [BAZEL_GUIDE.md](BAZEL_GUIDE.md) - Bazel build system guide
- [CLEAN_ARCHITECTURE.md](CLEAN_ARCHITECTURE.md) - Architecture patterns
- [KEYCLOAK_AUTH.md](KEYCLOAK_AUTH.md) - Authentication setup
- [PRODUCTION.md](PRODUCTION.md) - Production deployment

## Next Steps

1. **Enable Authentication**: See [KEYCLOAK_AUTH.md](KEYCLOAK_AUTH.md)
2. **Deploy to Production**: See [PRODUCTION.md](PRODUCTION.md)
3. **Add Tests**: Create `*_test.go` files and run `bazel test //...`
