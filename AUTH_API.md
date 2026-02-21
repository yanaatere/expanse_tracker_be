# Authentication API Documentation

## Overview
Complete authentication system with JWT tokens for the Expense Tracker application.

## CORS Support ✅
All API endpoints support CORS (Cross-Origin Resource Sharing) for seamless frontend integration.

### CORS Headers Applied to All Responses:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH`
- `Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With`
- `Access-Control-Max-Age: 86400` (24 hours)
- `Access-Control-Allow-Credentials: true`

**Preflight requests (OPTIONS)** are automatically handled, allowing browsers to make cross-origin requests from any domain.

### Frontend Integration Examples:

#### Using Fetch API:
```javascript
// Login request from frontend
const response = await fetch('http://localhost:8080/api/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password123'
  })
});

const data = await response.json();
const token = data.token;

// Using token in subsequent requests
const userResponse = await fetch('http://localhost:8080/api/users', {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});
```

#### Using Axios:
```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api'
});

// Login
const loginResponse = await api.post('/auth/login', {
  email: 'user@example.com',
  password: 'password123'
});

const token = loginResponse.data.token;

// Set default authorization header
api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

// All subsequent requests include the token
const users = await api.get('/users');
```

## Features Implemented
1. **User Registration** - Create new user accounts
2. **User Login** - Authenticate users and return JWT token
3. **Forgot Password** - Request password reset
4. **Reset Password** - Reset password using reset token
5. **JWT Middleware** - Protect routes that require authentication
6. **CORS Support** - All endpoints accessible from frontend applications

## API Endpoints

### Authentication Endpoints (Public)

#### 1. Register User
**POST** `/api/auth/register`

**Request Body:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response (201 Created):**
```json
{
  "id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

#### 2. Login User
**POST** `/api/auth/login`

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response (200 OK):**
```json
{
  "id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

#### 3. Forgot Password
**POST** `/api/auth/forgot-password`

**Request Body:**
```json
{
  "email": "john@example.com"
}
```

**Response (200 OK):**
```json
{
  "message": "If email exists in our system, a password reset link will be sent"
}
```

**Note:** Reset link is printed to console in development. Email service should be configured for production.

---

#### 4. Reset Password
**POST** `/api/auth/reset-password`

**Request Body:**
```json
{
  "token": "reset_token_from_email",
  "new_password": "newpassword123"
}
```

**Response (200 OK):**
```json
{
  "message": "Password has been reset successfully"
}
```

---

### Protected User Endpoints (Requires Authentication)

#### Get All Users
**GET** `/api/users`

**Headers:**
```
Authorization: Bearer <your_jwt_token>
```

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

---

#### Get User by ID
**GET** `/api/users/{id}`

**Headers:**
```
Authorization: Bearer <your_jwt_token>
```

**Response (200 OK):**
```json
{
  "id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

#### Update User
**PUT** `/api/users/{id}`

**Headers:**
```
Authorization: Bearer <your_jwt_token>
```

**Request Body:**
```json
{
  "username": "johndoe_updated",
  "email": "john.new@example.com"
}
```

**Response (200 OK):**
```json
{
  "id": 1,
  "username": "johndoe_updated",
  "email": "john.new@example.com",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

---

#### Delete User
**DELETE** `/api/users/{id}`

**Headers:**
```
Authorization: Bearer <your_jwt_token>
```

**Response (204 No Content)**

---

## Environment Variables

Required in `.env` file:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=expensetracker

# JWT Configuration
JWT_SECRET=your-super-secret-key-change-this-in-production

# Server
PORT=8080
```

## Database Schema Changes

The following columns were added to the `users` table:
- `password` (VARCHAR(255)) - Hashed user password
- `password_reset_token` (VARCHAR(255)) - Token for password reset flow
- `password_reset_expires` (TIMESTAMP) - Expiration time for reset token

## Security Features

1. **Password Hashing** - Using bcrypt with default cost
2. **JWT Tokens** - 24-hour expiration by default
3. **Password Reset Tokens** - 1-hour expiration for security
4. **Input Validation** - Required field validation on all endpoints
5. **Middleware Protection** - JWT middleware for protected endpoints

## Implementation Details

### File Structure
```
auth/
  ├── jwt.go           - JWT token generation and validation
  ├── password.go      - Password hashing and comparison
  ├── email.go         - Password reset token generation and email sending
  └── middleware.go    - JWT middleware for route protection

handlers/
  ├── auth_handler.go  - Authentication endpoint handlers
  └── user_handler.go  - (existing) User endpoint handlers

models/
  └── user.go          - (updated) User model with auth methods

controllers/
  └── user_controller.go - (updated) User controller with auth routes
```

### Key Methods

#### UserModel Methods
- `CreateWithPassword()` - Create user with hashed password
- `GetByEmail()` - Get user by email
- `GetByUsername()` - Get user by username
- `UpdatePassword()` - Update user password
- `SetPasswordResetToken()` - Set password reset token
- `GetByResetToken()` - Get user by reset token
- `ClearPasswordResetToken()` - Clear reset token after successful reset

#### Auth Utilities
- `HashPassword()` - Hash password using bcrypt
- `ComparePassword()` - Compare hashed password with plain text
- `GenerateToken()` - Generate JWT token
- `ValidateToken()` - Validate JWT token
- `GenerateResetToken()` - Generate secure reset token
- `SendPasswordResetEmail()` - Send password reset email (placeholder)

## Testing the API

### 1. Register a new user
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "testpass123"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpass123"
  }'
```

### 3. Access protected endpoint with token
```bash
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer <your_token_here>"
```

### 4. Forgot Password
```bash
curl -X POST http://localhost:8080/api/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```

### 5. Reset Password
```bash
curl -X POST http://localhost:8080/api/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "token_from_email",
    "new_password": "newpass123"
  }'
```

## Production Considerations

1. **JWT_SECRET** - Change to a strong random key
2. **Email Service** - Replace `SendPasswordResetEmail()` with actual email service (SendGrid, Mailgun, AWS SES, etc.)
3. **HTTPS** - Use HTTPS in production
4. **Token Expiration** - Adjust token expiration time as needed (currently 24 hours)
5. **Rate Limiting** - Consider adding rate limiting to auth endpoints
6. **CORS** - Configure CORS if frontend is on different domain
7. **Password Policy** - Consider adding password strength validation

## Next Steps

To further enhance the authentication system:

1. Add email verification for new registrations
2. Implement refresh tokens for longer sessions
3. Add 2FA (Two-Factor Authentication)
4. Implement password strength validation
5. Add login attempt tracking and account lockout
6. Add audit logging for authentication events
