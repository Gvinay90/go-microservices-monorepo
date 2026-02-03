# Keycloak Setup Guide

This guide walks you through setting up Keycloak for the expense sharing application.

## Quick Start

### 1. Access Keycloak Admin Console

```bash
# Keycloak is running at:
http://localhost:8080

# Default admin credentials:
Username: admin
Password: admin
```

### 2. Create Realm

1. Click on the dropdown in the top-left (currently shows "master")
2. Click "Create Realm"
3. Enter realm name: `expense-sharing`
4. Click "Create"

### 3. Create Client for API Gateway

1. In the `expense-sharing` realm, go to **Clients**
2. Click "Create client"
3. Configure:
   - **Client ID**: `expense-app`
   - **Client Protocol**: `openid-connect`
   - Click "Next"
4. **Capability config**:
   - ✅ Client authentication: ON
   - ✅ Authorization: OFF
   - ✅ Authentication flow:
     - ✅ Standard flow
     - ✅ Direct access grants (for password grant)
     - ✅ Service accounts roles (for Admin API)
   - Click "Next"
5. **Login settings**:
   - Valid redirect URIs: `http://localhost:3000/*`
   - Web origins: `http://localhost:3000`
   - Click "Save"

### 4. Get Client Secret

1. Go to **Clients** → `expense-app`
2. Click on **Credentials** tab
3. Copy the **Client Secret**
4. Update `config/base.yaml`:
   ```yaml
   auth:
     client_secret: "<paste-your-secret-here>"
   ```

### 5. Configure Service Account Roles

For the Admin API to create users, the client needs proper roles:

1. Go to **Clients** → `expense-app`
2. Click on **Service account roles** tab
3. Click "Assign role"
4. Filter by clients: Select `realm-management`
5. Assign these roles:
   - `manage-users`
   - `view-users`
   - `create-client`
6. Click "Assign"

### 6. Configure Realm Settings (Optional)

1. Go to **Realm settings**
2. **Login** tab:
   - ✅ User registration: OFF (we handle this via API)
   - ✅ Email as username: ON
   - ✅ Login with email: ON
3. **Tokens** tab:
   - Access Token Lifespan: 1 hour (default is fine)
   - Refresh Token Lifespan: 1 day

## Testing the Setup

### 1. Test User Registration

```bash
# Register a new user
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Expected Response:**
```json
{
  "user": {
    "id": "user_john@example.com_...",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "message": "User registered successfully. You can now login."
}
```

### 2. Verify User in Keycloak

1. Go to **Users** in Keycloak admin console
2. You should see `john@example.com` listed
3. Click on the user to see details

### 3. Test Login

```bash
# Login with the user
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Expected Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### 4. Test Authenticated Request

```bash
# Use the access token from login
TOKEN="<your-access-token>"

curl http://localhost:8081/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

## Architecture

```
Registration Flow:
1. User → API Gateway (/auth/register)
2. API Gateway → Keycloak Admin API (create user)
3. API Gateway → User Service (store user data)
4. Response → User

Login Flow:
1. User → API Gateway (/auth/login)
2. API Gateway → Keycloak Token Endpoint
3. Keycloak validates credentials
4. Keycloak → JWT Token
5. Response → User with JWT

Authenticated Request:
1. User → API Gateway (with JWT)
2. API Gateway validates JWT with Keycloak
3. API Gateway → Internal Service (forwarded JWT)
4. Response → User
```

## Troubleshooting

### Issue: "Failed to create user in Keycloak"

**Solution:**
- Check client secret is correct in `config/base.yaml`
- Verify service account roles are assigned
- Check Keycloak logs: `docker compose logs keycloak`

### Issue: "Invalid credentials" on login

**Solution:**
- Verify user exists in Keycloak admin console
- Check password was set correctly during registration
- Try resetting password in Keycloak admin console

### Issue: "Keycloak connection refused"

**Solution:**
- Ensure Keycloak is running: `docker compose ps keycloak`
- Check Keycloak URL in config uses service name: `http://keycloak:8080`
- Wait for Keycloak to fully start (can take 30-60 seconds)

### Issue: "Token validation failed"

**Solution:**
- Ensure auth is enabled in `config/base.yaml`
- Check realm name matches: `expense-sharing`
- Verify JWT is not expired

## Production Considerations

### Security

1. **Change Admin Password:**
   ```bash
   # In docker-compose.yml, update:
   KEYCLOAK_ADMIN_PASSWORD: <strong-password>
   ```

2. **Use HTTPS:**
   - Configure SSL certificates
   - Update URLs to use `https://`

3. **Rotate Client Secrets:**
   - Regularly rotate the client secret
   - Update in both Keycloak and config

### High Availability

1. **Database:**
   - Use external PostgreSQL cluster
   - Configure connection pooling

2. **Keycloak Clustering:**
   - Run multiple Keycloak instances
   - Configure load balancer

3. **Backup:**
   - Regular database backups
   - Export realm configuration

## Next Steps

1. ✅ Configure Keycloak realm
2. ✅ Create client and get secret
3. ✅ Update config with client secret
4. ✅ Test registration and login
5. 🔄 Rebuild and restart services
6. 🔄 Test from React frontend

## Useful Commands

```bash
# View Keycloak logs
docker compose logs -f keycloak

# Restart Keycloak
docker compose restart keycloak

# Access Keycloak database
docker compose exec keycloak-db psql -U keycloak

# Export realm configuration
# (Do this from Keycloak admin console: Realm settings → Export)
```

## Summary

With Keycloak integration:
- 🔒 Secure password hashing (bcrypt)
- 🎫 Industry-standard JWT tokens
- 👥 Centralized user management
- 🔑 OAuth2/OIDC support
- 📊 Built-in admin console
- 🚀 Production-ready authentication
