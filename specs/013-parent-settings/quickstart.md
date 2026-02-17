# Quickstart: Parent Settings Page

**Feature Branch**: `013-parent-settings`
**Date**: 2026-02-16

## Prerequisites

- Go 1.24+
- Node.js 18+
- PostgreSQL 17 running locally
- `bankofdad` database (dev) and `bankofdad_test` database (test) available

## Setup

### 1. Run the new migration

```bash
cd backend
migrate -path migrations -database "postgres://localhost:5432/bankofdad?sslmode=disable" up
```

For the test database:
```bash
migrate -path migrations -database "postgres://localhost:5432/bankofdad_test?sslmode=disable" up
```

### 2. Backend

```bash
cd backend
go test -p 1 ./...
```

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

## Key Files (to be created/modified)

### Backend
- `backend/migrations/003_family_timezone.up.sql` — Add timezone column to families
- `backend/migrations/003_family_timezone.down.sql` — Drop timezone column
- `backend/internal/store/family.go` — Add Timezone field to Family struct, add GetTimezone/UpdateTimezone methods
- `backend/internal/settings/handlers.go` — New settings handlers (GET settings, PUT timezone)
- `backend/internal/settings/handlers_test.go` — Handler tests
- `backend/internal/store/family_test.go` — Store-level tests for new methods
- `backend/main.go` — Register new routes

### Frontend
- `frontend/src/pages/SettingsPage.tsx` — New settings page with category navigation
- `frontend/src/components/TimezoneSelect.tsx` — Searchable timezone selector
- `frontend/src/components/Layout.tsx` — Add settings nav entry point
- `frontend/src/App.tsx` — Add `/settings` route
- `frontend/src/types.ts` — Add SettingsResponse type
- `frontend/src/api.ts` — Add getSettings/updateTimezone API functions

## Verification

1. Log in as a parent
2. Click the settings icon in the navigation
3. Verify the settings page loads with "General" category
4. Verify timezone defaults to "US Eastern Time (America/New_York)"
5. Change timezone to another value, save
6. Refresh the page — timezone should persist
7. Log in as a child — verify no settings link visible
