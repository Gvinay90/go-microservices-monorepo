# Go Microservices Mono-Repo

A **mono-repo** built to demonstrate designing and running a **microservice system with a single command**. It includes Go gRPC services (user, expense), a REST API gateway, React frontend, Keycloak authentication, and PostgreSQL—all buildable and testable via one script.

## 🎯 Goal

- **Single-command workflow** – Build with `./run.sh build` (Bazel) or `./run.sh build-docker` (Docker); run with `./run.sh run` (Docker) or `./run.sh run-local` (Bazel).
- **Organized microservices** – Clear separation of services (user, expense, API gateway, web) with shared libraries and protocol definitions.
- **Production-like setup** – Docker Compose, Keycloak auth, rate limiting, audit logging, and health checks.

**Highlights:** Keycloak is auto-configured on startup (service-account permissions and JWT roles scope) so registration and login work without manual realm setup. Users who log in via Keycloak are synced into the app database with a stable ID so friends and expenses work correctly.

## 🚀 Quick Start (First Time Setup)

### Prerequisites

- **Docker & Docker Compose**: [Installation Guide](https://docs.docker.com/get-docker/) — for the Docker workflow.
- **Bazel** (7.0+): [Installation Guide](https://bazel.build/install) — for the Bazel / local workflow.
- **grpcurl** (for gRPC testing): `brew install grpcurl` (Mac) or [Other platforms](https://github.com/fullstorydev/grpcurl)

### Step 1: Clone and Build

**Option A: Docker (full stack in containers)**

```bash
cd go-microservices-monorepo   # or your clone path

# Build all Docker images (Go builds inside container; first run may take a few minutes)
./run.sh build-docker
```

**Option B: Bazel (local Go build)**

```bash
cd go-microservices-monorepo

# Build proto, services, internal packages, and shared pkg
./run.sh build
```

Expected: `Building with Bazel...` → `✓ Build complete!`

### Step 2: Start Services

**Option A: Docker (recommended – full stack)**

```bash
./run.sh run
```

This starts all services (PostgreSQL, Keycloak, user service, expense service, API gateway, web UI, MailHog). **Keycloak is auto-configured** on first run (service-account roles and JWT scope mappers) so registration and login work out of the box.

**Option B: Local with Bazel (gRPC services only)**

Run in separate terminals (you need PostgreSQL and, for auth, Keycloak running separately, e.g. via Docker):

```bash
# Terminal 1 – User service
bazel run //services/user:user_service

# Terminal 2 – Expense service
APP_SERVER_PORT=:50052 bazel run //services/expense:expense_service
```

Or use the script:

```bash
./run.sh run-local
```

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
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check
```

Expected: `{"status": "SERVING"}`

#### Test 2: Register a User (REST API – recommended)

```bash
curl -s -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Smith", "email": "alice@example.com", "password": "password123"}'
```

Expected: JSON with `user` (id, name, email) and `"User registered successfully. You can now login."`

You can also **register and use the app in the browser** at http://localhost:3000 (Sign up → Login → Dashboard).

#### Test 2b: Register via gRPC (when auth is disabled)

```bash
grpcurl -plaintext \
  -d '{"name": "Bob", "email": "bob@example.com", "password": "password123"}' \
  localhost:50051 user.UserService/RegisterUser
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

### Step 4: Test with Authentication

Authentication is **enabled by default** (`config/base.yaml`). After `./run.sh run`, you can register and log in via the web app or the REST API. Pre-configured test users (see table below) are available for token-based testing.

#### Get Authentication Token

```bash
# Login as alice (use client_id=expense-app; pre-configured in Keycloak)
TOKEN=$(curl -s -X POST 'http://localhost:8080/realms/expense-sharing/protocol/openid-connect/token' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'client_id=expense-app' \
  -d 'client_secret=1aLlWaA8A37Qa6HyyUGdbx9tnGZRmXAH' \
  -d 'username=alice' \
  -d 'password=password' \
  -d 'grant_type=password' | jq -r '.access_token')

echo "Token: $TOKEN"
```

Or log in via the REST API and use the returned `access_token`:

```bash
curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "password"}' | jq -r '.access_token'
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

All orchestrated via `./run.sh`:

**Build**

| Command | Description |
|--------|--------------|
| `./run.sh build` | Build with **Bazel** (proto, services, internal, pkg) |
| `./run.sh build-docker` | Build **Docker** images (full stack) |

**Run**

| Command | Description |
|--------|--------------|
| `./run.sh run` | Start full stack with **Docker Compose** |
| `./run.sh run-local` | Run user & expense services **locally** with Bazel (no Docker) |
| `./run.sh stop` | Stop all Docker services |
| `./run.sh watch` | Run with hot-reload (Docker Compose Watch) |

**Utilities**

| Command | Description |
|--------|--------------|
| `./run.sh logs` | Stream logs from all services |
| `./run.sh health` | Check gRPC service health (ports 50051, 50052) |
| `./run.sh clean` | Remove containers, images, volumes, and Bazel cache |

**Alternative:** `make build` for Bazel build (see [Makefile](Makefile)).

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

This repo is a **full-stack mono-repo**: microservices backend, API gateway, and React frontend, all buildable and runnable with a single command.

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
├── Keycloak (port 8080)       # Authentication & user management
├── Keycloak init (one-shot)   # Assigns service-account roles & JWT scope mappers on first run
├── Keycloak DB (Postgres)     # Keycloak data store
└── MailHog (port 8025)        # Email testing (dev only)
```

### Technology Stack

| Layer           | Technology                  | Purpose                        |
| --------------- | --------------------------- | ------------------------------ |
| **Frontend**    | React 19, Vite, TailwindCSS | Modern SPA with fast build     |
| **API Gateway** | Go, Gin                     | REST → gRPC translation        |
| **Services**    | Go, gRPC, Protocol Buffers  | High-performance microservices |
| **Build / Run** | Bazel, Docker, Docker Compose | Local or containerized build and run |
| **Auth**        | Keycloak, JWT               | OpenID Connect authentication  |
| **Database**    | PostgreSQL 15               | Relational data store          |
| **State**       | Zustand                     | Lightweight React state        |

## ✨ Features

- ✅ **Single-command workflow** – Bazel (`./run.sh build`, `./run.sh run-local`) or Docker (`./run.sh build-docker`, `./run.sh run`) for the full stack
- ✅ **Clean Architecture** – Domain, Repository, Service, Handler layers per service
- ✅ **Keycloak Authentication** – JWT validation, RBAC; auto-configured service account and scopes
- ✅ **User sync on login** – Keycloak users are created in the app DB with stable IDs (friends, expenses work)
- ✅ **Rate limiting** – 100 requests/minute per user at the API gateway
- ✅ **Audit logging** – Security and request logging
- ✅ **Docker & Bazel** – Containerized full stack (Docker) or local gRPC services (Bazel)
- ✅ **Health checks** – gRPC health for services
- ✅ **Graceful shutdown** – Clean exit on SIGTERM

## 🎯 Next Steps

1. **Explore the code** – `internal/user/` and `internal/expense/` for service structure; `api-gateway/handlers/` for REST → gRPC.
2. **Add tests** – `*_test.go` and run with `go test ./...` or `bazel test //...`.
3. **Auth and Keycloak** – See [KEYCLOAK_AUTH.md](KEYCLOAK_AUTH.md) and [KEYCLOAK_SETUP.md](KEYCLOAK_SETUP.md).
4. **Deploy** – [PRODUCTION.md](PRODUCTION.md) for deployment and hardening.

## 🤝 About This Repo

This project demonstrates:

- A **mono-repo** with a **single entrypoint** (`./run.sh`) to build and run the system
- **Organized microservices** (user, expense, API gateway, web) with shared proto and packages
- **Go gRPC** backends and a REST API gateway for the frontend
- **Keycloak** for auth, with auto-setup so registration and login work out of the box
- **Clean Architecture** and production-oriented practices (rate limiting, audit, health checks)

## 📝 License

MIT License - Feel free to use this as a learning resource or template for your projects.

---

**Need help?** Check the troubleshooting section above or review the documentation files.
