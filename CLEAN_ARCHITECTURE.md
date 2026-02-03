# Clean Architecture Guide

This document explains the Clean Architecture (Hexagonal Architecture) pattern implemented in the User Service.

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                  Transport Layer                     │
│         internal/user/handler/grpc.go                │
│  (Converts Protobuf ↔ Domain, handles gRPC)         │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│               Business Logic Layer                   │
│         internal/user/service/user.go                │
│  (Validation, ID generation, business rules)         │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│              Data Access Layer                       │
│       internal/user/repository/gorm.go               │
│  (GORM operations, Domain ↔ Storage mapping)        │
└─────────────────────────────────────────────────────┘

         ┌──────────────────────────┐
         │    Domain Layer          │
         │ internal/user/domain/    │
         │ (Pure Go, no deps)       │
         └──────────────────────────┘
```

## Layer Responsibilities

### 1. Domain Layer (`internal/user/domain/`)
**Purpose**: Define core business entities and rules.

**Files**:
- `user.go`: Pure Go struct with no external dependencies
- `errors.go`: Domain-specific errors

**Key Principles**:
- ✅ No GORM tags
- ✅ No Protobuf tags
- ✅ No framework dependencies
- ✅ Business logic methods (e.g., `AddFriend()`)

**Example**:
```go
type User struct {
    ID      string
    Name    string
    Email   string
    Friends []string // Just IDs, not full objects
}

func (u *User) AddFriend(friendID string) error {
    if friendID == u.ID {
        return ErrCannotFriendSelf
    }
    // ...
}
```

### 2. Repository Layer (`internal/user/repository/`)
**Purpose**: Abstract data access.

**Files**:
- `repository.go`: Interface defining operations
- `gorm.go`: GORM implementation

**Key Principles**:
- ✅ Interface-first design (easy to swap DB)
- ✅ Maps `domain.User` ↔ `storage.User`
- ✅ Returns domain errors (not GORM errors)

**Example**:
```go
type Repository interface {
    Create(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id string) (*domain.User, error)
    // ...
}
```

### 3. Service Layer (`internal/user/service/`)
**Purpose**: Implement business logic.

**Files**:
- `service.go`: Interface defining business operations
- `user.go`: Implementation

**Key Principles**:
- ✅ Validates inputs
- ✅ Enforces business rules
- ✅ Orchestrates repository calls
- ✅ Returns domain models

**Example**:
```go
func (s *userService) Register(ctx context.Context, name, email string) (*domain.User, error) {
    // 1. Check if user exists
    existing, _ := s.repo.FindByEmail(ctx, email)
    if existing != nil {
        return nil, domain.ErrUserAlreadyExists
    }
    
    // 2. Create domain user (with validation)
    user, err := domain.NewUser(generateID(email), name, email)
    if err != nil {
        return nil, err
    }
    
    // 3. Persist
    return user, s.repo.Create(ctx, user)
}
```

### 4. Handler Layer (`internal/user/handler/`)
**Purpose**: Handle transport-specific concerns (gRPC).

**Files**:
- `grpc.go`: gRPC server implementation

**Key Principles**:
- ✅ Converts Protobuf ↔ Domain models
- ✅ Maps domain errors to gRPC status codes
- ✅ Thin layer (no business logic)

**Example**:
```go
func (h *GRPCHandler) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
    // Call service
    user, err := h.svc.Register(ctx, req.Name, req.Email)
    if err != nil {
        return nil, h.mapError(err) // Convert to gRPC error
    }
    
    // Convert to protobuf
    return &pb.RegisterUserResponse{
        User: toPBUser(user),
    }, nil
}
```

## Dependency Injection (Wiring)

In `services/user/main.go`, layers are wired together:

```go
// 1. Repository (depends on DB)
userRepo := repository.NewGormRepository(db)

// 2. Service (depends on Repository)
userService := service.NewUserService(userRepo, log)

// 3. Handler (depends on Service)
userHandler := handler.NewGRPCHandler(userService, log)

// 4. Register with gRPC
pb.RegisterUserServiceServer(s, userHandler)
```

**Dependency Flow**: `Handler → Service → Repository → DB`

## Benefits Demonstrated

### 1. Testability
Each layer can be tested independently:

```go
// Test Service without gRPC or DB
func TestUserService_Register(t *testing.T) {
    mockRepo := &MockRepository{}
    svc := service.NewUserService(mockRepo, log)
    
    user, err := svc.Register(ctx, "Alice", "alice@example.com")
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

### 2. Flexibility
Easy to add new transports (REST, GraphQL) without changing business logic:

```
gRPC Handler ──┐
               ├──> User Service ──> Repository
REST Handler ──┘
```

### 3. Maintainability
- Business rules live in **one place** (Service layer)
- Database logic isolated (Repository layer)
- Transport logic isolated (Handler layer)

### 4. Domain-Driven Design
The `domain` package represents your **business model** in pure Go, independent of frameworks.

## Comparison: Before vs After

### Before (Simplified)
```
services/user/server.go
  ├─ gRPC handler methods
  ├─ Input validation
  ├─ Business logic
  └─ GORM database calls
```
**Problem**: Everything coupled together. Hard to test. Hard to reuse logic.

### After (Clean Architecture)
```
internal/user/
  ├─ domain/       # Pure business model
  ├─ repository/   # Data access abstraction
  ├─ service/      # Business logic
  └─ handler/      # gRPC transport
```
**Benefit**: Clear separation. Easy to test. Easy to extend.

## When to Use This Pattern

| **Use Clean Architecture** | **Use Simplified** |
|:---|:---|
| Multiple transports (gRPC + REST) | Single transport (only gRPC) |
| Complex business rules | Simple CRUD operations |
| Team of 3+ developers | Solo developer or small team |
| Long-term maintenance (2+ years) | Rapid prototyping |

## Next Steps

To apply this pattern to the Expense Service:
1. Create `internal/expense/domain/expense.go`
2. Create `internal/expense/repository/repository.go`
3. Create `internal/expense/service/expense.go`
4. Create `internal/expense/handler/grpc.go`
5. Wire in `services/expense/main.go`

This creates a consistent, scalable architecture across all services.
