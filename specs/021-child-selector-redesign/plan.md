# Implementation Plan: Child Selector Redesign

**Branch**: `021-child-selector-redesign` | **Date**: 2026-02-23 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/021-child-selector-redesign/spec.md`

## Summary

Replace the vertical two-column child list/detail layout on ParentDashboard and Settings > Children with a horizontal chip-based child selector bar positioned above full-width content. The selector renders children as compact chips (avatar + name + balance) in a single scrollable row, supports toggle selection/deselection, and is implemented as a single reusable component for current and future parent-facing pages.

This is a **frontend-only** feature. No backend changes, no database migrations, no new API endpoints.

## Technical Context

**Language/Version**: TypeScript 5.3.3, React 18.2.0
**Primary Dependencies**: Vite, Tailwind CSS 4, lucide-react
**Storage**: N/A (uses existing `GET /children` API)
**Testing**: Manual acceptance testing (no frontend test infrastructure exists)
**Target Platform**: Web (desktop + mobile browsers)
**Project Type**: Web application (frontend only for this feature)
**Performance Goals**: Smooth horizontal scroll at 60fps, instant chip selection response
**Constraints**: Mobile-friendly (min 44px touch targets), supports 0–12 children
**Scale/Scope**: 2 page components refactored, 1 new component created, ~1 component removed from active use

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Research Gate

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Test-First Development** | JUSTIFIED DEVIATION | No frontend testing infrastructure exists in the project. This is a pure layout/UI refactor with no financial logic changes. All behavior is verifiable via manual acceptance scenarios. Setting up Vitest + RTL is out of scope and would violate Simplicity. |
| **II. Security-First Design** | PASS | No new data exposure. No new API endpoints. No authentication changes. The selector displays the same child data already shown on these pages. |
| **III. Simplicity** | PASS | Replaces complex two-column layouts with simpler stacked layout. Creates one focused component instead of modifying existing ChildList with variant logic. No premature abstractions. |

### Post-Design Gate

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Test-First Development** | JUSTIFIED DEVIATION | Same as above. The new ChildSelectorBar component is a pure presentational component with simple props. No financial calculations or data mutations. |
| **II. Security-First Design** | PASS | Component only renders data already fetched via authenticated API. No new security surface. |
| **III. Simplicity** | PASS | One new file, two refactored files. Native CSS horizontal scroll (no JS scroll libraries). Chip rendering follows existing AvatarPicker pattern. |

## Project Structure

### Documentation (this feature)

```text
specs/021-child-selector-redesign/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 research decisions
├── data-model.md        # Component interfaces (no new DB entities)
└── quickstart.md        # Developer quickstart guide
```

### Source Code (affected files)

```text
frontend/src/
├── components/
│   ├── ChildSelectorBar.tsx      # NEW — Reusable horizontal chip selector
│   ├── ChildrenSettings.tsx      # MODIFIED — Replace two-column with selector + full-width
│   ├── ManageChild.tsx           # MODIFIED — Remove close button (selector handles switching)
│   └── ChildList.tsx             # DEPRECATED — No longer imported (keep file, remove later)
└── pages/
    └── ParentDashboard.tsx       # MODIFIED — Replace two-column with selector + full-width
```

**Structure Decision**: Frontend-only changes within the existing `frontend/src/` directory. No new directories needed. The `contracts/` directory is omitted since no external interfaces are modified (existing API is unchanged).

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| No frontend tests (Principle I) | No test infrastructure exists; pure visual refactor with no financial logic | Setting up Vitest + RTL for a layout change is scope creep and violates Simplicity |
