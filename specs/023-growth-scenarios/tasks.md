# Tasks: Growth Projector Scenarios

**Input**: Design documents from `/specs/023-growth-scenarios/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Included per constitution (Test-First Development). Tests written before implementation.

**Organization**: Tasks grouped by user story. Frontend-only feature — no backend changes.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Types & Projection Engine)

**Purpose**: Establish shared types and projection engine updates that all user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T001 Add `ScenarioConfig`, `WeeklyDirection`, `OneTimeDirection` types and `SCENARIO_COLORS` palette constant, and update `ScenarioInputs` to include `weeklySavingsCents: number` field in `frontend/src/types.ts`
- [x] T002 Write tests for `weeklySavingsCents` support in `frontend/src/utils/projection.test.ts`: weekly saving adds to balance each week, saving + allowance stack, saving + interest compound correctly, component breakdown integrity includes savings
- [x] T003 Implement `weeklySavingsCents` in projection engine weekly loop (add to balance each week, track `totalSavingsCents` in result) in `frontend/src/utils/projection.ts` and update `ProjectionResult` in `frontend/src/types.ts`
- [x] T004 Write tests for `mapScenarioConfigToInputs` helper that converts `ScenarioConfig` → `ScenarioInputs` (spending direction maps to `weeklySpendingCents`, saving direction maps to `weeklySavingsCents`, deposit maps to `oneTimeDepositCents`, withdrawal maps to `oneTimeWithdrawalCents`) and `buildDefaultScenarios` that returns two default `ScenarioConfig[]` based on allowance presence in `frontend/src/utils/scenarioHelpers.test.ts`
- [x] T005 Implement `mapScenarioConfigToInputs` and `buildDefaultScenarios` in `frontend/src/utils/scenarioHelpers.ts`: default with allowance = [spending=0, spending=100% weekly allowance]; default without allowance = [spending=0, spending=$5/week]

**Checkpoint**: Types defined, projection engine updated, helper functions ready — user story implementation can begin

---

## Phase 2: User Story 1 - Compare Two Growth Scenarios Side-by-Side (Priority: P1) MVP

**Goal**: Display two default scenario lines on a single graph with distinct colors and a multi-scenario tooltip

**Independent Test**: Open the growth projector for a child with an allowance → two colored lines appear on the graph, each with a distinct color; hover shows tooltip with scenario info

### Implementation for User Story 1

- [x] T006 [US1] Update `GrowthChart.tsx` to accept an array of scenario data (each with `dataPoints`, `color`, `label`), render one `<Line>` per scenario with its assigned color, and merge all data points into a unified chart data array with keyed balance fields (`balanceCents_0`, `balanceCents_1`, etc.) in `frontend/src/components/GrowthChart.tsx`
- [x] T007 [US1] Update `CustomTooltip` in `GrowthChart.tsx` to display all scenario labels (use placeholder labels "Scenario 1", "Scenario 2", etc. — upgraded to generated titles in T018) and their projected balances at the hovered date in `frontend/src/components/GrowthChart.tsx`
- [x] T008 [US1] Rewrite `GrowthPage.tsx` `ProjectorContent` to manage `ScenarioConfig[]` state (initialized via `buildDefaultScenarios`), compute `calculateProjection` per scenario using `mapScenarioConfigToInputs`, merge results into chart data, and pass to updated `GrowthChart` in `frontend/src/pages/GrowthPage.tsx`
- [x] T009 [US1] Ensure horizon selector in `ProjectorContent` applies shared `horizonMonths` to all scenarios uniformly (FR-018) in `frontend/src/pages/GrowthPage.tsx`
- [x] T010 [US1] Create initial per-scenario `ScenarioControls.tsx` that renders a row per `ScenarioConfig` with: color indicator dot, two amount inputs ("Weekly" and "One time"), and callbacks to update scenario state; replace single-scenario controls in `frontend/src/components/ScenarioControls.tsx`
- [x] T011 [US1] Remove `GrowthExplanation.tsx` component file and remove all imports/usage from `GrowthPage.tsx` (FR-016) — delete `frontend/src/components/GrowthExplanation.tsx`, update `frontend/src/pages/GrowthPage.tsx`

**Checkpoint**: Two default scenario lines visible on graph, each with distinct color. Tooltip shows both. Basic editing works. GrowthExplanation removed.

---

## Phase 3: User Story 2 - Combined Input Fields with Toggles (Priority: P2)

**Goal**: Each scenario's inputs use combined fields with direction toggles: "Weekly" (spending/saving) and "One time" (deposit/withdrawal)

**Independent Test**: Open a scenario, toggle "Weekly" from spending to saving, enter a value → graph projects additional savings. Toggle "One time" from deposit to withdrawal → graph reflects the change.

### Implementation for User Story 2

- [x] T012 [US2] Add spending/saving toggle button to the "Weekly" field in each scenario row — toggling updates `weeklyDirection` on the `ScenarioConfig`, which triggers projection recalculation via the parent callback in `frontend/src/components/ScenarioControls.tsx`
- [x] T013 [US2] Add deposit/withdrawal toggle button to the "One time" field in each scenario row — toggling updates `oneTimeDirection` on the `ScenarioConfig` in `frontend/src/components/ScenarioControls.tsx`
- [x] T014 [US2] Add withdrawal validation: when `oneTimeDirection === 'withdrawal'`, show error if `oneTimeAmountCents > currentBalanceCents` (consistent with existing behavior) in `frontend/src/components/ScenarioControls.tsx`

**Checkpoint**: Combined fields with direction toggles work. Switching spending↔saving and deposit↔withdrawal updates the graph correctly.

---

## Phase 4: User Story 3 - Dynamic Scenario Titles (Priority: P3)

**Goal**: Each scenario row displays a plain-English title that dynamically updates based on inputs and allowance context

**Independent Test**: Change scenario inputs → title updates to match the expected phrasing rules from the feature description

### Tests for User Story 3

> **Write these tests FIRST, ensure they FAIL before implementation**

- [x] T015 [P] [US3] Write comprehensive tests for `generateScenarioTitle` covering all title rules from the feature description in `frontend/src/utils/scenarioTitle.test.ts`: (1) has allowance + no one-time: spending=0 → "save **all**", spending=allowance → "save **none**", 0<spending<allowance → "save **$Y** per week", saving>0 → "save **all** plus additional **$Y**"; (2) has allowance + deposit: all four spending variants with "and deposit $Z now"; (3) has allowance + withdrawal: all four spending variants with "but withdraw $Z now"; (4) no allowance + no one-time: spending=0 → "don't do anything", spending>0 → "spend $X per week", saving>0 → "save $X per week"; (5) no allowance + deposit; (6) no allowance + withdrawal

### Implementation for User Story 3

- [x] T016 [US3] Implement `generateScenarioTitle(context: ScenarioTitleContext): string` as a pure function in `frontend/src/utils/scenarioTitle.ts` — takes `hasAllowance`, `weeklyAllowanceCents`, `weeklyAmountCents`, `weeklyDirection`, `oneTimeAmountCents`, `oneTimeDirection` and returns the formatted title string with bold markers
- [x] T017 [US3] Display the generated scenario title as the label for each scenario row in `ScenarioControls.tsx` — call `generateScenarioTitle` with the current scenario config and allowance context, render the title with bold formatting in `frontend/src/components/ScenarioControls.tsx`
- [x] T018 [US3] Update `CustomTooltip` in `GrowthChart.tsx` to show each scenario's generated title above the date and balance (FR-009) in `frontend/src/components/GrowthChart.tsx`

**Checkpoint**: Scenario titles dynamically update in both the scenario card rows and the chart tooltip. All title rules from the feature description are covered.

---

## Phase 5: User Story 4 - Add and Remove Scenarios (Priority: P4)

**Goal**: Users can add new scenarios (up to 5 max) and remove existing ones (min 1 must remain)

**Independent Test**: Click "Add scenario" → third row + line appear with new color. Delete a scenario → row and line removed. Only one left → delete button disabled.

### Implementation for User Story 4

- [x] T019 [US4] Add "Add scenario" button below scenario rows in `ScenarioControls.tsx` — clicking adds a new `ScenarioConfig` with defaults (weeklyAmountCents=0, weeklyDirection='spending', oneTimeAmountCents=0, oneTimeDirection='deposit') and next available color from `SCENARIO_COLORS`; hide button when 5 scenarios exist (FR-012) in `frontend/src/components/ScenarioControls.tsx`
- [x] T020 [US4] Add delete button (X or trash icon) per scenario row in `ScenarioControls.tsx` — clicking removes the scenario from state; disable/hide when only 1 scenario remains (FR-011) in `frontend/src/components/ScenarioControls.tsx`
- [x] T021 [US4] Handle color reassignment in `GrowthPage.tsx` or `scenarioHelpers.ts`: when a scenario is removed and a new one added, assign the next unused color from the palette to avoid duplicate colors in `frontend/src/utils/scenarioHelpers.ts`

**Checkpoint**: Scenarios can be added (up to 5) and removed (min 1). Each has a unique color. Graph updates dynamically.

---

## Phase 6: User Story 5 - Bookmarkable Scenario URLs (Priority: P5)

**Goal**: Scenario state persists in URL query parameters for bookmarking and sharing

**Independent Test**: Configure scenarios → copy URL → navigate away → paste URL → same scenarios and graph restored

### Tests for User Story 5

> **Write these tests FIRST, ensure they FAIL before implementation**

- [x] T022 [P] [US5] Write tests for URL serialization/deserialization in `frontend/src/utils/scenarioUrl.test.ts`: (1) round-trip: serialize then deserialize returns equivalent scenarios; (2) serialize produces compact base64 URL param; (3) deserialize with missing param returns null; (4) deserialize with malformed base64 returns null; (5) deserialize with invalid JSON returns null; (6) horizon param round-trips correctly; (7) handles 1-5 scenarios

### Implementation for User Story 5

- [x] T023 [US5] Implement `serializeScenarios(scenarios: ScenarioConfig[], horizonMonths: number): string` and `deserializeScenarios(searchParams: URLSearchParams): { scenarios: ScenarioConfig[]; horizonMonths: number } | null` in `frontend/src/utils/scenarioUrl.ts` — use compact JSON keys (`w`, `wd`, `o`, `od`) and URL-safe base64 encoding (replace `+`→`-`, `/`→`_`, strip trailing `=`) for the `scenarios` param, plain number for `h` param
- [x] T024 [US5] Wire URL reading into `GrowthPage.tsx`: on mount (after child data loads), call `deserializeScenarios` with current `URLSearchParams`; if valid, use decoded scenarios instead of defaults from `buildDefaultScenarios` in `frontend/src/pages/GrowthPage.tsx`
- [x] T025 [US5] Wire URL writing into `GrowthPage.tsx`: when scenarios or horizon change, call `serializeScenarios` and update URL via `window.history.replaceState` (avoid navigation/re-render) in `frontend/src/pages/GrowthPage.tsx`

**Checkpoint**: Scenario state survives page reload via URL. Invalid/missing params fall back to defaults.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [x] T026 [P] Run TypeScript type check (`cd frontend && npx tsc --noEmit`) and fix any type errors
- [x] T027 [P] Run linter (`cd frontend && npm run lint`) and fix any warnings
- [x] T028 Run all Vitest tests (`cd frontend && npx vitest run`) and ensure all pass (existing + new)
- [ ] T029 Manual walkthrough: verify acceptance scenarios for all 5 user stories against a child with allowance and a child without allowance

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — start immediately. BLOCKS all user stories.
- **US1 (Phase 2)**: Depends on Phase 1 completion. Core MVP.
- **US2 (Phase 3)**: Depends on Phase 2 (needs scenario controls from US1)
- **US3 (Phase 4)**: Depends on Phase 2 (needs scenario config structure). Title tests (T015) can start after Phase 1.
- **US4 (Phase 5)**: Depends on Phase 2 (needs multi-scenario state from US1)
- **US5 (Phase 6)**: Depends on Phase 1 (needs ScenarioConfig type). URL utility tests (T022) can start after Phase 1. Wiring (T024-T025) depends on Phase 2.
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

```
Phase 1: Foundational
    │
    ├── Phase 2: US1 (Multi-line graph)  ← MVP
    │       │
    │       ├── Phase 3: US2 (Combined fields with toggles)
    │       │
    │       ├── Phase 4: US3 (Dynamic titles)
    │       │
    │       └── Phase 5: US4 (Add/remove scenarios)
    │
    └── Phase 6: US5 (URL bookmarks — utility is independent, wiring needs US1)
            │
            └── Phase 7: Polish
```

### Parallel Opportunities

- **Within Phase 1**: T002 and T004 can be written in parallel (different test files)
- **After Phase 1**: T015 (US3 title tests) and T022 (US5 URL tests) can start in parallel with US1 implementation
- **Phase 7**: T026 and T027 can run in parallel (different tools)

---

## Parallel Example: After Phase 1

```
# These can all start in parallel after Phase 1 completes:
Stream A: T006-T011 (US1 implementation — the MVP)
Stream B: T015 (US3 title tests — pure function, no UI dependency)
Stream C: T022-T023 (US5 URL serialization tests + implementation — pure function)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Foundational (T001-T005)
2. Complete Phase 2: US1 (T006-T011)
3. **STOP and VALIDATE**: Two scenario lines on graph, basic editing, GrowthExplanation removed
4. Deploy/demo if ready — this is the core value

### Incremental Delivery

1. Phase 1 → Foundation ready
2. + US1 → Two scenarios side-by-side (MVP!)
3. + US2 → Polished toggle inputs
4. + US3 → Dynamic titles for clarity
5. + US4 → Add/remove scenarios for power users
6. + US5 → Bookmarkable URLs
7. Polish → Ship it

---

## Notes

- Frontend-only feature — no backend tasks, migrations, or API changes
- All pure function logic (titles, URL, helpers) is fully unit-testable via Vitest
- `GrowthExplanation.tsx` is deleted in US1 (T011) and should not be referenced after that point
- The `ScenarioInputs` type retains backward compatibility — the new `weeklySavingsCents` field defaults to `0`
- Existing projection tests must continue to pass throughout (T003 must not break them)
