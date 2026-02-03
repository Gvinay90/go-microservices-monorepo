# Expense Sharing gRPC Application

A production-ready expense-sharing application built with Go, gRPC, Bazel, and Keycloak authentication, demonstrating Clean Architecture and enterprise security practices.

## 🚀 Quick Start (First Time Setup)

### Prerequisites

Install these tools before starting:

1. **Bazel** (7.0+): [Installation Guide](https://bazel.build/install)
2. **Docker & Docker Compose**: [Installation Guide](https://docs.docker.com/get-docker/)
3. **grpcurl** (for testing): `brew install grpcurl` (Mac) or [Other platforms](https://github.com/fullstorydev/grpcurl)

### Step 1: Clone and Build

```bash
# Navigate to project directory
cd go-grpc-bazel

# Build everything with Bazel
./run.sh build
```

**Expected output:**
```
Building with Bazel...
INFO: Build completed successfully, 26 total actions
✓ Build complete!
```

### Step 2: Start Services

**Option A: Local (Fast for Development)**
```bash
# Build services
./run.sh build

# Run in separate terminals
# Terminal 1:
bazel run //services/user:user_service

# Terminal 2:
APP_SERVER_PORT=:50052 bazel run //services/expense:expense_service
```

**Option B: Docker (Production-like Environment)**
```bash
# Build Docker images (Bazel builds inside Linux container)
# First build takes ~5 minutes, subsequent builds are cached
./run.sh build-docker

# Run services
./run.sh run
```

**Note:** Docker build compiles inside a Linux container, avoiding cross-compilation issues. First build is slow but subsequent builds use Docker layer caching.

**Services will be available at:**
- **Web Frontend**: `http://localhost:3000` (React app with TailwindCSS)
- **API Gateway**: `http://localhost:8081` (REST API for web app)
- **User Service (gRPC)**: `localhost:50051`
- **Expense Service (gRPC)**: `localhost:50052`
- **Keycloak**: `http://localhost:8080` (admin/admin)
- **MailHog UI**: `http://localhost:8025` (email testing)
- **PostgreSQL**: `localhost:5433` (DB for app services)

### Step 3: Test the Services

#### Test 1: Health Check

```bash
# Check if services are running
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check
```

**Expected output:**
```json
{
  "status": "SERVING"
}
```

#### Test 2: Register a User

```bash
grpcurl -plaintext \
  -d '{"name": "Alice Smith", "email": "alice@example.com"}' \
  localhost:50051 user.UserService/RegisterUser
```

**Expected output:**
```json
{
  "user": {
    "id": "user_...",
    "name": "Alice Smith",
    "email": "alice@example.com"
  }
}
```

#### Test 3: Create an Expense

```bash
grpcurl -plaintext \
  -d '{
    "description": "Team Dinner",
    "total_amount": 100.0,
    "paid_by": "alice@example.com",
    "splits": [
      {"user_id": "alice@example.com", "amount": 50.0},
      {"user_id": "bob@example.com", "amount": 50.0}
    ]
  }' \
  localhost:50052 expense.ExpenseService/CreateExpense
```

**Expected output:**
```json
{
  "expense": {
    "id": "exp_...",
    "description": "Team Dinner",
    "totalAmount": 100,
    "paidBy": "alice@example.com",
    "splits": [...]
  }
}
```

#### Test 4: List Users

```bash
grpcurl -plaintext localhost:50051 user.UserService/ListUsers
```

#### Test 5: List Expenses

```bash
grpcurl -plaintext localhost:50052 expense.ExpenseService/ListExpenses
```

### Step 4: Test with Authentication (Optional)

#### Enable Authentication

Edit `config/base.yaml`:
```yaml
auth:
  enabled: true  # Change from false to true
```

Restart services:
```bash
./run.sh stop
./run.sh run
```

#### Get Authentication Token

```bash
# Login as alice
TOKEN=$(curl -s -X POST 'http://localhost:8080/realms/expense-sharing/protocol/openid-connect/token' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'client_id=client-app' \
  -d 'username=alice' \
  -d 'password=password' \
  -d 'grant_type=password' | jq -r '.access_token')

echo "Token: $TOKEN"
```

#### Make Authenticated Request

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"name": "Bob Jones", "email": "bob@example.com"}' \
  localhost:50051 user.UserService/RegisterUser
```

**Without token (should fail):**
```bash
grpcurl -plaintext \
  -d '{"name": "Charlie", "email": "charlie@example.com"}' \
  localhost:50051 user.UserService/RegisterUser
```

**Expected error:**
```
Code: Unauthenticated
Message: missing authorization header
```

## 📋 Available Commands

### Build Commands
```bash
./run.sh build          # Build with Bazel
./run.sh build-docker   # Build Docker images
make build              # Alternative using Makefile
```

### Run Commands
```bash
./run.sh run            # Run with Docker Compose
./run.sh run-local      # Run locally without Docker
./run.sh stop           # Stop all services
```

### Utility Commands
```bash
./run.sh logs           # View service logs
./run.sh health         # Check service health
./run.sh clean          # Clean everything
```

### View Individual Service Logs
```bash
# View logs for specific service
docker logs go-grpc-bazel-user-service-1        # User service logs
docker logs go-grpc-bazel-expense-service-1     # Expense service logs
docker logs go-grpc-bazel-app-db-1              # PostgreSQL logs
docker logs go-grpc-bazel-keycloak-1            # Keycloak logs

# Follow logs in real-time (like tail -f)
docker logs -f go-grpc-bazel-user-service-1

# View last 50 lines
docker logs --tail 50 go-grpc-bazel-user-service-1

# View logs with timestamps
docker logs -t go-grpc-bazel-user-service-1

# View logs from last 5 minutes
docker logs --since 5m go-grpc-bazel-user-service-1
```

## 🧪 Complete Testing Workflow

Here's a complete workflow to test all features:

```bash
# 1. Start services
./run.sh run

# 2. Register users
grpcurl -plaintext -d '{"name": "Alice", "email": "alice@example.com"}' \
  localhost:50051 user.UserService/RegisterUser

grpcurl -plaintext -d '{"name": "Bob", "email": "bob@example.com"}' \
  localhost:50051 user.UserService/RegisterUser

# 3. Add friendship
grpcurl -plaintext -d '{"user_id": "alice@example.com", "friend_id": "bob@example.com"}' \
  localhost:50051 user.UserService/AddFriend

# 4. Create expense
grpcurl -plaintext -d '{
  "description": "Lunch",
  "total_amount": 50.0,
  "paid_by": "alice@example.com",
  "splits": [
    {"user_id": "alice@example.com", "amount": 25.0},
    {"user_id": "bob@example.com", "amount": 25.0}
  ]
}' localhost:50052 expense.ExpenseService/CreateExpense

# 5. Get user's friends
grpcurl -plaintext -d '{"user_id": "alice@example.com"}' \
  localhost:50051 user.UserService/GetFriends

# 6. List all expenses
grpcurl -plaintext localhost:50052 expense.ExpenseService/ListExpenses

# 7. Stop services
./run.sh stop
```

## 🔐 Test Users (When Auth is Enabled)

| Username | Password | Email             | Roles       |
| -------- | -------- | ----------------- | ----------- |
| alice    | password | alice@example.com | user        |
| bob      | password | bob@example.com   | user        |
| admin    | admin123 | admin@example.com | user, admin |

## 🐛 Troubleshooting

### Services won't start

```bash
# Check if ports are in use
lsof -i :50051
lsof -i :50052

# Kill processes if needed
lsof -ti:50051 | xargs kill -9
lsof -ti:50052 | xargs kill -9

# Clean and rebuild
./run.sh clean
./run.sh build
./run.sh run
```

### Build fails

```bash
# Clean Bazel cache
bazel clean --expunge

# Rebuild
./run.sh build
```

### Docker build fails

```bash
# Remove old containers and images
docker-compose down --rmi all --volumes

# Rebuild
./run.sh build-docker
```

### Can't connect to services

```bash
# Check if services are running
docker ps

# Check logs
./run.sh logs

# Or for specific service
docker logs go-grpc-bazel-user-service-1
```

## 📚 Documentation

- **[QUICKSTART.md](QUICKSTART.md)** - Quick reference guide
- **[BAZEL_GUIDE.md](BAZEL_GUIDE.md)** - Bazel build system details
- **[CLEAN_ARCHITECTURE.md](CLEAN_ARCHITECTURE.md)** - Architecture explanation
- **[KEYCLOAK_AUTH.md](KEYCLOAK_AUTH.md)** - Authentication setup
- **[PRODUCTION.md](PRODUCTION.md)** - Production deployment guide

## 🏗️ Architecture

This is a **full-stack mono-repo** with microservices backend, API gateway, and React frontend:

```
┌─────────────────────────────────────────────────────────────────┐
│                         MONO-REPO                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  web/                          # React Frontend (Vite + React)   │
│  ├── src/                                                        │
│  │   ├── components/          # Reusable UI components          │
│  │   ├── pages/               # Route pages                     │
│  │   ├── services/            # API client (axios)              │
│  │   └── store/               # State management (zustand)      │
│  ├── package.json             # Node dependencies               │
│  └── Dockerfile               # Production build with nginx     │
│                                                                   │
│  api-gateway/                  # HTTP/REST → gRPC Gateway        │
│  ├── main.go                  # HTTP server (port 8080)         │
│  ├── handlers/                # REST endpoints                  │
│  ├── clients/                 # gRPC clients                    │
│  └── middleware/              # Auth, CORS, logging             │
│                                                                   │
│  services/                     # gRPC Microservices              │
│  ├── user/                    # User service (port 50051)       │
│  │   └── main.go                                                │
│  └── expense/                 # Expense service (port 50052)    │
│      └── main.go                                                │
│                                                                   │
│  internal/                     # Service implementations         │
│  ├── user/                                                       │
│  │   ├── domain/              # Business models                 │
│  │   ├── repository/          # Data access layer               │
│  │   ├── service/             # Business logic                  │
│  │   └── handler/             # gRPC handlers                   │
│  └── expense/                                                    │
│      └── (same structure)                                        │
│                                                                   │
│  proto/                        # Protocol Buffers                │
│  ├── user/                    # User service definitions        │
│  │   ├── user.proto                                             │
│  │   └── BUILD.bazel          # Code generation rules           │
│  └── expense/                 # Expense service definitions     │
│      ├── expense.proto                                           │
│      └── BUILD.bazel                                             │
│                                                                   │
│  pkg/                          # Shared libraries                │
│  ├── auth/                    # JWT validation, middleware      │
│  ├── config/                  # Configuration parsing           │
│  ├── logger/                  # Structured logging              │
│  ├── server/                  # gRPC/HTTP server setup          │
│  └── keycloak/                # Keycloak integration            │
│                                                                   │
│  Infrastructure & Config                                         │
│  ├── docker-compose.yml       # Multi-service orchestration     │
│  ├── keycloak/                # Auth server config              │
│  ├── config/                  # Service configurations          │
│  └── scripts/                 # Deployment & util scripts       │
│                                                                   │
│  Build System                                                    │
│  ├── BUILD.bazel              # Root build file                 │
│  ├── MODULE.bazel             # Dependency management           │
│  ├── WORKSPACE                # Bazel workspace                 │
│  ├── deps.bzl                 # Go dependencies                 │
│  └── Makefile                 # Convenience commands            │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘

```

### Service Communication Flow

```
┌─────────┐      HTTP/REST      ┌──────────────┐      gRPC
│   Web   │ ───────────────────▶ │ API Gateway  │ ────────────┐
│ (React) │ ◀─────────────────── │   (Go)       │             │
└─────────┘       JSON           └──────────────┘             │
                                         │                     │
                                         │                     │
                                    ┌────▼────┐          ┌────▼────┐
                                    │  User   │          │ Expense │
                                    │ Service │          │ Service │
                                    │ :50051  │          │ :50052  │
                                    └────┬────┘          └────┬────┘
                                         │                     │
                                         └─────────┬───────────┘
                                                   │
                                              ┌────▼────┐
                                              │PostgreSQL│
                                              │  (App)   │
                                              └──────────┘

Additional Services:
├── Keycloak (port 8080)    # Authentication & user management
├── Keycloak DB (Postgres)  # Keycloak data store
└── MailHog (port 8025)     # Email testing (dev only)
```

### Technology Stack

| Layer           | Technology                  | Purpose                        |
| --------------- | --------------------------- | ------------------------------ |
| **Frontend**    | React 19, Vite, TailwindCSS | Modern SPA with fast build     |
| **API Gateway** | Go, Gin, gRPC-Gateway       | REST → gRPC translation        |
| **Services**    | Go, gRPC, Protocol Buffers  | High-performance microservices |
| **Build**       | Bazel, Docker               | Hermetic, reproducible builds  |
| **Auth**        | Keycloak, JWT               | OpenID Connect authentication  |
| **Database**    | PostgreSQL 15               | Relational data store          |
| **State**       | Zustand                     | Lightweight React state        |

## ✨ Features

- ✅ **Clean Architecture** - Separated layers (Domain, Repository, Service, Handler)
- ✅ **Keycloak Authentication** - JWT validation with RBAC
- ✅ **Rate Limiting** - 100 requests/minute per user
- ✅ **Audit Logging** - Security event tracking
- ✅ **TLS/HTTPS** - Encrypted communication
- ✅ **Bazel Builds** - Hermetic, reproducible builds
- ✅ **Docker Support** - Containerized deployment
- ✅ **Health Checks** - Kubernetes-ready
- ✅ **Graceful Shutdown** - Zero-downtime deploys

## 🎯 Next Steps

1. **Explore the code**: Check out `internal/user/` and `internal/expense/`
2. **Add tests**: Create `*_test.go` files and run `bazel test //...`
3. **Enable authentication**: Follow [KEYCLOAK_AUTH.md](KEYCLOAK_AUTH.md)
4. **Deploy to production**: Follow [PRODUCTION.md](PRODUCTION.md)

## 🤝 Contributing

This is a learning project demonstrating:
- Go microservices with gRPC
- Clean Architecture patterns
- Bazel build system
- Enterprise security (Keycloak, TLS, rate limiting)
- Production-ready practices

## 📝 License

MIT License - Feel free to use this as a learning resource or template for your projects.

---

**Need help?** Check the troubleshooting section above or review the documentation files.
