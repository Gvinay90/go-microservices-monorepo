# Keycloak Authentication Guide

## Quick Start

### 1. Start Keycloak

```bash
docker-compose up -d keycloak
```

Wait for Keycloak to be ready (about 30 seconds). Access at: http://localhost:8080

**Admin Credentials**: `admin` / `admin`

### 2. Enable Authentication

Edit `config/base.yaml`:
```yaml
auth:
  enabled: true  # Change from false to true
```

### 3. Test Users

Pre-configured users in the `expense-sharing` realm:

| Username | Password | Email | Roles |
|----------|----------|-------|-------|
| alice | password | alice@example.com | user |
| bob | password | bob@example.com | user |
| admin | admin123 | admin@example.com | user, admin |

### 4. Get a Token

```bash
# Login as alice
curl -X POST 'http://localhost:8080/realms/expense-sharing/protocol/openid-connect/token' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'client_id=client-app' \
  -d 'username=alice' \
  -d 'password=password' \
  -d 'grant_type=password' | jq -r '.access_token'
```

Save the token to a variable:
```bash
TOKEN=$(curl -s -X POST 'http://localhost:8080/realms/expense-sharing/protocol/openid-connect/token' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'client_id=client-app' \
  -d 'username=alice' \
  -d 'password=password' \
  -d 'grant_type=password' | jq -r '.access_token')
```

### 5. Make Authenticated Requests

Using `grpcurl`:

```bash
# Register a user (requires authentication)
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"name": "Alice Smith", "email": "alice@example.com"}' \
  localhost:50051 user.UserService/RegisterUser

# Create an expense
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{
    "description": "Dinner",
    "total_amount": 100.0,
    "paid_by": "alice@example.com",
    "splits": [
      {"user_id": "alice@example.com", "amount": 50.0},
      {"user_id": "bob@example.com", "amount": 50.0}
    ]
  }' \
  localhost:50052 expense.ExpenseService/CreateExpense
```

## Authorization Rules

### User Service

| Method | Required Role | Notes |
|--------|---------------|-------|
| RegisterUser | Authenticated | Any logged-in user |
| GetUser | Authenticated | Own data or admin |
| ListUsers | admin | Admin only |
| AddFriend | Authenticated | Any logged-in user |
| GetFriends | Authenticated | Own friends or admin |

### Expense Service

| Method | Required Role | Notes |
|--------|---------------|-------|
| CreateExpense | Authenticated | Any logged-in user |
| GetExpense | Authenticated | Participants or admin |
| ListExpenses | Authenticated | Own expenses or admin |
| SettleBalance | Authenticated | Involved users only |

## Development Mode

To disable authentication during development:

```yaml
# config/base.yaml
auth:
  enabled: false
```

Services will work without tokens when `enabled: false`.

## Keycloak Admin Console

Access: http://localhost:8080

1. Login with `admin` / `admin`
2. Select `expense-sharing` realm
3. Manage users, roles, and clients

### Adding New Users

1. Go to **Users** → **Add user**
2. Set username and email
3. Click **Save**
4. Go to **Credentials** tab
5. Set password (uncheck "Temporary")
6. Go to **Role mapping** tab
7. Assign roles (`user` or `admin`)

## Troubleshooting

### Token Expired
Tokens expire after 5 minutes by default. Get a new token.

### Invalid Token
Check that:
- Keycloak is running
- `auth.keycloak_url` in config matches Keycloak URL
- Token is passed in `authorization: Bearer <token>` header

### Permission Denied
Check user roles in Keycloak admin console.

## Architecture

```
Client Request
    ↓
[Authorization: Bearer <JWT>]
    ↓
gRPC Service
    ↓
Auth Interceptor
    ├─ Extract Token
    ├─ Validate with Keycloak
    ├─ Extract Claims (user ID, roles)
    └─ Inject into Context
    ↓
Service Layer
    ├─ Check Roles (admin vs user)
    └─ Check Ownership
    ↓
Response
```

## Production Considerations

1. **HTTPS**: Use TLS in production
2. **Token Storage**: Store tokens securely (not in plain text)
3. **Refresh Tokens**: Implement token refresh flow
4. **Rate Limiting**: Add rate limiting to prevent abuse
5. **Audit Logging**: Log all authentication attempts
