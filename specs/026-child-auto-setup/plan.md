# Implementation Plan: Child Auto-Setup

**Branch**: `026-child-auto-setup` | **Date**: 2026-03-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/026-child-auto-setup/spec.md`

## Summary

Add three optional fields (Initial Deposit, Weekly Allowance, Annual Interest) to the existing Add Child form. After child creation, the frontend calls existing backend API endpoints in sequence to set up deposit, allowance schedule, and interest schedule. No backend changes required — this is a frontend-only feature that orchestrates existing APIs.

## Technical Context

**Language/Version**: Go 1.24 (backend, unchanged), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: Vite, Tailwind CSS 4, lucide-react (all existing)
**Storage**: PostgreSQL 17 (existing, no schema changes)
**Testing**: `go test -p 1 ./...` (backend, unchanged), `npx tsc --noEmit && npm run build && npm run lint` (frontend)
**Target Platform**: Web (desktop + mobile responsive)
**Project Type**: Web application (full-stack)
**Performance Goals**: N/A — form submission with sequential API calls, sub-second expected
**Constraints**: Must work on mobile viewports (fields stack vertically on narrow screens)
**Scale/Scope**: 1 component modified (AddChildForm.tsx), 0 new files, 0 backend changes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | Frontend-only change; existing backend API endpoints already have full test coverage. Frontend validation uses same rules as existing forms. Manual verification via `tsc --noEmit`, `npm run build`, `npm run lint`. |
| II. Security-First Design | PASS | All API calls go through existing authenticated endpoints (`requireParent` middleware). No new endpoints, no new auth flows. Input validation reuses existing patterns. |
| III. Simplicity | PASS | No new abstractions, no new files, no backend changes. Reuses existing API functions (`deposit`, `setChildAllowance`, `setInterest`). Sequential API calls keep error handling straightforward. |

**Post-Phase 1 Re-check**: All gates still pass. No new dependencies, no schema changes, no new abstractions introduced.

## Project Structure

### Documentation (this feature)

```text
specs/026-child-auto-setup/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (empty — no new contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
frontend/
├── src/
│   ├── components/
│   │   └── AddChildForm.tsx   # MODIFIED — add 3 optional fields + post-create API calls
│   └── api.ts                 # UNCHANGED — already exports deposit, setChildAllowance, setInterest
```

**Structure Decision**: Frontend-only modification to the existing `AddChildForm.tsx` component. No new files needed. The existing API helper functions in `api.ts` already provide `deposit()`, `setChildAllowance()`, and `setInterest()` — these are called sequentially after child creation succeeds.
