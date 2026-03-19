# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the application
go run main.go                    # Starts server on PORT (default 8080)

# Build
go build -o main .

# Run tests
go test -v ./...
go test -v ./tests/...            # Run only tests in tests/ directory
go test -v -run TestFunctionName ./tests/   # Run a single test

# Database migration
go run cmd/migrate/main.go        # Runs all .sql files in migrations/ in order

# Regenerate SQLC code (after modifying query/*.sql)
sqlc generate

# Docker
docker-compose up --build         # Runs on port 8081
```

## Architecture

**Layered architecture:**
```
HTTP → Middleware → Controllers (route registration) → Handlers (HTTP logic) → Models (business logic) → SQLC DB layer → PostgreSQL
```

**Key layer responsibilities:**
- `controllers/`: Route registration using Gorilla Mux. Each controller receives a `*pgxpool.Pool` and wires handlers to routes.
- `handlers/`: HTTP request/response logic. All handlers use `WriteSuccess()` / `WriteError()` from `handlers/response.go`. Interfaces for all handler dependencies are defined in `handlers/interfaces.go`.
- `models/`: Business logic that implements the interfaces in `handlers/interfaces.go`. Calls SQLC-generated code.
- `internal/db/`: Auto-generated SQLC code — **do not edit manually**. Regenerate with `sqlc generate` after changing `query/*.sql`.
- `auth/`: JWT generation/validation (`jwt.go`), bcrypt password hashing (`password.go`), JWT middleware (`middleware.go`).
- `middleware/`: CORS and request logging (applied globally in `main.go`).

**Dependency injection flow:** `main.go` creates a `pgxpool.Pool` via `config.LoadConfig()`, passes it to each controller, which instantiates models and handlers.

**Database:** PostgreSQL via `pgx/v5`. Connection pooling with `pgxpool`. Schema managed by numbered SQL migration files in `migrations/` (run in alphabetical order by `cmd/migrate/main.go`).

**SQLC workflow:** SQL queries live in `query/*.sql`. Running `sqlc generate` produces type-safe Go code in `internal/db/`. The `db.DBTX` interface allows both `pgxpool.Pool` and `pgx.Tx` to be used interchangeably.

**Testing:** Tests in `tests/` use mock implementations of the interfaces from `handlers/interfaces.go`. Mocks are defined in `tests/mocks.go`.

**Auth:** JWT middleware (`auth.JWTMiddleware()`) is applied per-route in controllers. Public routes: `/api/auth/register`, `/api/auth/login`, password reset endpoints.

**API response format:**
```json
{ "msgId": "<uuid>", "status": "success|error", "data": {} }
```

**Swagger docs** are served at `/swagger/` and are generated in the `docs/` directory. Not served in production.

## Configuration

Environment variables loaded from `.env` (or system env):
- `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`
- `PORT` — HTTP server port (default `8080`)

SSL is enabled automatically when `DB_HOST` is not `localhost`. IPv4 is forced when running in Docker.
