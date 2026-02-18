# Implementation Plan: Savings Growth Projector

**Branch**: `016-savings-growth-projector` | **Date**: 2026-02-17 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/016-savings-growth-projector/spec.md`

## Summary

Add a child-facing "Growth" page that visualizes projected account balance over time using a line chart. The projection incorporates the child's current balance, active allowance schedule, and compound interest schedule. Children can explore what-if scenarios (weekly spending, one-time deposit/withdrawal) with live-updating graph and a plain-English breakdown of projected growth components. This is a **frontend-only feature** — no backend changes needed. One new dependency: Recharts for charting.

## Technical Context

**Language/Version**: TypeScript 5.3.3 + React 18.2.0 (frontend only — no backend changes)
**Primary Dependencies**: Recharts (new), react-router-dom, lucide-react, Vite (existing)
**Storage**: N/A — no persistence; all projections are computed client-side from existing API data
**Testing**: Vitest (frontend unit tests for projection math)
**Target Platform**: Web browser (desktop + mobile responsive)
**Project Type**: Web application (frontend only for this feature)
**Performance Goals**: Graph renders in <2s on page load; updates in <500ms on input change
**Constraints**: All monetary calculations in integer cents to avoid floating-point errors
**Scale/Scope**: 1 new page, 4 new components, 1 utility module, 2 file modifications

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Research Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | Projection engine is a pure function — ideal for TDD. Unit tests verify compound interest math against known calculations. |
| II. Security-First Design | PASS | Read-only feature; no new endpoints or mutations. Uses existing authenticated endpoints. Child can only see their own data. No user input is sent to the server. |
| III. Simplicity | PASS | One new dependency (Recharts) justified by clear value — building a chart library from scratch would be far more complex. Client-side calculation avoids unnecessary backend work. No persistence, no new abstractions. |

### Post-Design Re-Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | `projection.ts` will have comprehensive unit tests covering all compounding scenarios, edge cases (zero balance, paused schedules, depletion). |
| II. Security-First Design | PASS | No new attack surface. Scenario inputs are local state only — never sent to server. |
| III. Simplicity | PASS | Single new dependency. 4 focused components. Pure calculation function. No backend changes. No database changes. |

No constitution violations — Complexity Tracking section not needed.

## Project Structure

### Documentation (this feature)

```text
specs/016-savings-growth-projector/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: technology decisions
├── data-model.md        # Phase 1: client-side data interfaces
├── quickstart.md        # Phase 1: development setup guide
├── contracts/           # Phase 1: API contract documentation
│   └── README.md        # Documents existing endpoints used
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
frontend/
├── src/
│   ├── pages/
│   │   └── GrowthPage.tsx              # NEW — main page component
│   ├── components/
│   │   ├── GrowthChart.tsx             # NEW — Recharts line chart wrapper
│   │   ├── ScenarioControls.tsx        # NEW — what-if input form
│   │   └── GrowthExplanation.tsx       # NEW — plain-English summary
│   ├── utils/
│   │   └── projection.ts              # NEW — pure projection calculation engine
│   ├── App.tsx                         # MODIFY — add /child/growth route
│   ├── components/Layout.tsx           # MODIFY — add "Growth" nav item for children
│   └── types.ts                        # MODIFY — add projection-related interfaces
└── package.json                        # MODIFY — add recharts dependency
```

**Structure Decision**: Frontend-only changes following the existing web application structure. No backend modifications. New files are placed in existing directories (`pages/`, `components/`, `utils/`) following established patterns. The projection utility is isolated in `utils/` as a pure function for testability.
