# Quickstart: Child Visual Themes

**Feature**: 017-child-visual-themes
**Date**: 2026-02-18

## Prerequisites

- PostgreSQL 17 running locally with `bankofdad` and `bankofdad_test` databases
- Go 1.24+
- Node.js 18+ with npm
- Existing development environment set up (backend builds, frontend dev server works)

## Setup Steps

### 1. Apply Migration

```bash
cd backend
# Run migration 004 against dev database
migrate -path migrations -database "postgres://localhost:5432/bankofdad?sslmode=disable" up

# Run migration 004 against test database
migrate -path migrations -database "postgres://localhost:5432/bankofdad_test?sslmode=disable" up
```

### 2. Backend Changes

Files to modify:
- `backend/migrations/004_child_theme.up.sql` — new migration
- `backend/migrations/004_child_theme.down.sql` — rollback migration
- `backend/internal/store/child.go` — add `Theme` field, update SELECTs/Scans, add `UpdateTheme`
- `backend/internal/auth/handlers.go` — include `theme` in child `/auth/me` response
- `backend/internal/family/handlers.go` — add `HandleUpdateTheme` handler (or new child settings handler file)
- `backend/main.go` — register new route

### 3. Frontend Changes

New files:
- `frontend/src/pages/ChildSettingsPage.tsx` — child settings page
- `frontend/src/context/ThemeContext.tsx` — theme provider context
- `frontend/src/themes.ts` — theme definitions (colors, SVG patterns)

Files to modify:
- `frontend/src/App.tsx` — add `/child/settings` route, wrap with `ThemeProvider`
- `frontend/src/components/Layout.tsx` — add "Settings" nav item for child users
- `frontend/src/types.ts` — add `theme` to `ChildUser` interface
- `frontend/src/api.ts` — add `updateChildTheme` function
- `frontend/src/index.css` — no changes needed (CSS custom properties are overridden at runtime)

### 4. Run Tests

```bash
# Backend
cd backend && go test -p 1 ./...

# Frontend
cd frontend && npm test
```

### 5. Verify Manually

1. Log in as a parent, create a child if needed
2. Log in as the child
3. Navigate to Settings (new nav item)
4. Select each theme and verify:
   - Page background color changes
   - Accent/heading colors change
   - Subtle SVG pattern appears in background
5. Log out and back in — theme persists
6. Log in as parent — verify standard appearance (no theme applied)

## Architecture Summary

```
Child logs in
  → GET /auth/me returns { theme: "rainbow" }
  → ThemeProvider reads theme, applies CSS custom property overrides
  → All child pages render with themed colors + background SVG

Child changes theme in settings
  → PUT /api/child/settings/theme { theme: "piggybank" }
  → ThemeProvider updates CSS custom properties immediately
  → Theme persists in DB for next login
```
