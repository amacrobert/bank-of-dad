# Tasks: Child Management in Settings with Parent Onboarding

**Input**: Design documents from `/specs/018-child-management-settings/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: No frontend test framework in place. Tests not requested. All validation is manual.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Component Extraction)

**Purpose**: Extract the reusable `ChildAccountSettings` component from `ManageChild.tsx`. This is a blocking prerequisite — US1 needs this component for the Settings page, and US3 needs the extraction done before cleaning up the dashboard.

- [x] T001 Extract account settings UI (reset password form, update name/avatar form, delete account form) from `frontend/src/components/ManageChild.tsx` into new `frontend/src/components/ChildAccountSettings.tsx`. The component accepts `child: Child`, `onUpdated: () => void`, and `onDeleted: () => void` props. Move the password reset handler, name/avatar update handler, delete handler, and all related state (`newPassword`, `newName`, `newAvatar`, `passwordMsg`, `nameMsg`, `deleteConfirmName`, `deleting`, `error`) into the new component. Reuse existing `Card`, `Input`, `Button`, `AvatarPicker` imports. Each form calls existing API endpoints (`PUT /children/{id}/password`, `PUT /children/{id}/name`, `DELETE /children/{id}`).

**Checkpoint**: `ChildAccountSettings` renders independently and can be imported by other components.

---

## Phase 2: User Story 1 — Parent Manages Children from Settings (Priority: P1) 🎯 MVP

**Goal**: All child account management (add, edit, delete, reset password) available under Settings → Children.

**Independent Test**: Navigate to Settings → click "Children" tab → verify child list loads, can add a child, select a child to edit name/avatar, reset password, and delete account.

### Implementation for User Story 1

- [x] T002 [US1] Create `frontend/src/components/ChildrenSettings.tsx`. This component composes: `ChildList` (for selecting a child), `AddChildForm` (for creating new children, in a collapsible section), and `ChildAccountSettings` (for managing the selected child). Use a two-column layout on desktop (child list + add on left, account settings on right) similar to the current dashboard pattern. Include `childRefreshKey` state to trigger list refresh on add/update/delete. When a child is deleted, clear the selection and refresh the list.

- [x] T003 [US1] Add "Children" category to `frontend/src/pages/SettingsPage.tsx`. Add `{ key: "children", label: "Children", icon: Users }` to the `CATEGORIES` array (import `Users` from lucide-react). When `activeCategory === "children"`, render `<ChildrenSettings />`. Support URL search param `?tab=children` to allow deep-linking to the children tab (used by dashboard empty state). Default active category to URL param value if present, otherwise "general".

- [x] T004 [US1] Update `frontend/src/components/ChildList.tsx` empty state message. Change "Add your first child below!" to "Add your first child to get started." (context-neutral — works in both Settings and Dashboard contexts). Remove the specific directional reference since the component is now used in multiple locations.

**Checkpoint**: Settings → Children fully functional. Parents can add, edit, delete children, and reset passwords entirely from Settings. Existing dashboard still works (has duplicate functionality temporarily).

---

## Phase 3: User Story 2 — Parent Onboarding Includes Adding Children (Priority: P2)

**Goal**: New parents are guided through adding children during the setup flow, so they arrive at the dashboard with accounts ready.

**Independent Test**: Create a new parent account → complete family slug selection → verify step 2 shows add-children form → add a child → verify credentials shown → add another → click "Continue to Dashboard" → verify children appear on dashboard. Also test: skip step 2 → verify dashboard loads (empty state).

### Implementation for User Story 2

- [x] T005 [US2] Extend `frontend/src/pages/SetupPage.tsx` from 2-step to 3-step onboarding flow. After family creation (step 1), show a new step 2: "Add Your Children". This step includes: (a) `AddChildForm` for creating children with credentials shown after each creation, (b) a list of children already added during this session (store as local state array of `ChildCreateResponse`), (c) a "Continue to Dashboard" button (always visible, enabled even with 0 children), (d) a "Skip for now" text link for parents who want to skip. Update progress dots from 2 to 3. Step 3 remains the "You're all set!" confirmation with "Go to Dashboard" button. The add-children step is purely UI-driven — no gating on child count.

**Checkpoint**: Full onboarding flow works end-to-end. New parents can add children during setup. Skip works. Children appear on dashboard after onboarding.

---

## Phase 4: User Story 3 — Dashboard Shows Only Financial Management (Priority: P3)

**Goal**: Parent dashboard focuses exclusively on financial activity — no account administration UI.

**Independent Test**: Log in as parent with existing children → verify dashboard shows only balance, deposit/withdraw, transactions, allowance, interest. No add-child button, no name/avatar edit, no password reset, no delete. Also test with no children → verify empty state links to Settings → Children.

### Implementation for User Story 3

- [x] T006 [P] [US3] Remove account settings section from `frontend/src/components/ManageChild.tsx`. Delete the "Account Settings" collapsible button and the `showSettings` conditional block (reset password card, update name/avatar card, delete account card). Remove unused state variables: `newPassword`, `newName`, `newAvatar`, `passwordMsg`, `nameMsg`, `deleteConfirmName`, `deleting`, `showSettings`, `settingsRef`. Remove unused imports: `AvatarPicker`, `Trash2`, `ChevronDown`. Keep the `onClose` prop. The component now renders only: header, locked-account warning, balance card, deposit/withdraw, transactions, allowance form, interest form.

- [x] T007 [P] [US3] Remove add-child functionality from `frontend/src/pages/ParentDashboard.tsx`. Remove: `AddChildForm` import and usage, "Add a Child" collapsible button, `showAddChild` state, `addChildInitialized` ref, `handleChildAdded` function, and the `onLoaded` callback from `ChildList` that auto-expands the add form. Remove unused imports: `AddChildForm`, `ChevronDown`, `UserPlus`. Add empty state: when `ChildList` reports 0 children via a new `childCount` state, show a `Card` with message "No children yet" and a `Link` (or button using `navigate`) pointing to `/settings?tab=children` with text "Go to Settings → Children to add your first child."

- [x] T008 [US3] Verify the `is_locked` warning still displays correctly in `frontend/src/components/ManageChild.tsx` after the cleanup. The locked-account banner ("This account is locked. Reset the password to unlock it.") should remain visible on the dashboard but the reset-password action now lives in Settings. Update the banner text to: "This account is locked. Go to Settings → Children to reset the password." so parents know where to find the action.

**Checkpoint**: Dashboard is clean — financial management only. Empty state works. All account admin is exclusively in Settings.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final integration verification and edge case handling.

- [x] T009 Verify deep-link from dashboard empty state to Settings → Children works correctly. Navigate to `/settings?tab=children` and confirm the "Children" category is auto-selected and the children management panel loads. Test on both desktop (sidebar) and mobile (tabs).

- [x] T010 Manual end-to-end walkthrough of all acceptance scenarios from spec.md: (1) Settings CRUD operations, (2) onboarding with children, (3) onboarding skip, (4) dashboard with children shows only financial UI, (5) dashboard with no children shows empty state, (6) delete last child from Settings → dashboard shows empty state.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — start immediately
- **US1 Settings (Phase 2)**: Depends on Phase 1 (needs `ChildAccountSettings`)
- **US2 Onboarding (Phase 3)**: No dependency on Phase 1 or US1 — uses `AddChildForm` directly (already exists)
- **US3 Dashboard Cleanup (Phase 4)**: Depends on US1 (move functionality to Settings before removing from dashboard)
- **Polish (Phase 5)**: Depends on all user stories being complete

### User Story Dependencies

```
Phase 1: Foundational ──┬──> Phase 2: US1 (Settings) ──> Phase 4: US3 (Dashboard Cleanup)
                        │
                        └──> Phase 3: US2 (Onboarding) [independent]

All ──> Phase 5: Polish
```

- **US1 (P1)**: Depends on Phase 1 extraction. Must complete before US3.
- **US2 (P2)**: Independent — can run in parallel with US1. Only touches `SetupPage.tsx`.
- **US3 (P3)**: Depends on US1 — functionality must exist in Settings before removing from dashboard.

### Parallel Opportunities

- **T002 + T005**: US1 implementation and US2 implementation can proceed in parallel after Phase 1 (they touch different files)
- **T006 + T007**: Within US3, ManageChild cleanup and ParentDashboard cleanup can proceed in parallel (different files)

---

## Parallel Example: After Phase 1

```bash
# These can run simultaneously after T001 completes:
Task: "T002 [US1] Create ChildrenSettings component"
Task: "T005 [US2] Extend SetupPage onboarding flow"
```

## Parallel Example: Within Phase 4

```bash
# These can run simultaneously:
Task: "T006 [US3] Remove account settings from ManageChild"
Task: "T007 [US3] Remove add-child from ParentDashboard"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Extract `ChildAccountSettings`
2. Complete Phase 2: Build Settings → Children page
3. **STOP and VALIDATE**: Test child CRUD from Settings independently
4. Settings page is fully functional — dashboard still has duplicate controls (safe interim state)

### Incremental Delivery

1. Phase 1 → Component extraction done
2. Phase 2 (US1) → Settings → Children works → Validate
3. Phase 3 (US2) → Onboarding includes children → Validate (can run parallel with US1)
4. Phase 4 (US3) → Dashboard cleaned up → Validate (must follow US1)
5. Phase 5 → Polish and final verification

---

## Notes

- This is a **frontend-only** feature — no backend, database, or API changes
- No new dependencies needed — all existing packages sufficient
- `AddChildForm` is reused as-is (no modifications needed)
- `AvatarPicker` is reused as-is (no modifications needed)
- The interim state between US1 and US3 (settings + dashboard both have child management) is safe — removing from dashboard happens in US3
- Total: 10 tasks across 5 phases
