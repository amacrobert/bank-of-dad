# Tasks: Child Selector Redesign

**Input**: Design documents from `/specs/021-child-selector-redesign/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Not included (no frontend test infrastructure; justified deviation per constitution check).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Web app**: `frontend/src/` for all changes (frontend-only feature)

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: Create the shared ChildSelectorBar component that all user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T001 Create ChildSelectorBar component with chip rendering, selection toggle, and loading state in `frontend/src/components/ChildSelectorBar.tsx`
  - Props: `children: Child[]`, `selectedChildId: number | null`, `onSelectChild: (child: Child | null) => void`, `loading?: boolean`
  - Render each child as a horizontal chip button with: avatar (emoji or first-letter initial), first name (truncated with ellipsis via `truncate` class), balance (small `BalanceDisplay`), and lock icon (`lucide-react Lock`) if `is_locked`
  - Selected chip: `ring-2 ring-forest bg-forest/5` (matches existing ChildList pattern)
  - Unselected chip: `bg-white border border-sand hover:bg-cream-dark`
  - Toggle behavior: clicking selected chip calls `onSelectChild(null)` to deselect (FR-012)
  - Use `aria-pressed` on each chip button for accessibility
  - Horizontal flex container with `overflow-x: auto` for native scrolling (FR-005)
  - CSS gradient fade masks on left/right edges when content overflows to indicate scrollable area (research R1)
  - Minimum 44px chip height for mobile touch targets (SC-005)
  - Loading state shows `LoadingSpinner variant="inline"`
  - Follow existing design language: `rounded-xl`, forest/cream/sand palette, `transition-all duration-200`

**Checkpoint**: ChildSelectorBar component ready for integration into pages

---

## Phase 2: User Story 1 - Parent selects a child on the Dashboard (Priority: P1) MVP

**Goal**: Replace Dashboard two-column layout with horizontal selector above full-width content

**Independent Test**: Log in as a parent with 1-3 children, verify horizontal selector appears above content, selecting a child shows full-width management panel, clicking selected chip deselects

### Implementation for User Story 1

- [x] T002 [US1] Refactor ParentDashboard to replace two-column grid with ChildSelectorBar above full-width content in `frontend/src/pages/ParentDashboard.tsx`
  - Remove `md:grid md:grid-cols-[340px_1fr] md:gap-6` two-column layout
  - Add child data fetching (move from ChildList): `get<ChildListResponse>("/children")` with `refreshKey` dependency
  - Render ChildSelectorBar with fetched children, selectedChild state, and loading state
  - Render ManageChild at full width below selector when a child is selected
  - Keep existing 0-children empty state card (with "No children yet" guidance)
  - Keep existing family info header unchanged
  - Replace ChildList Card import with ChildSelectorBar import
  - Update selectedChild handling: when `onSelectChild` receives `null`, clear selection (toggle-off)
  - When child is updated via ManageChild, refresh children list to get updated balance

- [x] T003 [US1] Remove close button and header bar from ManageChild in `frontend/src/components/ManageChild.tsx`
  - Remove the `onClose` prop from ManageChildProps interface
  - Remove the header div containing the X close button (`<button onClick={onClose}>`)
  - Keep the "Manage {child.first_name}" heading but simplify it (no flex justify-between needed)
  - Remove `onClose` from ParentDashboard's ManageChild usage (done in T002)

**Checkpoint**: Dashboard shows horizontal chip selector, full-width content panel, toggle deselection works

---

## Phase 3: User Story 2 - Parent selects a child on Settings > Children (Priority: P2)

**Goal**: Replace Settings > Children two-column layout with selector between AddChildForm and account settings

**Independent Test**: Navigate to Settings > Children, verify AddChildForm at full width, selector below, selecting a child shows full-width account settings

### Implementation for User Story 2

- [x] T004 [US2] Refactor ChildrenSettings to replace two-column grid with stacked layout using ChildSelectorBar in `frontend/src/components/ChildrenSettings.tsx`
  - Remove `md:grid md:grid-cols-[300px_1fr] md:gap-6` two-column layout
  - Add child data fetching: `get<ChildListResponse>("/children")` with `refreshKey` dependency
  - Layout: AddChildForm (full width) → ChildSelectorBar (full width) → ChildAccountSettings (full width, when selected)
  - Handle `onChildAdded`: increment refreshKey so selector re-renders with new child (FR-009)
  - Handle `onChildDeleted`: clear selectedChild to null, increment refreshKey (edge case)
  - Handle `onChildUpdated`: increment refreshKey so selector shows updated name/avatar
  - Replace ChildList Card import with ChildSelectorBar import
  - Keep empty state: when no child selected, show prompt text at full width (not hidden on mobile)

**Checkpoint**: Settings > Children shows AddChildForm + horizontal selector + full-width settings, child add/delete dynamically updates selector

---

## Phase 4: User Story 3 - Selector handles varying family sizes (Priority: P2)

**Goal**: Ensure the selector scales gracefully from 0 to 12 children without layout degradation

**Independent Test**: Create families of sizes 0, 1, 4, 8, 12 and verify selector is usable at each size on both desktop and mobile viewports

### Implementation for User Story 3

- [x] T005 [US3] Refine ChildSelectorBar chip layout for single-child and large-family edge cases in `frontend/src/components/ChildSelectorBar.tsx`
  - Ensure chips are left-aligned (not centered) so 1-child families don't look awkward (use `justify-start` on flex container)
  - Verify chip `max-width` and `truncate` handle long names (20+ chars) gracefully (FR-011)
  - Ensure fade indicators only appear when there IS overflow (use scroll position detection or CSS-only approach)
  - Verify the selector row height stays fixed regardless of number of children (single row, no wrapping: `flex-nowrap`)
  - Confirm `gap` between chips provides comfortable spacing at all sizes

- [x] T006 [US3] Ensure 0-children state is handled correctly on both pages
  - ParentDashboard (`frontend/src/pages/ParentDashboard.tsx`): when `children.length === 0`, hide ChildSelectorBar entirely and show existing empty state card guiding parent to Settings > Children
  - ChildrenSettings (`frontend/src/components/ChildrenSettings.tsx`): when `children.length === 0`, hide ChildSelectorBar (AddChildForm is still visible above)

**Checkpoint**: Selector looks correct and is usable at 0, 1, 4, 8, and 12 children on desktop and mobile

---

## Phase 5: User Story 4 - Reusable selector for future features (Priority: P3)

**Goal**: Confirm the ChildSelectorBar is a clean, reusable component with no page-specific logic

**Independent Test**: Verify both Dashboard and Settings > Children use the identical ChildSelectorBar component with the same props interface and identical visual behavior

### Implementation for User Story 4

- [x] T007 [US4] Verify ChildSelectorBar reusability: confirm both pages import from the same component path and use consistent props pattern in `frontend/src/components/ChildSelectorBar.tsx`, `frontend/src/pages/ParentDashboard.tsx`, `frontend/src/components/ChildrenSettings.tsx`
  - Both pages must import `ChildSelectorBar` from the same file
  - Both pages must pass: `children`, `selectedChildId`, `onSelectChild`, `loading`
  - No page-specific branching or props inside ChildSelectorBar
  - Component should work on any page by providing children data and a selection callback

**Checkpoint**: Same component used identically on both pages; a new page could adopt it by providing the same props

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Cleanup and validation across all user stories

- [x] T008 Remove unused ChildList references and clean up deprecated code in `frontend/src/components/ChildList.tsx`
  - Verify ChildList is no longer imported by any file (ParentDashboard, ChildrenSettings were the only consumers)
  - If no remaining imports, delete `frontend/src/components/ChildList.tsx`
  - If still imported elsewhere, keep file but remove from modified pages

- [x] T009 Run TypeScript type check and lint to validate all changes in `frontend/`
  - Run `cd frontend && npx tsc --noEmit` — fix any type errors
  - Run `cd frontend && npm run lint` — fix any lint warnings
  - Run `cd frontend && npm run build` — verify production build succeeds

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — can start immediately. BLOCKS all user stories.
- **US1 Dashboard (Phase 2)**: Depends on Phase 1 completion
- **US2 Settings (Phase 3)**: Depends on Phase 1 completion. Can run in parallel with US1.
- **US3 Family Sizes (Phase 4)**: Depends on Phase 1 (T005) and Phase 2+3 (T006) completion
- **US4 Reusability (Phase 5)**: Depends on Phase 2 and Phase 3 completion
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Depends only on Foundational. No cross-story dependencies.
- **US2 (P2)**: Depends only on Foundational. Independent from US1.
- **US3 (P2)**: T005 depends on Foundational. T006 depends on US1 + US2 (both pages must exist).
- **US4 (P3)**: Depends on US1 + US2 (both pages must use the component to verify reusability).

### Within Each User Story

- Page refactor before ManageChild cleanup (T002 before T003)
- Component refinements after initial integration (T005 after T001)
- Verification after all implementations (T007 after T002 + T004)

### Parallel Opportunities

- **T002 [US1] and T004 [US2]** can run in parallel (different files, both depend only on T001)
- **T003 [US1]** can run in parallel with T004 [US2] (different files)
- **T005 [US3] and T006 [US3]** can run in parallel (different concerns: component vs pages)

---

## Parallel Example: User Stories 1 & 2

```bash
# After T001 (Foundational) completes, launch US1 and US2 in parallel:
Task: "T002 [US1] Refactor ParentDashboard in frontend/src/pages/ParentDashboard.tsx"
Task: "T004 [US2] Refactor ChildrenSettings in frontend/src/components/ChildrenSettings.tsx"

# Then T003 can follow T002:
Task: "T003 [US1] Remove close button from ManageChild in frontend/src/components/ManageChild.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Create ChildSelectorBar component (T001)
2. Complete Phase 2: Integrate into Dashboard (T002, T003)
3. **STOP and VALIDATE**: Test Dashboard with 1-3 children — full-width layout, chip selection, toggle-off
4. Deploy/demo if ready

### Incremental Delivery

1. T001 → ChildSelectorBar component ready
2. T002 + T003 → Dashboard redesigned (MVP!)
3. T004 → Settings > Children redesigned
4. T005 + T006 → Overflow and edge cases polished
5. T007 → Reusability verified
6. T008 + T009 → Cleanup and validation

### Sequential Solo Developer Strategy

1. T001 → T002 → T003 → T004 → T005 → T006 → T007 → T008 → T009
2. Each task builds on the previous, commit after each logical group
3. Stop after T003 for MVP demo opportunity

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- No backend changes, no database migrations, no new API endpoints
- Balance IS shown on chips (per research R3)
- Chips use existing BalanceDisplay component with size="small"
- Follow existing patterns: AvatarPicker for selection UX, ChildList for data fetching
- Manual acceptance testing per spec scenarios (no automated frontend tests)
- Commit after each task or logical group
