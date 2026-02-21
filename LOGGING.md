# Logging System Documentation

## Overview

The Expense Tracker application includes a comprehensive logging system to trace all HTTP requests and responses, helping with debugging, monitoring, and troubleshooting.

## Features

✅ **Request Logging** - Logs all incoming requests with method, path, and query parameters  
✅ **Response Logging** - Logs all outgoing responses with status code and duration  
✅ **Colored Output** - Color-coded status codes for easy visibility  
✅ **Structured Logging** - Multiple log levels (DEBUG, INFO, WARN, ERROR, FATAL)  
✅ **Performance Metrics** - Request duration tracking in milliseconds  
✅ **Security** - Sensitive data (tokens) are masked  

## Log Levels

| Level | Purpose | Example |
|-------|---------|---------|
| **DEBUG** | Detailed diagnostic information | Request headers, verbose tracing |
| **INFO** | General informational messages | Request/response pairs, startup |
| **WARN** | Warning messages about potential issues | Deprecated usage, configuration issues |
| **ERROR** | Error messages for failed operations | Failed database queries, validation errors |
| **FATAL** | Critical errors that stop the application | Database connection failures |

## Log Output Format

### Request Logs
```
[2026-02-21 10:15:32] INFO - [REQUEST] POST /api/auth/login | Remote: 127.0.0.1:54321 | Query: 
[2026-02-21 10:15:32] DEBUG -   Content-Type: application/json
[2026-02-21 10:15:32] DEBUG -   Authorization: Bearer [token present]
```

### Response Logs
```
[2026-02-21 10:15:32] INFO - [RESPONSE] POST /api/auth/login | Status: 200 | Duration: 45ms
[2026-02-21 10:15:32] INFO - [RESPONSE] GET /api/users | Status: 401 | Duration: 12ms
[2026-02-21 10:15:32] INFO - [RESPONSE] PUT /api/users/1 | Status: 500 | Duration: 278ms
```

## Status Code Colors

The logging middleware uses ANSI color codes for visual distinction:

- **Green** (2xx) - Successful responses
- **Cyan** (3xx) - Redirects
- **Yellow** (4xx) - Client errors (bad request, unauthorized, etc.)
- **Red** (5xx) - Server errors

## Usage Examples

### Using Logger in Code

```go
package handlers

import "github.com/yanaatere/expense_tracking/logger"

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    logger.Infof("User login attempt from IP: %s", r.RemoteAddr)
    
    // Your handler logic...
    
    if err != nil {
        logger.Errorf("Login failed for user: %v", err)
        http.Error(w, "Login failed", http.StatusInternalServerError)
        return
    }
    
    logger.Infof("User successfully logged in")
}
```

### Log Levels in Code

```go
// Debug level - for detailed diagnostic information
logger.Debug("Processing request")
logger.Debugf("User ID: %d, Email: %s", userID, email)

// Info level - for general informational messages
logger.Info("Server started successfully")
logger.Infof("Database connected to %s", dbName)

// Warn level - for potential issues
logger.Warn("Configuration incomplete, using default values")
logger.Warnf("Deprecated endpoint accessed: %s", endpoint)

// Error level - for errors that don't stop execution
logger.Error("Failed to send email")
logger.Errorf("Database query failed: %v", err)

// Fatal level - for critical errors that stop execution
logger.Fatal("Cannot connect to database")
logger.Fatalf("Missing required config: %s", configKey)
```

## Middleware Integration

The logging middleware is automatically applied to all HTTP requests in the middleware stack:

```go
// In main.go
handler := middleware.LoggingMiddleware(r)  // Applied first (innermost)
handler = middleware.CORSMiddleware(handler) // Applied second (outermost)
```

**Middleware Order (Important):**
1. **LoggingMiddleware** - Captures all requests/responses
2. **CORSMiddleware** - Handles cross-origin requests

## Request Tracing Flow

```
Client Request
    ↓
LoggingMiddleware (logs request details)
    ↓
CORSMiddleware (adds CORS headers)
    ↓
Router & Handler (processes request)
    ↓
CORSMiddleware (returns with CORS headers)
    ↓
LoggingMiddleware (logs response details)
    ↓
Client Response
```

## What Gets Logged

### For Every Request:
- ✅ HTTP Method (GET, POST, PUT, DELETE, etc.)
- ✅ Request URI/Path
- ✅ Query Parameters
- ✅ Remote IP Address
- ✅ Content-Type Header
- ✅ Authorization Header (masked for security)

### For Every Response:
- ✅ HTTP Status Code
- ✅ Response Duration (milliseconds)
- ✅ Color-coded visual indicator

## Security Considerations

### Sensitive Data Protection
The logging middleware **intentionally masks sensitive information**:

```
Authorization: Bearer [token present]
```

Instead of logging the actual JWT token.

### What's NOT Logged:
- ❌ Request body content (passwords, personal data)
- ❌ Response body content (sensitive user data)
- ❌ Actual JWT token values
- ❌ Database credentials

This prevents accidentally logging sensitive information to stdout or log files.

## Sampling Logs

### Example Login Request/Response Sequence

```bash
$ ./expense_tracker

======================== Server Starting ========================
[2026-02-21 10:15:20] INFO - Server starting on port 8080
[2026-02-21 10:15:20] INFO - Environment: development
[2026-02-21 10:15:20] INFO - Database: expensetracker
===========================================================

# User makes request
[2026-02-21 10:15:32] INFO - [REQUEST] POST /api/auth/login | Remote: 127.0.0.1:54321 | Query: 
[2026-02-21 10:15:32] DEBUG -   Content-Type: application/json
[2026-02-21 10:15:32] DEBUG -   Authorization: Bearer [token present]
[2026-02-21 10:15:32] INFO - [RESPONSE] POST /api/auth/login | Status: 200 | Duration: 45ms

# Get users request
[2026-02-21 10:15:35] INFO - [REQUEST] GET /api/users | Remote: 127.0.0.1:54321 | Query: 
[2026-02-21 10:15:35] DEBUG -   Authorization: Bearer [token present]
[2026-02-21 10:15:35] INFO - [RESPONSE] GET /api/users | Status: 200 | Duration: 23ms

# Failed request
[2026-02-21 10:15:40] INFO - [REQUEST] GET /api/invalid | Remote: 127.0.0.1:54321 | Query: 
[2026-02-21 10:15:40] INFO - [RESPONSE] GET /api/invalid | Status: 404 | Duration: 1ms
```

## Performance Impact

The logging middleware has minimal performance impact:
- Response duration includes all processing and logging overhead
- Logging is done synchronously after the response
- No database queries or external calls are made during logging

## Troubleshooting

### Issue: Not seeing logs

**Solution:** Make sure the log level is set appropriately. Default is INFO level, so DEBUG messages won't show.

```go
// Set log level to DEBUG for more verbose output
logger.SetLevel(logger.DEBUG)
```

### Issue: Logs not showing in Docker

**Solution:** Go logs write to stdout by default. Make sure Docker container is run with:

```bash
docker run -it your-image  # -it flags ensure logs are visible
```

### Issue: Too much logging output

**Solution:** If you have too many requests, you can set log level to WARN to reduce output:

```go
// In main.go before handlers
logger.SetLevel(logger.WARN)
```

## Best Practices

1. **Use appropriate log levels** - Don't log everything at INFO level
2. **Avoid logging sensitive data** - Passwords, tokens, SSNs, etc.
3. **Include context** - Log user IDs, request IDs for tracing
4. **Structured information** - Use Infof/Errorf for formatted messages
5. **Errors should be logged** - Always log errors before returning them to users

## File Structure

```
logger/
  └── logger.go        - Logger implementation with log levels

middleware/
  ├── cors.go          - CORS middleware
  └── logging.go       - Request/response logging middleware

main.go               - Integrated logging into server startup
```

## Future Enhancements

Potential improvements to the logging system:

- Log file rotation (daily/size-based)
- Structured JSON logging for log aggregation services
- Request ID tracking across log entries
- Performance metrics and histogram tracking
- Integration with external logging services (ELK, Splunk, etc.)
- Rate limiting for log output
- Configurable log output format
