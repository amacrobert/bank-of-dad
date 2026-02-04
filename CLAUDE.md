# bank-of-dad Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-01-29

## Active Technologies
- Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `modernc.org/sqlite`, `testify`, react-router-dom, Vite (002-account-balances)
- SQLite with WAL mode (existing pattern - separate read/write connections) (002-account-balances)

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
- 002-account-balances: Added Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `modernc.org/sqlite`, `testify`, react-router-dom, Vite

- 001-user-auth: Added Go 1.21 (backend), TypeScript 5.3 + React 18.2 (frontend) + `golang.org/x/oauth2` (Google OAuth), `golang.org/x/crypto/bcrypt` (password hashing), `modernc.org/sqlite` (database), `react-router-dom` (frontend routing), `testify` (Go test assertions)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
