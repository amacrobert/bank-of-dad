# bank-of-dad Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-01-29

## Active Technologies
- Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `modernc.org/sqlite`, `testify`, react-router-dom, Vite (002-account-balances)
- SQLite with WAL mode (existing pattern - separate read/write connections) (002-account-balances)
- Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `modernc.org/sqlite`, `testify`, react-router-dom, Vite (existing - no new deps) (003-allowance-scheduling)
- Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + golangci-lint (new, backend CI only), ESLint + typescript-eslint (new, frontend CI only), GitHub Actions runners (004-ci-github-actions)
- SQLite with WAL mode (separate read/write connections) (005-interest-accrual)
- Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `modernc.org/sqlite`, `testify`, react-router-dom, Vite, lucide-reac (010-child-avatars)
- Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend — unchanged) + `jackc/pgx/v5` + `pgx/v5/stdlib` (new), `golang-migrate/migrate/v4` (new), remove `modernc.org/sqlite` (011-sqlite-to-postgres)
- PostgreSQL 17 (replacing SQLite) (011-sqlite-to-postgres)
- Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `golang-jwt/jwt/v5` (new), `jackc/pgx/v5`, `testify`, react-router-dom, Vite (012-stateless-auth)
- PostgreSQL 17 — new `refresh_tokens` table, drop `sessions` table (012-stateless-auth)
- TypeScript 5.3.3 + React 18.2.0 + react-router-dom, Vite, lucide-react (icons) (014-combine-transactions)
- Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `jackc/pgx/v5`, `testify` (backend); `react-router-dom`, `lucide-react`, Vite (frontend) (013-parent-settings)
- PostgreSQL 17 — add `timezone` column to existing `families` table (013-parent-settings)
- PostgreSQL 17 — no schema changes; existing `families.timezone` column used (015-timezone-aware-scheduling)
- TypeScript 5.3.3 + React 18.2.0 (frontend only — no backend changes) + Recharts (new), react-router-dom, lucide-react, Vite (existing) (016-savings-growth-projector)
- N/A — no persistence; all projections are computed client-side from existing API data (016-savings-growth-projector)
- Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `jackc/pgx/v5`, `testify` (backend); react-router-dom, lucide-react, Vite (frontend) (017-child-visual-themes)
- PostgreSQL 17 — add `theme TEXT` column to existing `children` table (017-child-visual-themes)
- TypeScript 5.3.3 + React 18.2.0 + react-router-dom, lucide-react, Vite (018-child-management-settings)
- N/A — no schema changes; existing PostgreSQL backend + REST API unchanged (018-child-management-settings)

- Go 1.21 (backend), TypeScript 5.3 + React 18.2 (frontend) + `golang.org/x/oauth2` (Google OAuth), `golang.org/x/crypto/bcrypt` (password hashing), `modernc.org/sqlite` (database), `react-router-dom` (frontend routing), `testify` (Go test assertions) (001-user-auth)

## Project Structure

```text
src/
tests/
```

## Commands

npm test && npm run lint

## Code Style

Go 1.21 (backend), TypeScript 5.3 + React 18.2 (frontend): Follow standard conventions

## Recent Changes
- 018-child-management-settings: Added TypeScript 5.3.3 + React 18.2.0 + react-router-dom, lucide-react, Vite
- 017-child-visual-themes: Added Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `jackc/pgx/v5`, `testify` (backend); react-router-dom, lucide-react, Vite (frontend)
- 016-savings-growth-projector: Added TypeScript 5.3.3 + React 18.2.0 (frontend only — no backend changes) + Recharts (new), react-router-dom, lucide-react, Vite (existing)


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
