# bank-of-dad Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-01-29

## Active Technologies
- Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `modernc.org/sqlite`, `testify`, react-router-dom, Vite (002-account-balances)
- SQLite with WAL mode (existing pattern - separate read/write connections) (002-account-balances)
- Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `modernc.org/sqlite`, `testify`, react-router-dom, Vite (existing - no new deps) (003-allowance-scheduling)
- Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + golangci-lint (new, backend CI only), ESLint + typescript-eslint (new, frontend CI only), GitHub Actions runners (004-ci-github-actions)

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
- 004-ci-github-actions: Added Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + golangci-lint (new, backend CI only), ESLint + typescript-eslint (new, frontend CI only), GitHub Actions runners
- 003-allowance-scheduling: Added Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `modernc.org/sqlite`, `testify`, react-router-dom, Vite (existing - no new deps)
- 002-account-balances: Added Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `modernc.org/sqlite`, `testify`, react-router-dom, Vite


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
