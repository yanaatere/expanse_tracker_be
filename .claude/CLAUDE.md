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

**Dependency injection flow:** `main.go` creates a `pgxpool.Pool` via `config.LoadConfig()`, passes it to each controller, which instantiates models and handlers. The server uses `http.Server` with `ReadTimeout=10s`, `WriteTimeout=30s`, `IdleTimeout=60s`.

**Database:** PostgreSQL via `pgx/v5`. Connection pooling with `pgxpool`. Schema managed by numbered SQL migration files in `migrations/` (run in alphabetical order by `cmd/migrate/main.go`).

**SQLC workflow:** SQL queries live in `query/*.sql`. Running `sqlc generate` produces type-safe Go code in `internal/db/`. The `db.DBTX` interface allows both `pgxpool.Pool` and `pgx.Tx` to be used interchangeably.

**Testing:** Tests in `tests/` use mock implementations of the interfaces from `handlers/interfaces.go`. Mocks are defined in `tests/mocks.go`. Note: `tests/balance_handler_test.go` has pre-existing broken tests for `GetBalanceByCategory` / `models.CategoryBalance` which are not yet implemented — do not be alarmed by the compile failure in that file.

**Auth:** JWT middleware (`auth.JWTMiddleware()`) is applied per-route in controllers. Public routes: `/api/auth/register`, `/api/auth/login`, password reset endpoints. The JWT context key is an unexported typed `contextKey` string — always use `auth.GetUserIDFromContext(ctx)` to retrieve it, never access the context key directly. `JWT_SECRET` env var **must** be set — the app panics on startup if it is missing. bcrypt cost is 12.

**Security invariants to maintain:**
- All protected handlers extract `userID` from JWT context (`auth.GetUserIDFromContext`) — never from query params or body.
- User-scoped resources (transactions, wallets, balances) pass `userID` to the model layer which enforces ownership via `WHERE user_id = $1`.
- User profile endpoints (`GET/PUT/DELETE /api/users/{id}`) enforce that the path `id` matches the JWT `userID` — return `403` otherwise.
- Never return raw `err.Error()` to clients — use generic messages like `"Failed to retrieve user"`.
- `Access-Control-Allow-Credentials` must **not** be set (Bearer token auth uses `Authorization` header, not cookies).
- Auth endpoints are rate-limited via `middleware.RateLimit` (Redis-backed, Lua atomic INCR+EXPIRE). Limits: login 10/min, register 5/hr, forgot-password 3/hr, reset-password 5/15min. Fails open on Redis outage.
- MinIO bucket (`receipts`) has **no public policy** — all receipt images are served via 1-hour pre-signed URLs. Use `GET /api/uploads/receipts/{objectName}` to get a fresh URL. `UploadReceipt` returns both `object_name` and an initial `url` (already pre-signed). Store `object_name` in the DB, not the URL.
- `NewUploadController` no longer takes `minioPublicURL` — removed since direct public URLs are no longer issued.
- `NewUserController` takes `(db.DBTX, *redis.Client)` — Redis is required for rate limiting.

**API response format:**
```json
{ "msgId": "<uuid>", "status": "success|error", "data": {} }
```

**Swagger docs** are served at `/swagger/` and are generated in the `docs/` directory. Only enabled when `ENVIRONMENT` != `production`.

**File uploads:** Receipt images are uploaded to MinIO object storage (S3-compatible) and served via `MINIO_PUBLIC_URL`. The `upload_controller.go` handles multipart form uploads to the `receipts` bucket.

**Bot integration:** `handlers/bot_handler.go` + `controllers/bot_controller.go` implement Telegram bot linking via Redis. A 6-digit `link_code` (stored as `link_code:<code>` in Redis) maps to a Telegram `chatID`. Consuming the code writes the user's JWT and ID into the `session:<chatID>` Redis key. The `BotHandler` takes a `*redis.Client` directly (not an interface) — it is not covered by the mock-based test pattern.

**Google Sign-In:** `POST /api/auth/google` accepts a Google `id_token`, verifies it via Google's `tokeninfo` endpoint (no extra SDK), and returns a Monex JWT. Auto-creates a new account if the email is not registered (password stored as empty string). Optional audience check if `GOOGLE_CLIENT_ID` env var is set. Route is public (no JWT middleware).

**Recurring transactions:** `recurring_transactions` table stores scheduled transactions with a `frequency` (daily/weekly/monthly) and optional `end_date`. A background goroutine in `main.go` calls `RecurringTransactionModel.ProcessDue()` once at startup then every 24 hours — it creates real transactions for all due entries and advances their `next_run_date`. `RecurringTransactionModel` requires a `*pgxpool.Pool` (not `db.DBTX`) because `ProcessDue` spawns a `pgx.Tx` internally.

**Dual balance tracking:** Two parallel systems maintain balance:
- `balances` table — one row per user, cumulative `total_balance` adjusted atomically on every transaction create/update/delete via `AdjustBalance` (INSERT ... ON CONFLICT DO UPDATE SET balance + delta).
- `wallets.balance` — per-wallet balance adjusted the same way via `AdjustWalletBalance`.
- `TransactionModel` coordinates both in a single `pgx.Tx` so they never drift. The `GetBalanceByDateRange` SQL query returns a `balance` column that is the **net of transactions within that date range** (not the stored total) — callers needing the true stored balance must call `GetBalance` separately.

**Timezone:** All date-range month-boundary calculations use `Asia/Jakarta` (WIB, UTC+7).

**Adding a new endpoint (5-step pattern):**
1. Add SQL query to `query/*.sql` → run `sqlc generate` to produce typed Go in `internal/db/`
2. Implement business logic in `models/` using the generated SQLC function
3. Add the method signature to the relevant interface in `handlers/interfaces.go`
4. Add the HTTP handler to `handlers/*_handler.go` using `auth.GetUserIDFromContext(r.Context())` + `WriteSuccess`/`WriteError`
5. Register the route in `controllers/*_controller.go` (with `auth.JWTMiddleware` if protected) and add the mock method to `tests/mocks.go`

## Development Guidelines

Before making **any code changes** (new features, bug fixes, refactors, etc.), always invoke the `golang-dev` skill first:

```
Use the Skill tool with skill: "golang-dev"
```

This skill is defined in `.claude/skills/golang-dev/SKILL.md` and provides Go-specific development guidelines and patterns to follow when implementing changes in this codebase.

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
- `GOOGLE_CLIENT_ID` — optional; if set, the Google Sign-In endpoint validates the token's `aud` claim against this value

SSL is enabled automatically when `DB_HOST` is not `localhost` or `db`. IPv4 is forced for remote DB hosts.
