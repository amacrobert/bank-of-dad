# Tasks: Child Auto-Setup

**Input**: Design documents from `/specs/026-child-auto-setup/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Not explicitly requested — test tasks omitted. Verification via `npx tsc --noEmit && npm run build && npm run lint`.

**Organization**: Both user stories (US1: Onboarding, US2: Children Settings) share the same `AddChildForm.tsx` component, so they are implemented together. The form is used identically in both contexts — no separate implementation needed.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No setup needed — this is a modification to an existing component using existing APIs. No new dependencies, no new files, no migrations.

*(Phase skipped — nothing to initialize)*

---

## Phase 2: Foundational

**Purpose**: No foundational work needed — all backend endpoints and frontend API functions already exist.

*(Phase skipped — no blocking prerequisites)*

---

## Phase 3: User Story 1 & 2 — Add Child with Full Setup (Priority: P1)

**Goal**: Add Initial Deposit, Weekly Allowance, and Annual Interest fields to the Add Child form. After child creation, call existing APIs to set up each configured item. This single implementation serves both user stories (onboarding wizard and Children settings) since they share `AddChildForm.tsx`.

**Independent Test**: Create a child with all three fields populated → verify child has correct balance, active weekly allowance, and active monthly interest schedule. Then create a child with all fields empty → verify behavior is unchanged from today.

### Implementation

- [x] T001 [US1] Add state variables for `initialDeposit`, `weeklyAllowance`, and `annualInterest` (all string type) plus `setupWarning` (string | null) to `frontend/src/components/AddChildForm.tsx`
- [x] T002 [US1] Add three input fields in a responsive grid row (`grid-cols-1 sm:grid-cols-3`) below the avatar picker in `frontend/src/components/AddChildForm.tsx`: Initial Deposit ($ prefix, `type="number"`, `step="0.01"`, `max="999999.99"`), Weekly Allowance ($ prefix, same pattern), Annual Interest (% suffix, `type="number"`, `step="0.01"`, `max="100"`)
- [x] T003 [US1] Add post-creation API orchestration in `handleSubmit` of `frontend/src/components/AddChildForm.tsx`: after `post("/children", ...)` succeeds, sequentially call `deposit()` (if initialDeposit > 0), `setChildAllowance()` (if weeklyAllowance > 0, frequency="weekly", day_of_week=current day, note="Weekly allowance"), and `setInterest()` (if annualInterest > 0, frequency="monthly", day_of_month=1). Track failures in `setupWarning` without losing the created child.
- [x] T004 [US1] Import `deposit`, `setChildAllowance`, `setInterest` from `../api` in `frontend/src/components/AddChildForm.tsx`
- [x] T005 [US1] Update the success confirmation card in `frontend/src/components/AddChildForm.tsx` to show what was configured: initial balance amount (if deposited), "Weekly allowance active" (if set), "Interest active" (if set). Show `setupWarning` in an amber/terracotta warning box if any setup step failed.
- [x] T006 [US1] Reset the three new fields (`setInitialDeposit("")`, `setWeeklyAllowance("")`, `setAnnualInterest("")`) on successful submission alongside existing field resets in `frontend/src/components/AddChildForm.tsx`

**Checkpoint**: Add Child form shows 3 new optional fields. Creating a child with values populated results in deposit + allowance + interest being configured. Creating a child with empty fields behaves identically to before. Works in both onboarding (SetupPage) and Children settings (ChildrenSettings) since both use AddChildForm.

---

## Phase 4: Polish & Cross-Cutting Concerns

- [x] T007 Run `cd frontend && npx tsc --noEmit && npm run build && npm run lint` to verify no type errors, build succeeds, and lint is clean
- [x] T008 Visual verification: confirm fields are properly aligned in a row on desktop and stack vertically on mobile viewports

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 & 2**: Skipped — no setup or foundational work needed
- **Phase 3 (US1 & US2)**: Can start immediately — all prerequisites (backend APIs, frontend API helpers) already exist
- **Phase 4 (Polish)**: Depends on Phase 3 completion

### Task Dependencies (Phase 3)

- T001 → T002 (state must exist before inputs reference it)
- T001 → T003 (state must exist before orchestration reads it)
- T003 → T004 (imports needed for API calls, but can be done together)
- T003 → T005 (orchestration must exist before success card references results)
- T001 → T006 (state must exist before resets)

### Parallel Opportunities

- T004 can be done alongside T001 (different sections of the file — imports vs state)
- T002 and T003 are independent (UI rendering vs submission logic) but both depend on T001
- T005 and T006 can be done in parallel after T003

---

## Implementation Strategy

### MVP First

1. Complete T001–T004: Form fields + API orchestration (core functionality)
2. **STOP and VALIDATE**: Test child creation with all 3 fields → verify backend state
3. Complete T005–T006: Success display + field resets (polish)
4. Complete T007–T008: Build verification + visual check

### Single-Developer Flow

Since this is a single-file change, execute T001 → T002 → T003 → T004 → T005 → T006 sequentially, then verify with T007 → T008. Total: ~8 tasks, 1 file modified.

---

## Notes

- All changes are in `frontend/src/components/AddChildForm.tsx` — no other files need modification
- API functions `deposit`, `setChildAllowance`, `setInterest` already exist in `frontend/src/api.ts`
- No backend changes — all existing endpoints are reused as-is
- Both US1 (onboarding) and US2 (Children settings) are satisfied by the same component change
- Currency conversion: `Math.round(parseFloat(amount) * 100)` for cents
- Rate conversion: `Math.round(parseFloat(rate) * 100)` for basis points
- Day of week default: `new Date().getDay()` (0=Sunday, matches backend expectation)
