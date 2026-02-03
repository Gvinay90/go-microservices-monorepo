# Developer Guide: Bazel + Docker Architecture

A complete guide to understanding how Bazel and Docker work together in this project, from local development to production deployment.

---

## 🏗️ Architecture Overview

### The Big Picture

```
┌─────────────────────────────────────────────────────────────┐
│                    YOUR MAC (Development)                    │
│                                                              │
│  ┌────────────┐         ┌──────────────────────────────┐   │
│  │   Bazel    │────────▶│  Native Mac Binaries         │   │
│  │  (Local)   │         │  • user_service (Mac)        │   │
│  └────────────┘         │  • expense_service (Mac)     │   │
│                         └──────────────────────────────┘   │
│                                                              │
│  ┌────────────┐         ┌──────────────────────────────┐   │
│  │   Docker   │────────▶│  Linux Containers            │   │
│  │  (Local)   │         │  • user-service (Linux)      │   │
│  └────────────┘         │  • expense-service (Linux)   │   │
│                         │  • app-db (PostgreSQL)       │   │
│                         │  • keycloak                  │   │
│                         └──────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 📦 Build Systems Explained

### Option 1: Bazel (Local Development - Fast!)

**What is Bazel?**
Think of Bazel as a super-smart build system that:
- Remembers what you've already built
- Only rebuilds what changed
- Ensures everyone gets the same results

```
┌──────────────────────────────────────────────────────────┐
│                    BAZEL BUILD FLOW                       │
└──────────────────────────────────────────────────────────┘

Step 1: You run: bazel build //services/user:user_service
   │
   ├─▶ Bazel checks: "Did anything change?"
   │   • Source code? ✓
   │   • Dependencies? ✓
   │   • Build rules? ✓
   │
   ├─▶ If nothing changed: Use cached binary (instant!)
   │
   └─▶ If something changed: Build only what's needed
       │
       ├─▶ Download dependencies (if needed)
       ├─▶ Compile changed files
       ├─▶ Link binary
       └─▶ Cache result for next time

Result: bazel-bin/services/user/user_service_/user_service
        (Mac binary - runs directly on your Mac)
```

**When to use Bazel:**
- ✅ Daily development
- ✅ Quick iterations
- ✅ Testing changes locally
- ✅ Fast feedback loop (2-5 seconds for small changes)

### Option 2: Docker (Production-like Environment)

**What is Docker?**
Docker creates isolated "containers" - like mini Linux computers running on your Mac.

```
┌──────────────────────────────────────────────────────────┐
│                   DOCKER BUILD FLOW                       │
└──────────────────────────────────────────────────────────┘

Step 1: You run: ./run.sh build-docker
   │
   ├─▶ Docker reads Dockerfile
   │   
   ├─▶ Stage 1: Builder Container (Linux)
   │   │
   │   ├─▶ Pull golang:1.23-alpine image
   │   ├─▶ Copy your source code
   │   ├─▶ Run: go build (inside Linux!)
   │   └─▶ Creates Linux binaries
   │
   └─▶ Stage 2: Runtime Container (Linux)
       │
       ├─▶ Pull alpine:latest (tiny Linux)
       ├─▶ Copy binary from builder
       ├─▶ Copy config files
       └─▶ Create final image (50MB)

Result: Docker image ready to run anywhere
```

**When to use Docker:**
- ✅ Testing production-like setup
- ✅ Running all services together
- ✅ Testing with PostgreSQL
- ✅ Before deploying to production

---

## 🔄 Development Workflows

### Workflow 1: Fast Local Development (Recommended)

```
┌─────────────────────────────────────────────────────────┐
│         FAST ITERATION CYCLE (Bazel)                    │
└─────────────────────────────────────────────────────────┘

1. Edit Code
   ↓
   internal/user/service/user.go
   (Make your changes)
   
2. Build (2-5 seconds)
   ↓
   bazel build //services/user:user_service
   
3. Run
   ↓
   bazel run //services/user:user_service
   
4. Test
   ↓
   grpcurl -plaintext localhost:50051 user.UserService/ListUsers
   
5. See results immediately!
   ↓
   Repeat from step 1

Total time per iteration: ~10 seconds
```

**Terminal Setup:**
```bash
# Terminal 1: User Service
bazel run //services/user:user_service

# Terminal 2: Expense Service  
APP_SERVER_PORT=:50052 bazel run //services/expense:expense_service

# Terminal 3: PostgreSQL (if needed)
docker run -p 5433:5432 -e POSTGRES_PASSWORD=splitwise postgres:15-alpine

# Terminal 4: Your testing commands
grpcurl -plaintext localhost:50051 ...
```

### Workflow 2: Full Stack Testing (Docker)

```
┌─────────────────────────────────────────────────────────┐
│      FULL STACK TESTING (Docker Compose)                │
└─────────────────────────────────────────────────────────┘

1. Make Changes
   ↓
   Edit any service code
   
2. Rebuild Docker Images (2-3 min first time, 30s cached)
   ↓
   ./run.sh build-docker
   
3. Start Everything
   ↓
   ./run.sh run
   
   Starts:
   • user-service
   • expense-service
   • app-db (PostgreSQL)
   • keycloak
   
4. Test Full System
   ↓
   All services talk to each other
   
5. View Logs
   ↓
   ./run.sh logs

Total time per iteration: ~3-5 minutes
```

---

## 🌐 Network Architecture

### Local Development (Bazel)

```
┌──────────────────────────────────────────────────────────┐
│                  YOUR MAC (localhost)                     │
│                                                           │
│  ┌─────────────────┐         ┌──────────────────┐       │
│  │  User Service   │         │ Expense Service  │       │
│  │  :50051         │         │  :50052          │       │
│  └────────┬────────┘         └────────┬─────────┘       │
│           │                           │                  │
│           └───────────┬───────────────┘                  │
│                       │                                  │
│                       ▼                                  │
│              ┌─────────────────┐                         │
│              │   PostgreSQL    │                         │
│              │   :5433         │                         │
│              └─────────────────┘                         │
│                                                           │
│  Access from your Mac:                                   │
│  • localhost:50051  → User Service                       │
│  • localhost:50052  → Expense Service                    │
│  • localhost:5433   → PostgreSQL                         │
└──────────────────────────────────────────────────────────┘
```

**How to connect:**
```bash
# From your Mac terminal
grpcurl -plaintext localhost:50051 user.UserService/ListUsers

# From your code
conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
```

### Docker Development

```
┌──────────────────────────────────────────────────────────┐
│              DOCKER NETWORK (splitwise-net)               │
│                                                           │
│  ┌─────────────────┐         ┌──────────────────┐       │
│  │  user-service   │────────▶│ expense-service  │       │
│  │  (container)    │         │  (container)     │       │
│  └────────┬────────┘         └────────┬─────────┘       │
│           │                           │                  │
│           └───────────┬───────────────┘                  │
│                       │                                  │
│                       ▼                                  │
│              ┌─────────────────┐                         │
│              │     app-db      │                         │
│              │   (container)   │                         │
│              └─────────────────┘                         │
│                                                           │
│  Inside Docker:                                          │
│  • user-service:50051   (container name)                 │
│  • expense-service:50052                                 │
│  • app-db:5432                                           │
│                                                           │
│  From your Mac:                                          │
│  • localhost:50051  → user-service (port mapped)         │
│  • localhost:50052  → expense-service (port mapped)      │
│  • localhost:5433   → app-db (port mapped)               │
└──────────────────────────────────────────────────────────┘
```

**Port Mapping Explained:**
```
Container Port  →  Your Mac Port
app-db:5432     →  localhost:5433
user-service:50051 → localhost:50051
expense-service:50052 → localhost:50052
```

**Why different ports?**
- Inside Docker: Services use standard ports (5432, 50051, 50052)
- On your Mac: We map to avoid conflicts (5433 instead of 5432)

---

## 🚀 Production Deployment

### Production Architecture

```
┌──────────────────────────────────────────────────────────┐
│                  KUBERNETES CLUSTER                       │
│                                                           │
│  ┌─────────────────────────────────────────────────┐    │
│  │              Load Balancer                       │    │
│  │         (Public IP: 1.2.3.4)                     │    │
│  └────────┬──────────────────────┬──────────────────┘    │
│           │                      │                        │
│           ▼                      ▼                        │
│  ┌─────────────────┐    ┌──────────────────┐            │
│  │  user-service   │    │ expense-service  │            │
│  │  Pod 1, 2, 3    │    │  Pod 1, 2, 3     │            │
│  └────────┬────────┘    └────────┬─────────┘            │
│           │                      │                        │
│           └──────────┬───────────┘                        │
│                      │                                    │
│                      ▼                                    │
│         ┌─────────────────────────┐                      │
│         │   PostgreSQL (RDS)      │                      │
│         │   (Managed Database)    │                      │
│         └─────────────────────────┘                      │
└──────────────────────────────────────────────────────────┘

Users access: https://api.yourcompany.com
             ↓
        Load Balancer
             ↓
        Service Pods
             ↓
        Database
```

### How to Deploy

```
┌──────────────────────────────────────────────────────────┐
│              DEPLOYMENT PIPELINE                          │
└──────────────────────────────────────────────────────────┘

1. Push Code to GitHub
   ↓
   git push origin main

2. GitHub Actions (CI/CD)
   ↓
   • Runs on Linux (Ubuntu)
   • Builds with Bazel or Go
   • Creates Docker images
   • Pushes to registry
   
3. Docker Registry
   ↓
   ghcr.io/yourname/user-service:v1.0.0
   ghcr.io/yourname/expense-service:v1.0.0

4. Deploy to Kubernetes
   ↓
   kubectl apply -f k8s/deployment.yaml
   
5. Services Running!
   ↓
   Users can access your API
```

---

## 🔧 Common Development Tasks

### Task 1: Make a Code Change

```bash
# 1. Edit the code
vim internal/user/service/user.go

# 2. Quick test (Bazel - 5 seconds)
bazel build //services/user:user_service
bazel run //services/user:user_service

# 3. Test it
grpcurl -plaintext localhost:50051 user.UserService/ListUsers

# 4. If it works, test in Docker (optional)
./run.sh build-docker
./run.sh run
```

### Task 2: Add a New Dependency

```bash
# 1. Add to go.mod
go get github.com/some/package

# 2. Update Bazel
bazel run //:gazelle -- update-repos -from_file=go.mod

# 3. Rebuild
bazel build //...
```

### Task 3: Debug a Service

```bash
# Option 1: Bazel (see logs directly)
bazel run //services/user:user_service
# Logs appear in terminal

# Option 2: Docker (view logs)
./run.sh run
./run.sh logs
# or
docker logs go-grpc-bazel-user-service-1 -f
```

### Task 4: Test with Database

```bash
# Start PostgreSQL
docker run -d -p 5433:5432 \
  -e POSTGRES_PASSWORD=splitwise \
  -e POSTGRES_DB=splitwise \
  postgres:15-alpine

# Run service (it connects to localhost:5433)
bazel run //services/user:user_service

# Test
grpcurl -plaintext -d '{"name":"Alice","email":"alice@test.com"}' \
  localhost:50051 user.UserService/RegisterUser
```

---

## 🎯 Build Caching Explained

### Bazel Caching

```
First Build:
┌─────────────────────────────────────────┐
│ Download dependencies    │ 30s          │
│ Compile proto files      │ 10s          │
│ Compile Go code          │ 20s          │
│ Link binary              │ 5s           │
└─────────────────────────────────────────┘
Total: 65 seconds

Second Build (no changes):
┌─────────────────────────────────────────┐
│ Use cached binary        │ 0.5s         │
└─────────────────────────────────────────┘
Total: 0.5 seconds (130x faster!)

Third Build (changed one file):
┌─────────────────────────────────────────┐
│ Recompile changed file   │ 2s           │
│ Relink binary            │ 1s           │
└─────────────────────────────────────────┘
Total: 3 seconds (22x faster!)
```

### Docker Caching

```
First Build:
┌─────────────────────────────────────────┐
│ Pull base image          │ 30s          │
│ Install dependencies     │ 20s          │
│ Download Go modules      │ 40s          │
│ Build binaries           │ 60s          │
│ Create final image       │ 10s          │
└─────────────────────────────────────────┘
Total: 160 seconds

Second Build (code changed):
┌─────────────────────────────────────────┐
│ Use cached base image    │ 0s           │
│ Use cached dependencies  │ 0s           │
│ Use cached Go modules    │ 0s           │
│ Build binaries           │ 30s (faster) │
│ Create final image       │ 5s           │
└─────────────────────────────────────────┘
Total: 35 seconds (4.5x faster!)
```

---

## 📊 When to Use What?

### Decision Matrix

| Task | Use Bazel | Use Docker | Why |
|------|-----------|------------|-----|
| **Quick code change** | ✅ | ❌ | Bazel is 10x faster |
| **Test one service** | ✅ | ❌ | No need for containers |
| **Test all services together** | ❌ | ✅ | Need full stack |
| **Test with Keycloak** | ❌ | ✅ | Keycloak needs Docker |
| **Before pushing to Git** | ✅ | ✅ | Test both ways |
| **Production deployment** | ❌ | ✅ | Docker is portable |

### Recommended Daily Workflow

```
Morning:
├─ Start PostgreSQL in Docker
│  docker run -d -p 5433:5432 postgres:15-alpine
│
├─ Use Bazel for development
│  bazel run //services/user:user_service
│
└─ Make changes, test quickly with Bazel

Before Lunch:
├─ Test full stack with Docker
│  ./run.sh build-docker
│  ./run.sh run
│
└─ Make sure everything works together

Before Going Home:
├─ Run full Docker test
├─ Push to Git
└─ CI/CD builds and tests automatically
```

---

## 🐛 Troubleshooting Guide

### "Port already in use"

```bash
# Find what's using the port
lsof -i :50051

# Kill it
lsof -ti:50051 | xargs kill -9

# Or use different port
APP_SERVER_PORT=:50053 bazel run //services/user:user_service
```

### "Can't connect to database"

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -p 5433 -U splitwise -d splitwise

# Restart database
docker restart <container-id>
```

### "Bazel build is slow"

```bash
# Clean cache (last resort)
bazel clean --expunge

# Check disk space
df -h

# Rebuild
bazel build //...
```

### "Docker build fails"

```bash
# Clean everything
docker system prune -a

# Rebuild from scratch
./run.sh build-docker
```

---

## 📝 Quick Reference

### Essential Commands

```bash
# BAZEL
bazel build //...                    # Build everything
bazel run //services/user:user_service  # Build and run
bazel clean                          # Clean build cache

# DOCKER
./run.sh build-docker                # Build images
./run.sh run                         # Start services
./run.sh stop                        # Stop services
./run.sh logs                        # View logs
./run.sh clean                       # Clean everything

# TESTING
grpcurl -plaintext localhost:50051 list  # List services
grpcurl -plaintext localhost:50051 user.UserService/ListUsers  # Call method

# DATABASE
docker exec -it go-grpc-bazel-app-db-1 psql -U splitwise  # Connect to DB
```

### File Structure

```
go-grpc-bazel/
├── services/          # Service entry points
├── internal/          # Business logic (Clean Architecture)
├── pkg/              # Shared packages
├── proto/            # gRPC definitions
├── config/           # Configuration files
├── Dockerfile        # Docker build instructions
├── docker-compose.yml # Multi-container setup
├── WORKSPACE         # Bazel workspace
└── BUILD.bazel       # Bazel build rules
```

---

## 🎓 Key Concepts Summary

1. **Bazel** = Fast local builds with smart caching
2. **Docker** = Portable containers for production
3. **Localhost** = Your Mac's network
4. **Container Network** = Isolated Docker network
5. **Port Mapping** = Bridge between Mac and containers
6. **Caching** = Reuse previous work to speed up builds
7. **Clean Architecture** = Organized code in layers
8. **gRPC** = Fast, type-safe API communication

---

**You're now ready to be productive!** 🚀

Start with Bazel for daily development, use Docker before deploying, and you'll have a smooth workflow.
