# Implementation Plan: Parent Settings Page

**Branch**: `013-parent-settings` | **Date**: 2026-02-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/013-parent-settings/spec.md`

## Summary

Add a parent-only settings page with extensible category-based navigation. The first implemented setting is family timezone selection (IANA identifiers, defaulting to `America/New_York`). Backend: new migration adding `timezone` column to `families`, new settings handlers. Frontend: new `/settings` route with category sidebar layout, searchable timezone selector, and nav entry point.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: `jackc/pgx/v5`, `testify` (backend); `react-router-dom`, `lucide-react`, Vite (frontend)
**Storage**: PostgreSQL 17 — add `timezone` column to existing `families` table
**Testing**: `go test -p 1 ./...` (backend), manual (frontend)
**Target Platform**: Web application (Docker, Docker Compose)
**Project Type**: Web (backend + frontend)
**Performance Goals**: Settings page interactive within 2 seconds
**Constraints**: Parent-only access, family-level timezone, IANA identifiers only
**Scale/Scope**: ~8 files modified/created, 2 new API endpoints

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-research check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | Plan includes store tests, handler tests, and contract tests for both endpoints. TDD cycle will be followed. |
| II. Security-First Design | PASS | Settings endpoints use `RequireParent` middleware. Input validated server-side via `time.LoadLocation()`. No sensitive data exposed. |
| III. Simplicity | PASS | Single column on existing table. No new tables, no abstractions. Curated timezone list in frontend (no external deps). |

### Post-design check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | Tests defined for: store methods (GetTimezone, UpdateTimezone), handler endpoints (GET /settings, PUT /settings/timezone), validation (invalid timezone rejection). |
| II. Security-First Design | PASS | RequireParent guards both endpoints. FamilyID from JWT context — no user-supplied family IDs. Timezone validated against Go stdlib. |
| III. Simplicity | PASS | Minimal scope: 1 migration, 1 new handler package (2 methods), 1 new page, 1 new component. No over-engineering. |

## Project Structure

### Documentation (this feature)

```text
specs/013-parent-settings/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 research output
├── data-model.md        # Phase 1 data model
├── quickstart.md        # Phase 1 quickstart guide
├── contracts/           # Phase 1 API contracts
│   └── settings-api.md
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── migrations/
│   └── 003_family_timezone.{up,down}.sql    # NEW: timezone column
├── internal/
│   ├── store/
│   │   └── family.go                        # MODIFIED: Timezone field + methods
│   │   └── family_test.go                   # MODIFIED: tests for new methods
│   ├── settings/
│   │   └── handlers.go                      # NEW: GET /settings, PUT /settings/timezone
│   │   └── handlers_test.go                 # NEW: handler tests
│   └── ...
└── main.go                                  # MODIFIED: register settings routes

frontend/
├── src/
│   ├── pages/
│   │   └── SettingsPage.tsx                 # NEW: settings page with category nav
│   ├── components/
│   │   └── TimezoneSelect.tsx               # NEW: searchable timezone selector
│   │   └── Layout.tsx                       # MODIFIED: add settings nav entry
│   ├── App.tsx                              # MODIFIED: add /settings route
│   ├── types.ts                             # MODIFIED: add SettingsResponse
│   └── api.ts                               # MODIFIED: add settings API functions
└── ...
```

**Structure Decision**: Follows existing web application structure. New `settings` handler package mirrors pattern of `family`, `balance`, `allowance`, `interest` packages. Frontend follows existing page + component patterns.

## Complexity Tracking

> No constitution violations. No complexity justifications needed.
