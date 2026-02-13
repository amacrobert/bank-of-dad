# Quickstart: SQLite to PostgreSQL Migration

## Prerequisites

- Docker and Docker Compose
- Go 1.24+
- Node.js 20+ (frontend, unchanged)

## Local Development Setup

### 1. Start PostgreSQL

```bash
docker-compose up -d postgres
```

This starts PostgreSQL 17 on port 5432 with:
- User: `bankofdad`
- Password: `bankofdad`
- Database: `bankofdad`

### 2. Set environment variables

Copy the example env file:

```bash
cp .env.example .env
```

The default `DATABASE_URL` points to the local Docker Postgres instance:

```
DATABASE_URL=postgres://bankofdad:bankofdad@localhost:5432/bankofdad?sslmode=disable
```

### 3. Run the backend

```bash
cd backend
go run .
```

Migrations run automatically on startup. You should see the server start without errors.

### 4. Run tests

Tests require a running Postgres instance. The test database is separate from the dev database:

```bash
# Create the test database (one-time setup)
docker exec -it bank-of-dad-postgres-1 createdb -U bankofdad bankofdad_test

# Run tests
cd backend
TEST_DATABASE_URL="postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable" go test ./...
```

### 5. Full stack with Docker Compose

```bash
docker-compose up
```

This starts PostgreSQL, the backend, and the frontend. The backend waits for PostgreSQL to be healthy before starting.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://bankofdad:bankofdad@localhost:5432/bankofdad?sslmode=disable` |
| `TEST_DATABASE_URL` | PostgreSQL connection string for tests | `postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable` |

## Migration Commands

Migrations are embedded in the binary and run automatically on startup. No manual migration commands are needed for normal development.

To inspect the migration state:

```bash
# Check which migrations have been applied
docker exec -it bank-of-dad-postgres-1 psql -U bankofdad -c "SELECT * FROM schema_migrations;"
```

## Differences from SQLite Setup

| Before (SQLite) | After (PostgreSQL) |
|-----------------|-------------------|
| `DATABASE_PATH=bankodad.db` | `DATABASE_URL=postgres://...` |
| No external dependencies | Requires running Postgres (via Docker) |
| Two DB connections (read/write) | Single connection pool |
| Tests create temp `.db` files | Tests connect to `bankofdad_test` database |
| No test database setup needed | One-time `createdb` for test database |
