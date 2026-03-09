# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go REST API for a hymn management system (Himnario), built with Gin and hexagonal architecture. Uses a custom microservice framework (`lightms`) for initialization.

**Go version:** 1.25.5

## Build & Development Commands

```bash
make run            # Run in development (go run ./cmd/main.go)
make build          # Compile to bin/app (with -tags musl)
make exec           # Run compiled binary
make test           # Run tests with HTML coverage report
make race-test      # Run tests with race condition detection
make lint           # Run golangci-lint
make lint-fast      # Fast lint (pre-commit)
make check-all      # test + race-test + lint
make pre-commit     # clean + lint + build + test
make profile        # Benchmark with memory profiling
make clean-mod      # Clear Go module cache
```

**Database:** Start PostgreSQL with `docker compose -f docker/docker-compose.yml up -d`. Connects on `localhost:5432`, database `hymns_db`, user `hymns_user`, password `hymns_pass`.

Migrations run automatically on startup via golang-migrate from `resources/migrations/`.

## Key Dependencies

- `github.com/gin-gonic/gin v1.11.0` — HTTP framework
- `github.com/jackc/pgx/v5 v5.8.0` — PostgreSQL driver and connection pool
- `github.com/golang-migrate/migrate/v4 v4.19.1` — Database migrations
- `github.com/go-playground/validator/v10 v10.30.1` — Struct validation
- `github.com/golang-jwt/jwt/v5 v5.3.1` — JWT tokens
- `golang.org/x/crypto v0.48.0` — bcrypt hashing
- `github.com/JGLTechnologies/gin-rate-limit v1.5.6` — Rate limiting
- `github.com/gin-contrib/size v1.0.2` — Request size limiting
- `github.com/gin-contrib/timeout v1.1.0` — Request timeouts
- `github.com/lib/pq v1.11.2` — PostgreSQL driver (used by golang-migrate)
- `github.com/goccy/go-yaml v1.18.0` — YAML config parsing
- `go.uber.org/mock v0.5.0` — Mocking for tests

## Architecture

Hexagonal (ports & adapters) with three layers:

- **`internal/core/`** — Domain logic, independent of infrastructure
  - `domain/` — Entities (`hymns.go`, `user.go`), errors (`errors.go`), response wrappers (`responses.go`)
  - `port/` — Interfaces defining persistence contracts (`HymnsPersistence`, `UserPersistence`)
  - `service/` — Business logic singletons (`HymnsService`, `UserService`) implementing use cases from `usecase.go`

- **`internal/infra/primary/`** — Inbound adapters (HTTP controllers)
  - Each controller has: `controller.go` (handlers), `mapper.go` (DTO→domain), `request/dto.go`, `validations/`
  - Controllers: `user/` (create, authenticate) and `hymns/` (create, list all, get by ID)

- **`internal/infra/secondary/`** — Outbound adapters (PostgreSQL repositories)
  - `users/client.go` and `hymnary/client.go` implement port interfaces
  - `users/responses/user.go` — DB scan targets for user queries

- **`internal/infra/config/`** — Cross-cutting setup
  - `controller/process.go` — Gin engine, middleware chain, server config
  - `database/` — pgx connection pool (`postgresql_connection.go`) and migration runner (`migrations.go`)
  - `instance/primary.go` — Dependency injection wiring
  - `property/` — Typed config structs: `server.go`, `database.go`, `jwt.go`, `validations.go`
  - `security/` — JWT (`jwt.go`, HS256, 60min expiry) and bcrypt (`hash.go`, cost=12) utilities
  - `validation/` — go-playground/validator singleton (`validator.go`, `validatable.go`)

## Entry Point Flow

`cmd/main.go` → registers property types and controllers → `lightms.Run()`:
1. Loads `resources/properties.yml`
2. Registers: server, database, JWT, validations properties
3. Initializes DI (`instance/primary.go`): DB pool → repositories → services → controllers
4. Builds Gin engine with middleware chain
5. Starts HTTP server on port 8080

## API Endpoints

All requests require these headers:
- `X-Client`: `mbh` or `front-end-himnary`
- `country`: `CL`
- `event-id`: any non-empty string
- `Content-Type`: `application/json`
- `Accept`: `application/json`

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/user/create` | — | Register user (email, password ≥8 chars, name, role: admin\|editor\|viewer) |
| `POST` | `/user/authentication` | — | Login, returns JWT token + user data |
| `POST` | `/hymn/create` | JWT (admin\|editor) | Create hymn; validates identity match between token and body |
| `GET` | `/hymn/all` | — | List all hymns (returns id, title, created_at — no content) |
| `GET` | `/hymn/:id` | — | Get hymn by ID (returns id, title, content JSONB, created_at) |
| `GET` | `/health` | — | Health check |

Allowed paths must be registered in `properties.yml` under `server.allowed-paths`; any unlisted path returns 403.

## Database Schema

**Tables:**
- `users` — accounts with roles (admin, editor, viewer); email unique, lowercase normalized
- `hymns` — records with `content` as JSONB; `title` unique (case-insensitive); GIN index on content
- `categories` — hymn categories
- `hymn_categories` — many-to-many: hymns ↔ categories
- `favorites` — user↔hymn favorites tracking
- `setlists` — collections of hymns for events
- `setlist_hymns` — ordered hymn membership in setlists

**Indexes:**
- `idx_hymns_title` — `ON hymns(title)`
- `idx_hymns_content_gin` — `ON hymns USING GIN(content)`

**Stored Procedures:**
- `sp_create_user(email, password_hash, name, role)` — creates user, raises if email duplicate
- `sp_get_user_auth_by_email(email)` — fetches user record for login
- `sp_create_hymn_by_author_email(title, content, author_email)` — creates hymn, validates title uniqueness
- `sp_list_hymns()` — returns all hymns ordered by title (id, title, created_at only)
- `sp_get_hymn_by_id(id)` — returns full hymn with content JSONB

## Middleware Chain (execution order)

1. **Logging** — custom formatter including JSON request body
2. **Timeout** — 30s per request; returns custom timeout response
3. **Allowed Paths** — rejects any path not in `properties.yml` allowed list
4. **Request Size** — rejects bodies larger than 1 MB
5. **Rate Limiting** — 100 requests/min per client IP (in-memory store)
6. **Recovery** — panic recovery
7. **Security Headers** — CORS, CSP, XSS, HSTS, Referrer-Policy, Permissions-Policy, X-Frame-Options, X-Content-Type-Options
8. **Header Validations** — validates `country`, `x-client`, `Content-Type`, `Accept`, `event-id`
9. **JWT Validation** — applied per-route (currently: POST `/hymn/create`)

## Security

**Password hashing:** bcrypt, cost=12

**JWT:**
- Algorithm: HS256
- Expiry: 60 minutes
- Claims include: user email, name, role, issued-at, expiration, issuer

**Security Headers applied on every response:**
- `X-Frame-Options: DENY`
- `Content-Security-Policy` (restricted; self-only with limited exceptions)
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains; preload`
- `Referrer-Policy: strict-origin`
- `X-Content-Type-Options: nosniff`
- `Permissions-Policy` (geolocation, midi, microphone, camera, etc. all disabled)

## Response Wrapper Pattern

All HTTP responses use a `ResponseWrapper` struct containing either `ResponseOk` or `ResponseError` (never both):

- **`ResponseOk`** — `{ status, message, data }`
- **`ResponseError`** — `{ status, error, fields[] }` (field-level validation errors when applicable)

## Key Patterns

- **Singletons with `sync.Once`** for services, DB pool, JWT properties, server properties, and validator
- **Stored procedures** for all DB writes and complex reads (no raw SQL in repository layer)
- **pgx context timeouts** — each DB query runs with a 5-second context deadline
- **Role-based access control** — JWT claims carry user role; controllers verify role and identity (email + name) before write operations
- **Email normalization** — lowercased and trimmed at multiple layers (DTO mapper, DB procedure)
- **Config-driven** — all server, database, JWT, and validation settings in `resources/properties.yml`

## Docker

**`docker/docker-compose.yml`:**
- Service: `hymns-postgres` (PostgreSQL 16 Alpine)
- Port: `5432`
- Volume: `hymns_pgdata` for data persistence
- Health check: `pg_isready` every 5 seconds
- Init scripts: `docker/postgres/initdb/00-settings.sql` (sets timezone to UTC)

**`Dockerfile`:** builds the Go binary for production (uses `-tags musl` for static linking)

## Database Connection Pool (pgx)

Configured via `properties.yml`:
- Max connections: 10
- Min connections: 1
- Max connection lifetime: 30s
- Max connection idle time: 30s
- Health check interval: 5s
- Statement cache capacity: 256
