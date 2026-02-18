# Tasks: Savings Growth Projector

**Input**: Design documents from `/specs/016-savings-growth-projector/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Unit tests included for the projection engine per Constitution Principle I (TDD) — financial calculations MUST have corresponding tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Web app (frontend only)**: `frontend/src/`
- No backend changes required for this feature

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Install dependencies, set up test infrastructure, define shared TypeScript interfaces

- [ ] T001 Install recharts dependency and configure vitest for unit testing in frontend/package.json
- [ ] T002 [P] Add projection-related TypeScript interfaces (ScenarioInputs, ProjectionDataPoint, ProjectionResult, ProjectionConfig) to frontend/src/types.ts

---

## Phase 2: Foundational (Projection Engine)

**Purpose**: Build and test the pure projection calculation engine that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete. The projection engine is the foundation for the graph, scenario controls, explanation text, and horizon selector.

- [ ] T003 Write unit tests for calculateProjection covering: (a) allowance-only linear growth, (b) interest-only compound growth, (c) combined allowance + interest, (d) weekly/biweekly/monthly frequencies, (e) zero balance with no schedules (flat line), (f) balance floor at $0, (g) depletion week detection, (h) one-time deposit/withdrawal adjustments, (i) paused schedules excluded — in frontend/src/utils/projection.test.ts
- [ ] T004 Implement calculateProjection() pure function using weekly-step iteration with schedule-based discrete compounding (per-period rate = interest_rate_bps / 10000 / periods_per_year) that passes all tests — in frontend/src/utils/projection.ts

**Checkpoint**: Projection engine complete with passing tests. All financial math verified against known compound interest calculations.

---

## Phase 3: User Story 1 - View Projected Account Growth (Priority: P1) 🎯 MVP

**Goal**: A child navigates to the Growth page and sees a line chart showing projected balance over a default 1-year horizon, incorporating their current balance, active allowance, and active interest schedule.

**Independent Test**: Log in as a child with balance + allowance + interest configured. Click "Growth" in sidebar. Verify graph shows an upward curve over 1 year. Verify child with no schedules sees flat line. Verify child with only allowance sees linear growth.

### Implementation for User Story 1

- [ ] T005 [US1] Create GrowthChart component wrapping Recharts ResponsiveContainer + LineChart + Line + XAxis + YAxis + Tooltip, accepting ProjectionDataPoint[] as data and rendering balance over time with dollar-formatted Y axis and date-formatted X axis — in frontend/src/components/GrowthChart.tsx
- [ ] T006 [US1] Create GrowthPage component with auth check (redirect non-child users), parallel data fetching (getBalance, getChildAllowance, getInterestSchedule), projection calculation via calculateProjection with default 1-year horizon, and GrowthChart rendering — in frontend/src/pages/GrowthPage.tsx
- [ ] T007 [P] [US1] Add `/child/growth` route pointing to GrowthPage — in frontend/src/App.tsx
- [ ] T008 [P] [US1] Add "Growth" navigation item with TrendingUp icon for child users, positioned after "Home" in both desktop sidebar and mobile bottom tab bar — in frontend/src/components/Layout.tsx

**Checkpoint**: Child can navigate to Growth page via sidebar and see projected balance graph. MVP is functional and independently testable.

---

## Phase 4: User Story 2 - Explore What-If Scenarios (Priority: P2)

**Goal**: Child can adjust weekly spending, one-time deposit, and one-time withdrawal inputs. Graph updates live as inputs change.

**Independent Test**: On Growth page, enter $5 in weekly spending — graph curve should flatten/decrease. Enter $50 in one-time deposit — graph should shift upward. Enter more than current balance in withdrawal — validation message appears.

### Implementation for User Story 2

- [ ] T009 [US2] Create ScenarioControls component with three dollar-amount inputs (weekly spending, one-time deposit, one-time withdrawal) using the existing Input component, with withdrawal validation capped at current balance displaying an error message — in frontend/src/components/ScenarioControls.tsx
- [ ] T010 [US2] Integrate ScenarioControls into GrowthPage: lift scenario state, wire onChange handlers to recalculate projection and update GrowthChart on every input change — in frontend/src/pages/GrowthPage.tsx

**Checkpoint**: Child can adjust all three scenario inputs and see graph update in real-time. Withdrawal validation prevents exceeding balance.

---

## Phase 5: User Story 3 - Read Plain-English Growth Explanation (Priority: P3)

**Goal**: Below the graph, show a plain-English summary breaking down projected total into principal, interest, allowance, and spending components. Updates live with scenario changes.

**Independent Test**: Verify explanation text shows correct dollar amounts that sum to the projected total. Change scenario inputs — text updates. Verify child with no schedules sees encouraging "ask your parent" message.

### Implementation for User Story 3

- [ ] T011 [US3] Create GrowthExplanation component that accepts ProjectionResult and renders plain-English text: "If you keep saving your $X weekly allowance, in [horizon] your account will have $Y total. That's $A from what you have now, $B from interest, and $C from allowance deposits [minus $D in spending]." Include depletion warning when depletionWeek is set, and encouraging message when no growth sources configured — in frontend/src/components/GrowthExplanation.tsx
- [ ] T012 [US3] Integrate GrowthExplanation into GrowthPage below the chart and scenario controls, passing the current ProjectionResult — in frontend/src/pages/GrowthPage.tsx

**Checkpoint**: Child sees plain-English explanation that updates with graph. Component breakdown sums to projected total.

---

## Phase 6: User Story 4 - Adjust Projection Time Horizon (Priority: P4)

**Goal**: Child can select from 3 months, 6 months, 1 year, 2 years, or 5 years. Graph, explanation, and all projected values update on selection.

**Independent Test**: Switch to 5 years — graph X axis extends, projected values increase. Switch to 3 months — graph contracts. Verify scenario inputs persist across horizon changes.

### Implementation for User Story 4

- [ ] T013 [US4] Add time horizon selector (button group or segmented control with options: 3mo, 6mo, 1yr, 2yr, 5yr) to GrowthPage above or near the chart, wire selection to horizonMonths in scenario state triggering projection recalculation — in frontend/src/pages/GrowthPage.tsx
- [ ] T014 [US4] Update GrowthExplanation to dynamically reference the selected time horizon in the text (e.g., "in 3 months", "in 5 years") instead of hardcoded "in 1 year" — in frontend/src/components/GrowthExplanation.tsx

**Checkpoint**: Child can switch between all 5 time horizons. Graph, explanation, and values update. Scenario inputs persist.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Edge case handling, loading states, and visual polish

- [ ] T015 Add no-growth-sources empty state to GrowthPage: when child has no active allowance and no active interest, show flat-line graph with an encouraging message ("Ask your parent about setting up an allowance!") — in frontend/src/pages/GrowthPage.tsx
- [ ] T016 [P] Add paused-schedule notices to GrowthExplanation: when allowance or interest is paused, note it in the explanation text (e.g., "Your allowance is currently paused.") — in frontend/src/components/GrowthExplanation.tsx
- [ ] T017 Add loading spinner while data fetches, error state for failed API calls, and fade-in-up animation (animate-fade-in-up class) consistent with dashboard pages — in frontend/src/pages/GrowthPage.tsx

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on T002 (types) from Setup — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 completion (projection engine)
- **US2 (Phase 4)**: Depends on US1 (needs GrowthPage to integrate into)
- **US3 (Phase 5)**: Depends on US1 (needs GrowthPage); benefits from US2 (scenario state) but can be done with defaults only
- **US4 (Phase 6)**: Depends on US1 (needs GrowthPage with chart)
- **Polish (Phase 7)**: Depends on all user stories being complete

### Within Each Phase

- T001 and T002 can run in parallel (different files)
- T003 MUST complete before T004 (TDD: tests first)
- T005 can run in parallel with T007, T008 (different files)
- T006 depends on T005 (composes GrowthChart)
- T007 and T008 can run in parallel (different files)
- T009 can start independently; T010 depends on T009
- T011 can start independently; T012 depends on T011
- T015, T016 can run in parallel (different files)

### Parallel Opportunities

```
Phase 1:  T001 ─┬─ (parallel)
          T002 ─┘

Phase 2:  T003 → T004  (sequential: TDD)

Phase 3:  T005 ─┬─ → T006  (T005 before T006)
          T007 ─┤   (parallel with T005)
          T008 ─┘   (parallel with T005)

Phase 4:  T009 → T010  (sequential)

Phase 5:  T011 → T012  (sequential)

Phase 6:  T013 ─┬─ (parallel)
          T014 ─┘

Phase 7:  T015 ─┬─ (parallel)
          T016 ─┤
          T017 ─┘
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (install recharts, add types)
2. Complete Phase 2: Foundational (projection engine + tests)
3. Complete Phase 3: User Story 1 (chart, page, route, nav)
4. **STOP and VALIDATE**: Child can see projected growth graph
5. Deploy/demo if ready — delivers core educational value

### Incremental Delivery

1. Setup + Foundational → Projection engine ready with tests
2. Add US1 → Child sees growth graph → Deploy (MVP!)
3. Add US2 → Child can explore scenarios → Deploy
4. Add US3 → Child reads plain-English explanation → Deploy
5. Add US4 → Child adjusts time horizons → Deploy
6. Polish → Edge cases, loading states, animation → Deploy

---

## Notes

- All monetary calculations use integer cents to avoid floating-point errors
- Projection engine is a pure function — no side effects, easily testable
- No backend changes — all data comes from existing authenticated endpoints
- Recharts is the only new dependency (React-native charting library)
- Scenario inputs are ephemeral (reset on page navigation)
- Constitution TDD requirement satisfied by T003/T004 (tests before implementation for financial math)
