# Implementation Plan: Child Avatars

**Branch**: `010-child-avatars` | **Date**: 2026-02-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/010-child-avatars/spec.md`

## Summary

Add optional emoji avatars to child accounts. A nullable `avatar` TEXT column is added to the `children` table. Parents select from a fixed grid of 16 emojis when creating or updating a child. The avatar replaces the first-letter initial in the child selection list. Existing endpoints are extended with an optional `avatar` field — no new endpoints needed.

## Technical Context

**Language/Version**: Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: `modernc.org/sqlite`, `testify`, react-router-dom, Vite, lucide-react
**Storage**: SQLite with WAL mode (separate read/write connections)
**Testing**: `go test ./...` (backend), TypeScript type checking (frontend)
**Target Platform**: Web application (browser + Go HTTP server)
**Project Type**: Web application (backend + frontend)
**Performance Goals**: N/A — cosmetic feature, no performance implications
**Constraints**: Emoji must render correctly cross-platform (standard Unicode)
**Scale/Scope**: 1 new column, 1 new component, ~5 modified files per side

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development

- **Store tests**: Add tests for avatar in Create, Get, List, Update operations
- **No financial data affected**: Avatars are cosmetic, no balance/transaction impact
- **Status**: PASS — tests planned for all store mutations

### II. Security-First Design

- **No sensitive data**: Avatars are emoji strings, not PII or credentials
- **Existing auth enforced**: All endpoints use `requireParent` middleware
- **Input handling**: Avatar stored as-is; no execution risk from emoji strings
- **Status**: PASS — no security implications

### III. Simplicity

- **YAGNI**: No avatar validation, no avatar history, no avatar upload — just store a string
- **Minimal dependencies**: No new dependencies added
- **Single responsibility**: One shared `AvatarPicker` component reused in two forms
- **Extends existing patterns**: Same column migration, struct field, and handler extension patterns used throughout codebase
- **Status**: PASS — minimal additions

### Post-Design Re-check

- **PASS**: All three principles satisfied. No violations to track.

## Project Structure

### Documentation (this feature)

```text
specs/010-child-avatars/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── api.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── store/
│   │   ├── sqlite.go          # Migration: add avatar column
│   │   ├── child.go           # Struct + query changes
│   │   └── child_test.go      # Avatar tests
│   └── family/
│       └── handlers.go        # Create, List, Update handler changes

frontend/
├── src/
│   ├── types.ts               # Add avatar to Child interface
│   ├── components/
│   │   ├── AvatarPicker.tsx   # NEW: shared emoji grid component
│   │   ├── AddChildForm.tsx   # Add avatar picker
│   │   ├── ManageChild.tsx    # Rename form, add avatar picker
│   │   └── ChildList.tsx      # Conditional avatar display
│   └── api.ts                 # No changes needed (uses existing post/put)
```

**Structure Decision**: Follows established web application pattern with `backend/` and `frontend/` directories. One new file (`AvatarPicker.tsx`), rest are modifications to existing files.

## Complexity Tracking

> No constitution violations. Table intentionally left empty.
