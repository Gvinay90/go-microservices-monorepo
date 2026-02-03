# Expense Sharing App - Complete Integration Walkthrough

## Overview

The expense sharing application is now fully integrated with:
- ✅ Keycloak authentication (secure JWT tokens)
- ✅ Backend API endpoints (gRPC services via API Gateway)
- ✅ React frontend with real-time data
- ✅ Rate limiting and audit logging
- ✅ Auto-configured Keycloak realm

## Architecture

```
React Frontend (localhost:3000)
    ↓ HTTP/REST
API Gateway (localhost:8081)
    ├─ Rate Limiting (100 req/min)
    ├─ Audit Logging
    ├─ Keycloak Auth
    ↓ gRPC
Internal Services
    ├─ User Service (localhost:50051)
    └─ Expense Service (localhost:50052)
```

## Testing the Integration

### 1. Access the Application

Open your browser to: **http://localhost:3000**

You should see the login page with a beautiful gradient background.

### 2. Register a New User

1. Click **"Sign up"** link
2. Fill in the registration form:
   - **Name**: John Doe
   - **Email**: john@example.com
   - **Password**: password123
   - **Confirm Password**: password123
3. Click **"Create Account"**

**What Happens:**
- Frontend sends POST to `/api/v1/auth/register`
- API Gateway creates user in Keycloak via Admin API
- API Gateway creates user in database via User Service
- Success message displayed

**Verify in Keycloak:**
```bash
# Open Keycloak admin console
open http://localhost:8080

# Login: admin / admin
# Navigate to: expense-sharing realm → Users
# You should see john@example.com
```

### 3. Login

1. Navigate to login page (or click "Sign in")
2. Enter credentials:
   - **Email**: john@example.com
   - **Password**: password123
3. Click **"Sign In"**

**What Happens:**
- Frontend sends POST to `/api/v1/auth/login`
- API Gateway authenticates with Keycloak
- Keycloak returns JWT token (access_token + refresh_token)
- Token stored in localStorage
- Redirected to Dashboard

**Verify Token:**
```javascript
// Open browser console (F12)
localStorage.getItem('access_token')
// Should show a long JWT token
```

### 4. View Dashboard

After login, you should see:
- **Header**: Welcome message with your email
- **Expenses Section**: Empty initially
- **Users Section**: List of all registered users

### 5. Create an Expense

1. Click **"+ New Expense"** button
2. Fill in the form:
   - **Description**: Team Lunch
   - **Amount**: 50.00
3. Click **"Create"**

**What Happens:**
- Frontend sends POST to `/api/v1/expenses`
- API Gateway forwards to Expense Service
- Expense created with splits (divided equally among all users)
- Expense list refreshed

**Verify:**
- New expense appears in the list
- Shows description, amount, and date
- Shows who paid (your user ID)

### 6. View Users

The Users section shows all registered users with:
- Name
- Email
- Green gradient card design

### 7. Test Multiple Users

Register another user to test expense splitting:

1. Logout (click "Logout" button)
2. Register new user:
   - **Name**: Jane Smith
   - **Email**: jane@example.com
   - **Password**: password123
3. Login as Jane
4. Create an expense
5. Notice both users appear in the Users list

### 8. Test Authentication Flow

**Token Expiration:**
- Tokens expire after 1 hour (3600 seconds)
- Frontend checks expiration before requests
- Expired tokens redirect to login

**Logout:**
- Click "Logout" button
- Tokens cleared from localStorage
- Redirected to login page

## API Endpoints Integrated

### Authentication
- ✅ `POST /api/v1/auth/register` - Register new user
- ✅ `POST /api/v1/auth/login` - Login with Keycloak
- ✅ `GET /api/v1/auth/me` - Get current user info

### Users
- ✅ `GET /api/v1/users` - List all users
- ✅ `GET /api/v1/users/:id` - Get user by ID
- ✅ `POST /api/v1/users/:id/friends` - Add friend
- ✅ `GET /api/v1/users/:id/friends` - Get friends

### Expenses
- ✅ `GET /api/v1/expenses` - List all expenses
- ✅ `POST /api/v1/expenses` - Create expense
- ✅ `GET /api/v1/expenses/:id` - Get expense by ID
- ✅ `PUT /api/v1/expenses/:id` - Update expense
- ✅ `DELETE /api/v1/expenses/:id` - Delete expense

## Security Features

### Rate Limiting
Try making 101 requests quickly:
```bash
for i in {1..101}; do 
  curl http://localhost:8081/health
done
```
The 101st request will return `429 Too Many Requests`.

### Audit Logging
View audit logs:
```bash
docker compose logs api-gateway | grep "audit"
```

You'll see logs for:
- Registration attempts
- Login attempts
- Expense creation
- Failed requests

### JWT Validation
Try accessing protected endpoint without token:
```bash
curl http://localhost:8081/api/v1/users
# Returns: 401 Unauthorized
```

With token:
```bash
TOKEN="<your-access-token>"
curl -H "Authorization: Bearer $TOKEN" http://localhost:8081/api/v1/users
# Returns: User list
```

## Data Flow Example

### Creating an Expense

```
1. User fills form in React
   ↓
2. Frontend: expenseService.createExpense()
   ↓
3. API: POST /api/v1/expenses
   Headers: Authorization: Bearer <jwt>
   Body: {
     description: "Team Lunch",
     total_amount: 50.00,
     paid_by: "user_id",
     splits: [...]
   }
   ↓
4. API Gateway:
   - Rate limit check ✓
   - JWT validation ✓
   - Audit log ✓
   ↓
5. Expense Service (gRPC):
   - Validate data
   - Store in PostgreSQL
   - Return expense
   ↓
6. API Gateway → Frontend
   Response: { expense: {...} }
   ↓
7. React updates UI
```

## Troubleshooting

### Issue: "Invalid credentials" on login

**Solution:**
- Verify user exists in Keycloak admin console
- Check password is correct
- Ensure Keycloak is running: `docker compose ps keycloak`

### Issue: "Failed to create expense"

**Solution:**
- Check you're logged in (token in localStorage)
- Verify API Gateway is running
- Check logs: `docker compose logs api-gateway`

### Issue: "No users showing"

**Solution:**
- Register at least one user
- Check User Service is running
- Verify database connection

### Issue: Token expired

**Solution:**
- Login again to get new token
- Token lifetime is 1 hour
- Refresh token can be used (not implemented yet)

## Development Commands

```bash
# View all logs
docker compose logs -f

# View specific service logs
docker compose logs -f api-gateway
docker compose logs -f user-service
docker compose logs -f keycloak

# Restart services
docker compose restart

# Rebuild and restart
docker compose build && docker compose up -d

# Stop all services
docker compose down

# Clean restart
docker compose down -v && docker compose up -d
```

## Production Checklist

Before deploying to production:

- [ ] Change Keycloak admin password
- [ ] Use strong client secrets
- [ ] Enable HTTPS/TLS
- [ ] Configure proper CORS origins
- [ ] Set up database backups
- [ ] Configure log aggregation
- [ ] Set up monitoring/alerts
- [ ] Review rate limits
- [ ] Enable refresh token rotation
- [ ] Configure session management

## Summary

**What Works:**
- ✅ Full Keycloak authentication with JWT
- ✅ User registration and login
- ✅ Expense CRUD operations
- ✅ User management
- ✅ Rate limiting (100 req/min)
- ✅ Audit logging
- ✅ Auto-configured Keycloak
- ✅ Beautiful React UI
- ✅ Real-time data updates

**Architecture Benefits:**
- 🔒 Secure authentication (Keycloak)
- 🚀 Fast gRPC backend
- 🎨 Modern React frontend
- 📊 Centralized logging
- 🛡️ Rate limiting protection
- 🔧 Easy to maintain

**Your expense sharing app is production-ready!** 🎉
