# API Gateway Security Architecture

This document explains the security architecture of the API Gateway pattern implementation.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    API GATEWAY (Security Layer)              │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Rate Limiting│  │ Audit Logging│  │ Auth/AuthZ   │     │
│  │ 100 req/min  │  │ All requests │  │ JWT Validate │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                              │
└────────────────────────────┬─────────────────────────────────┘
                             │ Forwarded JWT
┌────────────────────────────┼─────────────────────────────────┐
│              INTERNAL SERVICES (Business Logic)              │
│                            │                                  │
│  ┌─────────────────────────▼──────────────┐                 │
│  │  Lightweight Auth (Trust Gateway)      │                 │
│  │  - Extract claims from forwarded JWT   │                 │
│  │  - Business-level authorization        │                 │
│  └────────────────────────────────────────┘                 │
│                                                              │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │ User Service │         │Expense Service│                 │
│  │   (:50051)   │         │   (:50052)    │                 │
│  └──────────────┘         └──────────────┘                 │
└──────────────────────────────────────────────────────────────┘
```

## Security Layers

### 1. API Gateway (Primary Security)

**Location:** `api-gateway/middleware/`

#### Rate Limiting
- **File:** `ratelimit.go`
- **Algorithm:** Token bucket
- **Default:** 100 requests per minute per user/IP
- **Identification:** User ID (from JWT) or IP address
- **Benefits:**
  - Protects all backend services
  - Prevents abuse and DoS attacks
  - Single point of control

#### Audit Logging
- **File:** `audit.go`
- **Logs:**
  - All authentication attempts
  - All write operations (POST, PUT, DELETE)
  - All failed requests (4xx, 5xx)
- **Data Captured:**
  - User ID, email
  - IP address, user agent
  - Request method, path
  - Response status, duration
- **Benefits:**
  - Centralized audit trail
  - Security monitoring
  - Compliance requirements

#### Authentication/Authorization
- **File:** `auth.go`
- **Validates:**
  - JWT token signature
  - Token expiration
  - User roles and permissions
- **Forwards:**
  - Valid JWT to internal services
  - User claims in context

### 2. Internal Services (Defense in Depth)

**Location:** `services/user/main.go`, `services/expense/main.go`

#### Lightweight Auth
- **Purpose:** Trust but verify
- **Validates:**
  - JWT format (already validated by gateway)
  - Extracts user claims
  - Business-level authorization
- **Does NOT:**
  - Re-validate JWT signature (gateway did this)
  - Rate limit (gateway handles this)
  - Audit log HTTP requests (gateway logs this)

## Request Flow

### 1. Unauthenticated Request
```
Client → API Gateway
         ├─ Rate Limit Check ✓
         ├─ Audit Log: "auth attempt"
         └─ Auth Check ✗ → 401 Unauthorized
```

### 2. Authenticated Request
```
Client → API Gateway
         ├─ Rate Limit Check ✓
         ├─ Auth Check ✓
         ├─ Audit Log: "POST /api/v1/expenses"
         └─ Forward JWT → Internal Service
                          ├─ Extract Claims
                          ├─ Business Logic
                          └─ Response
```

### 3. Rate Limited Request
```
Client → API Gateway
         ├─ Rate Limit Check ✗
         ├─ Audit Log: "rate limit exceeded"
         └─ 429 Too Many Requests
```

## Configuration

### API Gateway
```yaml
# No rate limit config needed - hardcoded in middleware
# Rate: 100 requests/minute per user/IP
```

### Internal Services
```yaml
auth:
  enabled: true  # Lightweight auth for defense in depth
  # No rate limiting or audit logging config
```

## Benefits of This Architecture

### ✅ Centralized Security
- All security policies in one place
- Easier to update and maintain
- Consistent enforcement

### ✅ Performance
- No duplicate JWT validation
- No duplicate rate limiting
- Faster internal service calls

### ✅ Simplicity
- Internal services focus on business logic
- Less code to maintain
- Clearer separation of concerns

### ✅ Defense in Depth
- Gateway validates everything
- Services still verify basics
- Multiple layers of protection

## Monitoring

### Metrics to Track
1. **Rate Limit Hits:** How often users hit limits
2. **Auth Failures:** Failed login attempts
3. **Audit Events:** Security-relevant actions
4. **Response Times:** Impact of middleware

### Logs to Monitor
```
# Rate limiting
level=WARN msg="Rate limit exceeded" identifier=user:123 path=/api/v1/expenses

# Audit logging
level=INFO msg="audit" method=POST path=/api/v1/auth/login status=401 ip=192.168.1.1

# Authentication
level=WARN msg="Invalid token" error="token expired"
```

## Security Best Practices

### ✅ Do
- Keep rate limits reasonable (100/min is good start)
- Monitor audit logs regularly
- Review failed auth attempts
- Update JWT secrets regularly

### ❌ Don't
- Disable rate limiting in production
- Skip audit logging
- Trust internal services completely (keep lightweight auth)
- Log sensitive data (passwords, tokens)

## Future Enhancements

### Potential Improvements
1. **Dynamic Rate Limiting:** Adjust based on user tier
2. **Distributed Rate Limiting:** Use Redis for multi-instance
3. **Advanced Audit:** Send to SIEM system
4. **Metrics Export:** Prometheus integration
5. **Circuit Breaker:** Protect backend services

## Troubleshooting

### Rate Limit Issues
```bash
# Check current rate limit status
curl -v http://localhost:8081/api/v1/users
# Look for X-RateLimit headers (if implemented)
```

### Audit Log Issues
```bash
# View audit logs
docker compose logs api-gateway | grep "audit"
```

### Auth Issues
```bash
# Test authentication
curl -H "Authorization: Bearer <token>" http://localhost:8081/api/v1/auth/me
```

## Summary

This architecture provides **defense in depth** while maintaining **simplicity and performance**:

- **API Gateway:** Full security (rate limit + audit + auth)
- **Internal Services:** Lightweight auth only
- **Result:** Secure, fast, maintainable system
