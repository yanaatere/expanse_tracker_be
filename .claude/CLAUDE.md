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

# Regenerate Swagger docs (after modifying handler annotations)
swag init

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
- `auth/`: JWT generation/validation (`jwt.go`), bcrypt password hashing (`password.go`), JWT middleware (`middleware.go`), password reset token generation (`email.go` — email sending is a stub, integrate a real provider for production).
- `middleware/`: CORS and request logging (applied globally in `main.go`; logging wraps first, then CORS).
- `logger/`: Structured logger used throughout the app instead of `log` directly.

**Dependency injection flow:** `main.go` creates a `pgxpool.Pool` via `config.LoadConfig()`, passes it to each controller, which instantiates models and handlers.

**Database:** PostgreSQL via `pgx/v5`. Connection pooling with `pgxpool`. Schema managed by numbered SQL migration files in `migrations/` (run in alphabetical order by `cmd/migrate/main.go`).

**SQLC workflow:** SQL queries live in `query/*.sql`. Running `sqlc generate` produces type-safe Go code in `internal/db/`. The `db.DBTX` interface allows both `pgxpool.Pool` and `pgx.Tx` to be used interchangeably.

**Testing:** Tests in `tests/` use mock implementations of the interfaces from `handlers/interfaces.go`. Mocks are defined in `tests/mocks.go`.

**Auth:** JWT middleware (`auth.JWTMiddleware()`) is applied per-route in controllers. Public routes: `/api/auth/register`, `/api/auth/login`, password reset endpoints.

**API response format:**
```json
{ "msgId": "<uuid>", "status": "success|error", "data": {} }
```

**Swagger docs** are served at `/swagger/` and are generated in the `docs/` directory. Only enabled when `ENVIRONMENT` != `production`.

**File uploads:** Receipt images are uploaded to MinIO object storage (S3-compatible) and served via `MINIO_PUBLIC_URL`. The `upload_controller.go` handles multipart form uploads to the `receipts` bucket.

**Bot integration:** `handlers/bot_handler.go` + `controllers/bot_controller.go` implement Telegram bot linking via Redis. A 6-digit `link_code` (stored as `link_code:<code>` in Redis) maps to a Telegram `chatID`. Consuming the code writes the user's JWT and ID into the `session:<chatID>` Redis key. The `BotHandler` takes a `*redis.Client` directly (not an interface) — it is not covered by the mock-based test pattern.

## Development Guidelines

Before making **any code changes** (new features, bug fixes, refactors, etc.), always invoke the `golang-dev` skill first:

```
Use the Skill tool with skill: "golang-dev"
```

This skill provides Go-specific development guidelines and patterns to follow when implementing changes in this codebase.

## Configuration

Environment variables loaded from `.env` (or system env):
- `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`
- `PORT` — HTTP server port (default `8080`)
- `ENVIRONMENT` — set to `production` to disable Swagger UI
- `REDIS_URL` — Redis connection URL (default `redis://localhost:6379/0`); used for bot session linking
- `MINIO_ENDPOINT` — MinIO server address (default `localhost:9000`)
- `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY` — MinIO credentials (default `minioadmin`/`minioadmin`)
- `MINIO_USE_SSL` — set to `true` to enable TLS for MinIO (default `false`)
- `MINIO_PUBLIC_URL` — base URL for serving uploaded files (default `http://localhost:9000`)

SSL is enabled automatically when `DB_HOST` is not `localhost` or `db`. IPv4 is forced for remote DB hosts.
