# Implementation Plan: Growth Projector Scenarios

**Branch**: `023-growth-scenarios` | **Date**: 2026-02-23 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/023-growth-scenarios/spec.md`

## Summary

Enhance the Growth Projector to display multiple "what if" scenarios as distinct colored lines on a single graph. Simplify inputs by combining mutually exclusive fields (deposit/withdrawal, spending/saving) with direction toggles. Add dynamic plain-English scenario titles, add/remove scenario controls, and persist scenario state in URL query parameters for bookmarking.

This is a **frontend-only** feature. No backend changes, migrations, or new API endpoints are required.

## Technical Context

**Language/Version**: TypeScript 5.3.3, React 18.2.0
**Primary Dependencies**: Recharts (charting), react-router-dom (URL params), Vite (build)
**Storage**: N/A (client-side only, URL query parameters for persistence)
**Testing**: Vitest (unit tests for pure functions)
**Target Platform**: Web browser (desktop + mobile responsive)
**Project Type**: Web application (frontend only for this feature)
**Performance Goals**: Graph renders within 500ms of input change, page loads in under 2 seconds
**Constraints**: Max 5 concurrent scenarios, URL must remain shareable/bookmarkable
**Scale/Scope**: Single page enhancement, ~8 files modified/created, ~1 file deleted

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Research Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | Title generation, URL serialization, and projection engine changes will have comprehensive unit tests via Vitest. TDD cycle: write failing tests first, then implement. |
| II. Security-First Design | PASS | No new endpoints, no user data transmitted. URL params contain only numeric scenario values (no PII). Input validation for withdrawal amounts against balance. |
| III. Simplicity | PASS | Reuses existing `calculateProjection` engine with minimal extension. Combined fields reduce UI complexity. No new dependencies needed. |

### Post-Design Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | Test files identified: `projection.test.ts` (extend), `scenarioTitle.test.ts` (new), `scenarioUrl.test.ts` (new). All business logic is in pure functions. |
| II. Security-First Design | PASS | No changes to auth or API layer. Withdrawal validation preserved. URL params validated on decode with fallback to defaults. |
| III. Simplicity | PASS | No new npm dependencies. Existing Recharts multi-line support used. Pure utility functions for testable logic. |

## Project Structure

### Documentation (this feature)

```text
specs/023-growth-scenarios/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 research decisions
├── data-model.md        # Phase 1 data model
├── quickstart.md        # Phase 1 quickstart guide
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (repository root)

```text
frontend/src/
├── types.ts                           # Update: add ScenarioConfig, update ScenarioInputs
├── utils/
│   ├── projection.ts                  # Update: add weeklySavingsCents support
│   ├── projection.test.ts             # Update: add saving-direction tests
│   ├── scenarioTitle.ts               # NEW: title generation pure function
│   ├── scenarioTitle.test.ts          # NEW: exhaustive title rule tests
│   ├── scenarioUrl.ts                 # NEW: URL serialization/deserialization
│   └── scenarioUrl.test.ts            # NEW: URL round-trip tests
├── components/
│   ├── GrowthChart.tsx                # Update: multi-line support, updated tooltip
│   ├── ScenarioControls.tsx           # Rewrite: per-scenario combined fields, add/remove
│   └── GrowthExplanation.tsx          # DELETE: replaced by per-scenario titles
└── pages/
    └── GrowthPage.tsx                 # Update: multi-scenario state, URL sync, wire together
```

**Structure Decision**: Frontend-only changes within existing directory structure. Three new utility files for testable business logic, modifications to three existing components, deletion of one component, and type updates.

## Detailed Design

### 1. Type Changes (`types.ts`)

**Add `ScenarioConfig`** — the UI-facing per-scenario model:
```typescript
interface ScenarioConfig {
  id: string;                                // "s0", "s1", etc.
  weeklyAmountCents: number;                 // >= 0
  weeklyDirection: 'spending' | 'saving';
  oneTimeAmountCents: number;                // >= 0
  oneTimeDirection: 'deposit' | 'withdrawal';
  color: string;                             // hex color from palette
}
```

**Update `ScenarioInputs`** — add `weeklySavingsCents`:
```typescript
interface ScenarioInputs {
  weeklySpendingCents: number;
  weeklySavingsCents: number;     // NEW
  oneTimeDepositCents: number;
  oneTimeWithdrawalCents: number;
  horizonMonths: number;
}
```

### 2. Projection Engine (`projection.ts`)

Add weekly savings to the simulation loop. In each week iteration, after applying allowance and interest:
- Subtract `weeklySpendingCents` (existing)
- Add `weeklySavingsCents` (new)

Track total savings in `ProjectionResult` (new `totalSavingsCents` field) and include in component breakdown integrity.

### 3. Title Generation (`scenarioTitle.ts`)

Pure function: `generateScenarioTitle(context: ScenarioTitleContext): string`

Implements all title rules from the feature description. Branches on:
- `hasAllowance` (boolean)
- `weeklyDirection` + `weeklyAmountCents` (spending 0, spending == allowance, spending < allowance, saving > 0)
- `oneTimeDirection` + `oneTimeAmountCents` (none, deposit, withdrawal)

Returns a string with markdown-style bold markers (e.g., `**all**`, `**$15**`). The rendering component strips or converts bold markers as needed.

### 4. URL Serialization (`scenarioUrl.ts`)

Two functions:
- `serializeScenarios(scenarios: ScenarioConfig[], horizonMonths: number): string` — returns query string
- `deserializeScenarios(searchParams: URLSearchParams): { scenarios: ScenarioConfig[]; horizonMonths: number } | null` — returns parsed scenarios or null (invalid/missing)

Format: `?scenarios=<base64(JSON)>&h=<months>`

Compact JSON keys: `w` (weeklyAmountCents), `wd` (weeklyDirection: `"s"`/`"v"`), `o` (oneTimeAmountCents), `od` (oneTimeDirection: `"d"`/`"w"`).

### 5. Chart Updates (`GrowthChart.tsx`)

- Accept array of `{ dataPoints, color, title }` instead of single `dataPoints`
- Merge all scenario data points into a single array with keyed balance fields (`balanceCents_0`, `balanceCents_1`, etc.)
- Render one `<Line>` per scenario with corresponding color
- Update `CustomTooltip` to show all scenario titles + balances

### 6. Scenario Controls (`ScenarioControls.tsx`)

Complete rewrite:
- Renders a list of scenario rows, each with:
  - Color indicator dot
  - Dynamic title (from `generateScenarioTitle`)
  - "Weekly" field: amount input + spending/saving toggle
  - "One time" field: amount input + deposit/withdrawal toggle
  - Delete button (disabled when only 1 scenario)
- "Add scenario" button at the bottom (hidden when at 5 scenarios)
- "WHAT IF..." header retained

### 7. Page Integration (`GrowthPage.tsx`)

- Replace single `scenario` state with `scenarios: ScenarioConfig[]` array
- Shared `horizonMonths` state (separate from per-scenario data)
- Initialize defaults from allowance data (per spec default rules)
- Read URL params on mount via `deserializeScenarios`; if present, use those; otherwise use defaults
- On scenario change, call `serializeScenarios` and update URL via `useSearchParams` or `window.history.replaceState`
- Run `calculateProjection` for each scenario, merge results for chart
- Remove `GrowthExplanation` import and usage

## Complexity Tracking

> No constitution violations. All design choices align with the three core principles.
