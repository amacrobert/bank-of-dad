# Tasks: Routed Selections

**Input**: Design documents from `/specs/022-routed-selections/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Route Definitions)

**Purpose**: Update App.tsx with all parameterized routes that all user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T001 Add parameterized routes and settings redirect in `frontend/src/App.tsx`. Replace the three static parent routes with: (1) `Navigate` import from react-router-dom, (2) `/dashboard/:childName?` route for ParentDashboard, (3) `/growth/:childName?` route for GrowthPage, (4) `/settings` index route that renders `<Navigate to="/settings/general" replace />`, (5) `/settings/:category` route for SettingsPage, (6) `/settings/children/:childName` route for SettingsPage. Ensure the `/settings/children/:childName` route is listed before `/settings/:category` so it matches first. Keep all routes inside the existing `AuthenticatedLayout userType="parent"` wrapper.

**Checkpoint**: All parameterized routes defined. Existing pages still render (params are unused until their story is implemented).

---

## Phase 2: User Story 1 — Settings Category Navigation via URL (Priority: P1) 🎯 MVP

**Goal**: Replace `?tab=` query parameter with URL path segments for settings categories. `/settings` redirects to `/settings/general`. Invalid categories redirect to `/settings/general`.

**Independent Test**: Navigate to `/settings` → redirects to `/settings/general`. Click categories → URL updates. Refresh → stays on same category. Browser back → returns to previous category. `/settings/nonexistent` → redirects to `/settings/general`.

### Implementation for User Story 1

- [x] T002 [US1] Update SettingsPage to use URL params for category in `frontend/src/pages/SettingsPage.tsx`. Replace `useSearchParams` import with `useParams` and `useNavigate` from react-router-dom. Read `category` and `childName` from `useParams<{ category: string; childName?: string }>()`. Remove the `initialTab` variable and `activeCategory` state — derive `activeCategory` directly from the `category` URL param. Add a `useEffect` that validates the category param: if `category` is not in the CATEGORIES keys array, call `navigate("/settings/general", { replace: true })`. Replace both category button `onClick` handlers (mobile line 98 and desktop line 123) to call `navigate(\`/settings/${cat.key}\`)` instead of `setActiveCategory(cat.key)`. The conditional content rendering (`activeCategory === "general"`, etc.) should now reference the `category` param directly. Remove the `useSearchParams` import entirely.

**Checkpoint**: Settings category navigation is fully URL-driven. `?tab=` no longer used. Refresh and back/forward work for category selection.

---

## Phase 3: User Story 2 — Child Selection on Dashboard via URL (Priority: P1)

**Goal**: Child selection on dashboard persists in URL as `/dashboard/{firstName}`. Refresh preserves selection. Browser back/forward navigates between selections.

**Independent Test**: Select a child → URL shows `/dashboard/{name}`. Refresh → child still selected. Select different child → URL updates. Click same child → deselects, URL returns to `/dashboard`. Navigate to `/dashboard/nonexistent` → redirects to `/dashboard`.

### Implementation for User Story 2

- [x] T003 [P] [US2] Update ParentDashboard for URL-driven child selection in `frontend/src/pages/ParentDashboard.tsx`. Add `useParams` import from react-router-dom. Read `childName` from `useParams<{ childName?: string }>()`. Remove the `selectedChild` state (`useState<Child | null>(null)`). Add a derived `selectedChild` variable computed from the children array: if `childName` is set and children are loaded, find the first child where `child.first_name.toLowerCase() === childName.toLowerCase()`; otherwise `null`. Add a `useEffect` that watches `childName` and `children`: if `childName` is set but no match is found in a non-empty children array, call `navigate("/dashboard", { replace: true })`. Replace `setSelectedChild` in the ChildSelectorBar `onSelectChild` prop with a function that calls `navigate(\`/dashboard/${child.first_name.toLowerCase()}\`)` for selection or `navigate("/dashboard")` for deselection (when `child` is null). In the `useEffect` that fetches children (childRefreshKey), remove the `selectedChild` state update logic (the URL param now drives selection). Update the existing `navigate("/settings?tab=children")` call (line 91) to `navigate("/settings/children")`.

**Checkpoint**: Dashboard child selection is fully URL-driven. Works independently of settings routing.

---

## Phase 4: User Story 3 — Child Selection on Settings Children Page via URL (Priority: P2)

**Goal**: Selecting a child on the settings children page updates the URL to `/settings/children/{firstName}`. Refresh preserves selection. Invalid names redirect to `/settings/children`.

**Independent Test**: Navigate to `/settings/children` → select a child → URL shows `/settings/children/{name}`. Refresh → child still selected. Deselect → URL returns to `/settings/children`. Navigate to `/settings/children/nonexistent` → redirects to `/settings/children`.

**Depends on**: US1 (T002 — settings category routing must be in place)

### Implementation for User Story 3

- [x] T004 [US3] Convert ChildrenSettings to controlled child selection in `frontend/src/components/ChildrenSettings.tsx`. Add two new props to the component: `selectedChildName?: string` and `onChildSelect: (child: Child | null) => void`. Remove the internal `selectedChild` state. Add a derived `selectedChild` variable: if `selectedChildName` is set and children are loaded, find the first child where `child.first_name.toLowerCase() === selectedChildName.toLowerCase()`; otherwise `null`. Add a `useEffect` that watches `selectedChildName` and `children`: if `selectedChildName` is set but no match found in a non-empty children array, call `onChildSelect(null)`. Replace `setSelectedChild` in the ChildSelectorBar `onSelectChild` prop with `onChildSelect`. In the `handleChildDeleted` callback, replace `setSelectedChild(null)` with `onChildSelect(null)`. In the children fetch `useEffect`, remove the `selectedChild` state update logic (the prop now drives selection — the derived variable will auto-update when children list refreshes).

- [x] T005 [US3] Pass child name and selection callback from SettingsPage to ChildrenSettings in `frontend/src/pages/SettingsPage.tsx`. In the `activeCategory === "children"` conditional rendering, update `<ChildrenSettings />` to pass two props: `selectedChildName={childName}` (from the URL param already read in T002) and `onChildSelect` — a callback function that navigates to `/settings/children/${child.first_name.toLowerCase()}` when a child is selected, or `/settings/children` when deselected (child is null).

**Checkpoint**: Child selection on settings children page is fully URL-driven. Works as an extension of US1 settings routing.

---

## Phase 5: User Story 4 — Child Selection on Growth Page via URL (Priority: P2)

**Goal**: Child selection on growth page persists in URL as `/growth/{firstName}`. Same behavior pattern as dashboard (US2).

**Independent Test**: Select a child → URL shows `/growth/{name}`. Refresh → child still selected. Deselect → URL returns to `/growth`. Navigate to `/growth/nonexistent` → redirects to `/growth`.

### Implementation for User Story 4

- [x] T006 [P] [US4] Update GrowthPage for URL-driven child selection in `frontend/src/pages/GrowthPage.tsx`. Add `useParams` import from react-router-dom. Read `childName` from `useParams<{ childName?: string }>()`. Remove the `selectedChild` state (`useState<Child | null>(null)`). Add a derived `selectedChild` variable: if `childName` is set and children are loaded, find the first child where `child.first_name.toLowerCase() === childName.toLowerCase()`; otherwise `null`. Add a `useEffect` that watches `childName` and `children`: if `childName` is set but no match found in a non-empty children array, call `navigate("/growth", { replace: true })`. Replace `setSelectedChild` in the ChildSelectorBar `onSelectChild` prop (line 179) with a function that calls `navigate(\`/growth/${child.first_name.toLowerCase()}\`)` for selection or `navigate("/growth")` for deselection. Update the `childId` derivation to use the new derived `selectedChild` variable (should work as-is since the variable name is the same). Update the existing `navigate("/settings?tab=children")` call (line 160) to `navigate("/settings/children")`.

**Checkpoint**: Growth page child selection is fully URL-driven. Consistent pattern with dashboard.

---

## Phase 6: Polish & Verification

**Purpose**: Verify all changes compile and build correctly

- [x] T007 Run TypeScript type check and production build in `frontend/`. Execute `cd frontend && npx tsc --noEmit && npm run build` to verify no type errors and the production build succeeds. Also run `npm run lint` to check for linting issues. Fix any errors found.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — start immediately
- **US1 (Phase 2)**: Depends on Foundational (T001)
- **US2 (Phase 3)**: Depends on Foundational (T001) — independent of US1
- **US3 (Phase 4)**: Depends on US1 (T002) — extends settings routing with child selection
- **US4 (Phase 5)**: Depends on Foundational (T001) — independent of US1, US2, US3
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after T001 — No dependencies on other stories
- **US2 (P1)**: Can start after T001 — Independent of US1 (different file)
- **US3 (P2)**: Can start after T002 — Depends on US1 (extends SettingsPage.tsx)
- **US4 (P2)**: Can start after T001 — Independent of US1, US2, US3 (different file)

### Parallel Opportunities

After T001 completes:
- **US1 (T002)** and **US2 (T003)** and **US4 (T006)** can all run in parallel (different files)
- **US3 (T004, T005)** must wait for US1 (T002) since T005 modifies SettingsPage.tsx

```
T001 (App.tsx)
├── T002 (SettingsPage.tsx) ─── T004 (ChildrenSettings.tsx) ─── T005 (SettingsPage.tsx)
├── T003 (ParentDashboard.tsx)  [parallel]
└── T006 (GrowthPage.tsx)      [parallel]
                                └── T007 (verify build)
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. Complete T001: Route definitions
2. Complete T002: Settings category routing
3. **STOP and VALIDATE**: Settings categories work via URL, refresh, back/forward
4. Deploy/demo if ready

### Incremental Delivery

1. T001 → Foundation ready
2. T002 → Settings categories routed (MVP!)
3. T003 → Dashboard child selection routed
4. T004 + T005 → Settings children child selection routed
5. T006 → Growth page child selection routed
6. T007 → Build verification

### Fastest Parallel Path

1. T001 (foundational)
2. T002 + T003 + T006 in parallel (US1 + US2 + US4)
3. T004 + T005 sequentially (US3, after T002)
4. T007 (verification)

---

## Notes

- All changes are frontend-only — no backend or database modifications
- No new files or dependencies — only modifications to 5 existing files
- ChildSelectorBar component is NOT modified — it already supports controlled selection via props
- The `?tab=` query parameter is fully eliminated (no backward compatibility needed)
- Commit after each task or logical group
