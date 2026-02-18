# Tasks: Child Visual Themes

**Input**: Design documents from `/specs/017-child-visual-themes/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/api.md, research.md, quickstart.md

**Tests**: Included per constitution (Test-First Development principle). Backend tests use TDD; frontend has no test framework (manual verification only).

**Organization**: Tasks grouped by user story. US3 (theme definitions) → US2 (settings page) → US1 (theme selection) reflects execution dependencies.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Database migration for theme column

- [x] T001 [P] Create up migration to add theme column to children table in `backend/migrations/004_child_theme.up.sql` — `ALTER TABLE children ADD COLUMN theme TEXT;`
- [x] T002 [P] Create down migration to drop theme column in `backend/migrations/004_child_theme.down.sql` — `ALTER TABLE children DROP COLUMN theme;`
- [x] T003 Apply migration 004 to test database (`bankofdad_test`) so subsequent store tests can use the new column

---

## Phase 2: Foundational (Backend Data Layer)

**Purpose**: Backend store, handler, route, and auth changes that ALL user stories depend on

**CRITICAL**: No frontend theme work can begin until the API is ready

### Tests (TDD — write first, verify they fail)

- [x] T004 Write store test for `UpdateTheme` method in `backend/internal/store/child_test.go` — test updating theme to each valid slug (`sapling`, `piggybank`, `rainbow`) and reading it back via `GetByID`. Verify `Theme` field is `*string` and nil for new children.
- [x] T005 Write store test verifying `GetByID`, `GetByFamilyAndName`, and `ListByFamily` all scan and return the `Theme` field correctly in `backend/internal/store/child_test.go`
- [x] T006 Write handler test for `HandleUpdateTheme` in `backend/internal/family/handlers_test.go` — test valid theme update (200), invalid theme value (400), and non-child user (403). Follow existing handler test patterns.
- [x] T007 Write test verifying `HandleGetMe` includes `theme` field in child response in `backend/internal/auth/handlers_test.go`

### Implementation

- [x] T008 Add `Theme *string` field to `Child` struct in `backend/internal/store/child.go` — add after the existing `Avatar *string` field
- [x] T009 Update all SELECT queries and `.Scan()` calls in `backend/internal/store/child.go` — add `theme` column to `GetByID`, `GetByFamilyAndName`, `ListByFamily`, and `Create` methods. Scan into `sql.NullString` and convert to `*string`, matching the `avatar` pattern exactly.
- [x] T010 Add `UpdateTheme(childID int64, theme string) error` method to `ChildStore` in `backend/internal/store/child.go` — `UPDATE children SET theme = $1, updated_at = NOW() WHERE id = $2`, check `RowsAffected() > 0`
- [x] T011 Add `HandleUpdateTheme` method to family `Handlers` in `backend/internal/family/handlers.go` — extract child ID from auth context, validate `user_type == "child"`, validate theme against allowlist `["sapling", "piggybank", "rainbow"]`, call `childStore.UpdateTheme`, return `{"message": "Theme updated", "theme": "..."}`. Return 403 for non-child users, 400 for invalid theme.
- [x] T012 Include `theme` field in child response of `HandleGetMe` in `backend/internal/auth/handlers.go` — add `"theme": child.Theme` to the child user map (it will serialize as `null` when nil)
- [x] T013 Register route `PUT /api/child/settings/theme` with `requireAuth` middleware in `backend/main.go` — wire to `familyHandlers.HandleUpdateTheme`
- [x] T014 Update TRUNCATE list in test helper `backend/internal/store/store_test_helpers_test.go` if needed — ensure `children` table truncation uses `RESTART IDENTITY CASCADE` (already done for existing tests, verify theme column doesn't need special handling)
- [x] T015 Run `go test -p 1 ./...` from `backend/` and verify all tests pass (existing + new T004-T007 tests)

**Checkpoint**: Backend API is complete. `PUT /api/child/settings/theme` works, `/auth/me` returns theme for child users.

---

## Phase 3: User Story 3 - Theme Visual Design (Priority: P1)

**Goal**: Define three visually distinct themes with colors and SVG background patterns, and create the ThemeProvider that applies them globally.

**Independent Test**: Apply each theme programmatically and verify the correct CSS custom properties and background image are set on the document.

### Implementation

- [x] T016 [P] [US3] Add `theme` field to `ChildUser` interface in `frontend/src/types.ts` — `theme?: string | null;` (null/undefined = "sapling")
- [x] T017 [P] [US3] Add `updateChildTheme` API function in `frontend/src/api.ts` — `PUT /api/child/settings/theme` with `{ theme: string }` body, returning `{ message: string, theme: string }`
- [x] T018 [US3] Create theme definitions file `frontend/src/themes.ts` — export a `THEMES` object keyed by slug (`sapling`, `piggybank`, `rainbow`). Each entry: `{ label: string, colors: { forest, forestLight, cream, creamDark }, backgroundSvg: string }`. Use exact hex values from plan.md color table. Design three inline SVG data URIs: (1) Sapling: scattered small leaves at ~8% opacity in muted green, (2) Piggy Bank: faint coin circles at ~6% opacity in muted rose, (3) Rainbow: tiny scattered stars at ~6% opacity in muted purple. SVGs should tile as `background-repeat: repeat` patterns. Export `getTheme(slug: string | null | undefined)` helper that defaults to sapling.
- [x] T019 [US3] Create `ThemeProvider` context in `frontend/src/context/ThemeContext.tsx` — modeled after `TimezoneContext.tsx`. Provides `{ theme: string, setTheme: (slug: string) => void }`. On mount/change: applies CSS custom property overrides to `document.documentElement.style` for `--color-forest`, `--color-forest-light`, `--color-cream`, `--color-cream-dark` and sets `background-image` on `document.body.style`. On cleanup (unmount or parent user): removes all overrides to restore defaults. Export `useTheme()` hook.
- [x] T020 [US3] Wrap app with `ThemeProvider` in `frontend/src/App.tsx` — nest inside existing `TimezoneProvider`. ThemeProvider should read initial theme from the child's `/auth/me` response (pass through from pages or read independently).

**Checkpoint**: Theme system is functional. CSS custom properties and background SVGs apply when theme is set programmatically. All three themes have distinct, appealing visuals.

---

## Phase 4: User Story 2 - Child Settings Page Navigation (Priority: P2)

**Goal**: Add a Settings page for child users accessible from the sidebar/bottom tab navigation.

**Independent Test**: Log in as a child, verify "Settings" appears in nav, click it, see the settings page with category navigation layout.

### Implementation

- [x] T021 [US2] Create `ChildSettingsPage` component in `frontend/src/pages/ChildSettingsPage.tsx` — follow `SettingsPage.tsx` pattern exactly: auth guard (redirect if not child), `<Layout user={user} maxWidth="wide">`, page header with Settings icon, two-column layout with category nav sidebar (mobile horizontal tabs + desktop vertical pills). Single category: "Appearance" with `Palette` icon from lucide-react. Content area shows a placeholder card for now (theme picker added in US1).
- [x] T022 [US2] Add `/child/settings` route in `frontend/src/App.tsx` — `<Route path="/child/settings" element={<ChildSettingsPage />} />` between existing child routes
- [x] T023 [US2] Add "Settings" nav item for child users in `frontend/src/components/Layout.tsx` — add in both desktop sidebar and mobile bottom tab bar, after "Growth" and before logout. Use `Settings` icon (already imported). Route to `/child/settings`. Active state: `bg-forest text-white` (desktop) / `text-forest` (mobile), matching existing patterns.

**Checkpoint**: Child settings page is accessible and renders with the correct layout. Navigation shows Settings item for child users on both desktop and mobile.

---

## Phase 5: User Story 1 - Select a Visual Theme (Priority: P1)

**Goal**: Children can select a theme from the settings page, see it applied immediately, and have it persist across sessions.

**Independent Test**: Log in as child, go to Settings, select each theme — verify colors and background change. Log out and back in — verify theme persists.

**Dependencies**: Requires US3 (theme definitions + ThemeProvider) and US2 (settings page)

### Implementation

- [x] T024 [US1] Build theme picker UI in `ChildSettingsPage` (`frontend/src/pages/ChildSettingsPage.tsx`) — inside the "Appearance" category content area, display three theme cards in a responsive grid. Each card shows: theme name, a visual preview rectangle with the theme's background color and accent color, and a small sample of the SVG pattern. The currently active theme has a highlighted border (using the theme's accent color). Clicking a card calls `updateChildTheme` API, updates ThemeProvider via `setTheme`, and shows success/error feedback.
- [x] T025 [US1] Integrate ThemeProvider with `/auth/me` response in child pages — ensure `ThemeProvider` reads the `theme` field from the authenticated child user data. On initial load, apply the saved theme before rendering content. Handle null/undefined as "sapling" default.
- [x] T026 [US1] Handle theme cleanup on logout in `frontend/src/context/ThemeContext.tsx` — when user logs out (tokens cleared) or when a parent user is detected, remove all CSS custom property overrides and background image from document, restoring the default Sapling appearance.
- [x] T027 [US1] Handle unknown theme fallback — if a child's saved theme slug doesn't match any known theme (future-proofing), fall back to "sapling" gracefully in both `ThemeContext.tsx` and `ChildSettingsPage.tsx`.

**Checkpoint**: Full feature is functional. Children can select themes, see immediate visual changes, and preferences persist across sessions. Parents see no theme effects.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and edge case handling

- [x] T028 Run all backend tests from `backend/` with `go test -p 1 ./...` and verify 100% pass rate
- [x] T029 Run frontend dev server and manually verify all three themes per quickstart.md verification steps: (1) log in as child, (2) navigate to Settings, (3) select each theme and verify colors + background change, (4) log out and back in to verify persistence, (5) log in as parent to verify no theme effects
- [x] T030 Verify theme previews in settings page are visually appealing and clearly differentiated — adjust SVG patterns or colors in `frontend/src/themes.ts` if needed
- [x] T031 Verify mobile responsiveness — check settings page layout, theme cards, and nav item on small screens

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (migration applied to test DB)
- **US3 - Theme Visual Design (Phase 3)**: Depends on Phase 2 (needs types + API function, which need backend ready)
- **US2 - Settings Page Nav (Phase 4)**: Depends on Phase 2 (needs auth patterns). Can run in parallel with Phase 3.
- **US1 - Theme Selection (Phase 5)**: Depends on Phase 3 (ThemeProvider) AND Phase 4 (settings page)
- **Polish (Phase 6)**: Depends on all previous phases

### User Story Dependencies

- **US3 (P1)**: Can start after Foundational — provides theme infrastructure for US1
- **US2 (P2)**: Can start after Foundational — provides settings page container for US1. **Can run in parallel with US3.**
- **US1 (P1)**: Depends on BOTH US3 and US2 — combines theme system with settings page for full feature

### Within Each Phase

- Tests (T004-T007) MUST be written and FAIL before implementation (T008-T015)
- Store changes (T008-T010) before handler (T011) before route (T013)
- Frontend types (T016) and API (T017) before theme definitions (T018) before ThemeProvider (T019)
- Settings page (T021) before theme picker UI (T024)

### Parallel Opportunities

```
Phase 1:  T001 ║ T002        (parallel: different files)
Phase 2:  T004-T007          (sequential: TDD write tests first)
          T008 ║ T009 → T010 (T008 parallel with start of T009, T010 after T009)
          T011 → T012 → T013 (sequential: handler → auth → route)
Phase 3:  T016 ║ T017        (parallel: different files)
          T018 → T019 → T020 (sequential: themes → provider → wiring)
Phase 4:  T021 → T022 ║ T023 (page first, then route and nav in parallel)
Phase 5:  T024 → T025 → T026 → T027 (sequential within settings page)
Phase 6:  T028 ║ T029        (parallel: backend tests vs frontend manual test)
          T030 ║ T031        (parallel: visual polish vs mobile check)

Cross-phase parallelism:
  Phase 3 (T016-T020) ║ Phase 4 (T021-T023) — can run simultaneously after Phase 2
```

---

## Implementation Strategy

### MVP First (US3 + US2 + US1 core)

1. Complete Phase 1: Setup (migration)
2. Complete Phase 2: Foundational (backend API + tests)
3. Complete Phase 3: US3 (theme definitions + ThemeProvider)
4. Complete Phase 4: US2 (settings page + nav)
5. Complete Phase 5: US1 (theme picker + integration)
6. **STOP and VALIDATE**: Test full flow per quickstart.md

### Incremental Delivery

1. After Phase 2: Backend is production-ready (theme column exists, API works, /auth/me returns theme)
2. After Phase 3 + 4: Settings page visible, themes defined but no selection UI yet
3. After Phase 5: Full feature complete — children can browse, select, and persist themes
4. Phase 6: Polish pass for visual quality and edge cases

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Backend tests follow TDD per constitution — write tests first, verify failure, then implement
- Frontend has no test framework — verification is manual per quickstart.md
- Theme slugs: `sapling`, `piggybank`, `rainbow` (lowercase, no spaces)
- NULL in DB = `sapling` default (application layer handles this)
- CSS custom property overrides: `--color-forest`, `--color-forest-light`, `--color-cream`, `--color-cream-dark`
- SVG backgrounds: inline data URIs, low opacity (6-8%), repeating tile patterns
