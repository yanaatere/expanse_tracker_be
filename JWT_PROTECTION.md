# API Authentication & Authorization Summary

## ✅ JWT Protection Status

All sensitive API endpoints now require JWT authentication. This document provides a complete overview of which endpoints are protected and which are public.

---

## 🔓 Public Endpoints (No Authentication Required)

### Authentication Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| **POST** | `/api/auth/register` | Register new user account |
| **POST** | `/api/auth/login` | Authenticate user and get JWT token |
| **POST** | `/api/auth/forgot-password` | Request password reset |
| **POST** | `/api/auth/reset-password` | Reset password using token |

---

## 🔒 Protected Endpoints (JWT Authentication Required)

### User Endpoints

All user endpoints require valid JWT token in `Authorization` header.

| Method | Endpoint | Purpose | Auth |
|--------|----------|---------|------|
| **GET** | `/api/users` | Get all users | ✅ Required |
| **GET** | `/api/users/{id}` | Get specific user | ✅ Required |
| **PUT** | `/api/users/{id}` | Update user | ✅ Required |
| **DELETE** | `/api/users/{id}` | Delete user | ✅ Required |

### Balance Endpoints

All balance endpoints require valid JWT token in `Authorization` header.

| Method | Endpoint | Purpose | Auth |
|--------|----------|---------|------|
| **GET** | `/api/balance` | Get current balance | ✅ Required |
| **GET** | `/api/balance/monthly` | Get monthly balance | ✅ Required |
| **GET** | `/api/balance/range` | Get balance for date range | ✅ Required |
| **GET** | `/api/balance/category` | Get balance by category | ✅ Required |
| **POST** | `/api/balance/recalculate` | Recalculate balance | ✅ Required |

### Category Endpoints

All category endpoints require valid JWT token in `Authorization` header.

| Method | Endpoint | Purpose | Auth |
|--------|----------|---------|------|
| **GET** | `/api/categories` | Get all categories | ✅ Required |
| **GET** | `/api/categories/{id}` | Get specific category | ✅ Required |
| **POST** | `/api/categories` | Create new category | ✅ Required |
| **PUT** | `/api/categories/{id}` | Update category | ✅ Required |
| **DELETE** | `/api/categories/{id}` | Delete category | ✅ Required |

### Transaction Endpoints

All transaction endpoints require valid JWT token in `Authorization` header.

| Method | Endpoint | Purpose | Auth |
|--------|----------|---------|------|
| **GET** | `/api/transactions` | Get all transactions | ✅ Required |
| **GET** | `/api/transactions/{id}` | Get specific transaction | ✅ Required |
| **POST** | `/api/transactions` | Create new transaction | ✅ Required |
| **PUT** | `/api/transactions/{id}` | Update transaction | ✅ Required |
| **DELETE** | `/api/transactions/{id}` | Delete transaction | ✅ Required |

### Dashboard Endpoints

Dashboard stats endpoint requires valid JWT token in `Authorization` header.

| Method | Endpoint | Purpose | Auth |
|--------|----------|---------|------|
| **GET** | `/api/dashboard/stats` | Get dashboard statistics | ✅ Required |

---

## 📋 Complete Authentication Flow

### Step 1: Registration or Login

First, you must register or login to get a JWT token:

```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "securepassword123"
  }'

# OR Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

**Response:**
```json
{
  "id": 1,
  "username": "john_doe",
  "email": "john@example.com",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Step 2: Use Token for Protected Endpoints

Include the token in the `Authorization` header of subsequent requests:

```bash
curl -X GET http://localhost:8080/api/balance \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## 🔑 Token Details

### JWT Token Structure

Each JWT token contains:
- **UserID** - Unique user identifier
- **Email** - User's email address
- **Username** - User's username
- **Expiration** - Token expires after 24 hours
- **Issued At** - Token creation timestamp

### Token Lifespan

- **Duration:** 24 hours from creation
- **Renewal:** User must login again after expiration

### Token Storage (Frontend)

Best practice is to store the token in:
```javascript
// Option 1: localStorage
localStorage.setItem('token', response.token);

// Option 2: sessionStorage (more secure)
sessionStorage.setItem('token', response.token);

// Option 3: httpOnly cookie (most secure)
// Set server-side during login response
```

---

## ⚠️ What Happens Without Token

If you attempt to access a protected endpoint without a valid token:

```bash
curl -X GET http://localhost:8080/api/balance

# Response:
# HTTP Status: 401 Unauthorized
# Body: "Authorization header required"
```

### Missing Token Scenarios

1. **No Authorization header:** Returns `401 Unauthorized`
2. **Invalid format:** Returns `401 Unauthorized`
3. **Expired token:** Returns `401 Unauthorized`
4. **Malformed token:** Returns `401 Unauthorized`

---

## 📝 Authorization Header Format

The Authorization header must follow this exact format:

```
Authorization: Bearer <token>
```

### Valid Examples:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Authorization: Bearer token_string_here
```

### Invalid Examples:
```
Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  ❌ Missing "Bearer"
Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  ❌ Wrong prefix
Authorization: BearereyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  ❌ Missing space
```

---

## 🔐 Security Implementation

### Password Protection
- Passwords are hashed using **bcrypt** (not stored in plaintext)
- Default cost factor: 10 (takes ~100ms to hash)

### Token Security
- Generated using **HMAC-SHA256**
- Secret key configured via `JWT_SECRET` environment variable
- 24-hour expiration window

### Sensitive Data
- Tokens are masked in logs as `[token present]`
- Request bodies (containing passwords) are never logged
- Response bodies (containing sensitive data) are never logged

---

## 🧪 Testing Protected Endpoints

### With Valid Token:
```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}' | jq -r '.token')

# Use token
curl -X GET http://localhost:8080/api/balance \
  -H "Authorization: Bearer $TOKEN"
```

### Without Token (Should Fail):
```bash
curl -X GET http://localhost:8080/api/balance
# Result: 401 Unauthorized
```

### With Expired Token (Should Fail):
```bash
# After 24 hours have passed
curl -X GET http://localhost:8080/api/balance \
  -H "Authorization: Bearer <old_token>"
# Result: 401 Unauthorized
```

---

## 📊 Summary Table

| Category | Total | Protected | Public |
|----------|-------|-----------|--------|
| Auth | 4 | 0 | 4 |
| User | 4 | 4 | 0 |
| Balance | 5 | 5 | 0 |
| Category | 5 | 5 | 0 |
| Transaction | 6 | 6 | 0 |
| Dashboard | 1 | 1 | 0 |
| **TOTAL** | **25** | **21** | **4** |

---

## 🚀 Frontend Integration Example

```javascript
// Fetch API with interceptor pattern
class API {
  constructor(baseURL) {
    this.baseURL = baseURL;
  }

  async login(email, password) {
    const response = await fetch(`${this.baseURL}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    
    if (!response.ok) throw new Error('Login failed');
    const data = await response.json();
    localStorage.setItem('token', data.token);
    return data;
  }

  async fetchWithAuth(endpoint, options = {}) {
    const token = localStorage.getItem('token');
    
    if (!token) {
      throw new Error('Not authenticated');
    }

    const headers = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...options.headers
    };

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers
    });

    if (response.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('token');
      window.location.href = '/login';
      throw new Error('Authentication required');
    }

    return response.json();
  }

  // Usage
  async getBalance() {
    return this.fetchWithAuth('/api/balance');
  }
}

// Usage in app
const api = new API('http://localhost:8080');
api.getBalance().then(balance => console.log(balance));
```

---

## ✅ Verification Checklist

After deployment, verify JWT protection is working:

- [ ] Test registration endpoint works without token ✅
- [ ] Test login endpoint works without token ✅
- [ ] Test balance endpoint fails without token ❌ (401)
- [ ] Test balance endpoint works with valid token ✅
- [ ] Test balance endpoint fails with invalid token ❌ (401)
- [ ] Test category endpoint requires token ✅
- [ ] Test transaction endpoint requires token ✅
- [ ] Test user endpoints require token ✅
- [ ] Test dashboard endpoint requires token ✅
- [ ] Verify token expires after 24 hours ✅

---

## 🔗 Related Documentation

- [AUTH_API.md](AUTH_API.md) - Complete authentication API documentation
- [LOGGING.md](LOGGING.md) - Request/response logging details
- [README.md](README.md) - Project overview

