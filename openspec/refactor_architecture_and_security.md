# Refactoring Proposal: Architectural Improvements and Security Hardening

**Date:** December 6, 2025
**Status:** Proposed
**Author:** Gemini Agent

## 1. Objective
To improve the maintainability, security, and cleanliness of the `nofx` codebase by addressing technical debt identified in the architectural audit.

## 2. Key Changes

### 2.1 Root Directory Cleanup
**Problem:** The root directory is cluttered with documentation, scripts, and SQL files, making navigation difficult.
**Solution:**
- Move `*_REPORT.md`, `*.md` (except `README.md`, `LICENSE`, `go.mod`, `go.sum`, etc.) to `docs/` or `docs/reports/`.
- Move `*.sh` to `scripts/`.
- Move `*.sql` to `database/migrations/` (verifying code references first).

### 2.2 Security Hardening (JWT)
**Problem:** The application falls back to a hardcoded insecure JWT secret if one is not provided.
**Solution:**
- Remove the hardcoded fallback string in `main.go`.
- Enforce application panic/exit if `JWT_SECRET` is missing in production/non-dev environments, or generate a random one for ephemeral sessions (though enforcing config is safer).

### 2.3 Refactoring `api/server.go`
**Problem:** `api/server.go` is a "God Class" (~2400 lines) handling routing, logic, and data access.
**Solution:**
- Extract handlers into a new package `api/handlers`.
- Group handlers by domain:
    - `api/handlers/auth.go`: Login, Register, Password Reset.
    - `api/handlers/trader.go`: Trader management (CRUD), actions (start/stop).
    - `api/handlers/market.go`: Market data, public info.
    - `api/handlers/config.go`: System configs, models, exchanges.
- `api/server.go` will retain responsibility only for router setup and middleware wiring.

### 2.4 Database Layer Improvement
**Problem:** Fragile manual string replacement (`strings.ReplaceAll(query, "$1", "?")`) for SQLite compatibility.
**Solution:**
- Introduce a simpler, safer query binding mechanism or helper method within the `Database` struct to handle dialect differences, paving the way for a future migration to a query builder like `squirrel` or an ORM like `GORM`. *Immediate step: Centralize the query reformatting logic to prevent ad-hoc replacements throughout the code.*

### 2.5 Testing Strategy
**Problem:** Lack of unit tests for HTTP handlers.
**Solution:**
- Add `web/package.json` test script.
- Add sample unit tests for the new `api/handlers` package.

### 2.6 Switch to pgx Driver
**Problem:** The `lib/pq` driver has issues with prepared statements in certain environments (like Neon/PgBouncer), causing "bind message supplies X parameters, but prepared statement requires 1" errors.
**Solution:**
- Switch to `github.com/jackc/pgx/v5/stdlib` which is actively maintained and handles connection pooling/prepared statements more robustly.

## 3. Implementation Plan
1. **Cleanup:** Execute file moves and update paths in `main.go` or scripts if necessary.
2. **Security:** Modify `main.go` to validate JWT secret presence.
3. **Refactor API:** Create `api/handlers`, move code, update `api/server.go`.
4. **Database:** Refactor `database/database.go` to use a centralized `Rebind` method.
5. **Driver Upgrade:** Replace `lib/pq` with `pgx` in `config/database.go`, `database/database.go`, and `database/migrate.go`.
6. **Verify:** Compile and run tests.
