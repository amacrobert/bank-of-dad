# Quickstart: Growth Projector Scenarios

## Overview

This feature enhances the existing Growth Projector (016) to support multiple "what if" scenarios displayed as distinct colored lines on a single graph. It also simplifies the input controls by combining mutually exclusive fields and adds dynamic plain-English scenario titles.

## Prerequisites

- Node.js and npm installed
- Frontend dev server: `cd frontend && npm run dev`
- At least one child account with balance data (use existing seed data or create via the app)

## Key Files to Modify

### Existing files (modify)

| File | Change |
|------|--------|
| `frontend/src/types.ts` | Add `ScenarioConfig`, update `ScenarioInputs` with `weeklySavingsCents` |
| `frontend/src/utils/projection.ts` | Handle `weeklySavingsCents` in weekly loop |
| `frontend/src/utils/projection.test.ts` | Add saving-direction tests |
| `frontend/src/pages/GrowthPage.tsx` | Multi-scenario state, URL sync, remove GrowthExplanation |
| `frontend/src/components/GrowthChart.tsx` | Multiple `<Line>` elements, updated tooltip |
| `frontend/src/components/ScenarioControls.tsx` | Rewrite for per-scenario combined fields |

### New files (create)

| File | Purpose |
|------|---------|
| `frontend/src/utils/scenarioTitle.ts` | Pure function for title generation |
| `frontend/src/utils/scenarioTitle.test.ts` | Exhaustive title rule tests |
| `frontend/src/utils/scenarioUrl.ts` | URL serialization/deserialization |
| `frontend/src/utils/scenarioUrl.test.ts` | URL round-trip tests |

### Files to delete

| File | Reason |
|------|--------|
| `frontend/src/components/GrowthExplanation.tsx` | Replaced by per-scenario titles (FR-016) |

## Development Flow

1. **Start with types + projection engine** — Update `ScenarioInputs`, add `weeklySavingsCents` support to `calculateProjection`, write tests
2. **Title generation** — Implement and test `generateScenarioTitle` pure function
3. **URL serialization** — Implement and test `serializeScenarios` / `deserializeScenarios`
4. **Chart updates** — Multi-line support in `GrowthChart`, updated tooltip
5. **Scenario controls** — Rewrite `ScenarioControls` for combined fields, add/remove UI
6. **Page integration** — Wire everything together in `GrowthPage`, add URL sync, remove `GrowthExplanation`

## Running Tests

```bash
cd frontend && npx vitest run
```

## No Backend Changes Required

This is a frontend-only feature. All projection calculations happen client-side. No new API endpoints, migrations, or backend code changes are needed.
