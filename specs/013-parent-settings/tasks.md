# Tasks: Parent Settings Page

**Input**: Design documents from `/specs/013-parent-settings/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/settings-api.md

**Tests**: Included per constitution principle I (Test-First Development).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Database migration for timezone column

- [x] T001 Create migration `backend/migrations/003_family_timezone.up.sql` — `ALTER TABLE families ADD COLUMN timezone TEXT NOT NULL DEFAULT 'America/New_York'`
- [x] T002 [P] Create migration `backend/migrations/003_family_timezone.down.sql` — `ALTER TABLE families DROP COLUMN timezone`
- [x] T003 Run migration against dev and test databases to verify it applies cleanly

---

## Phase 2: Foundational (Backend Data & API Layer)

**Purpose**: Backend store methods, handlers, and routes that MUST be complete before any frontend work

**⚠️ CRITICAL**: No frontend user story work can begin until this phase is complete

### Tests (write first, verify they fail)

- [x] T004 Write store-level tests for `GetTimezone(familyID)` and `UpdateTimezone(familyID, timezone)` in `backend/internal/store/family_test.go` — test default value is `America/New_York`, test updating to valid timezone, test that invalid timezone is rejected at handler level (store does not validate)
- [x] T005 [P] Write handler tests for `GET /api/settings` in `backend/internal/settings/handlers_test.go` — test returns `{"timezone": "America/New_York"}` for a family with default timezone, test 401 for unauthenticated requests
- [x] T006 [P] Write handler tests for `PUT /api/settings/timezone` in `backend/internal/settings/handlers_test.go` — test valid timezone update returns 200 with `{"message": "Timezone updated", "timezone": "America/Chicago"}`, test invalid timezone returns 400, test empty body returns 400

### Implementation

- [x] T007 Update `Family` struct in `backend/internal/store/family.go` — add `Timezone string` field, update all `SELECT` queries (`GetByID`, `GetBySlug`) and `Scan` calls to include `timezone` column
- [x] T008 Add `GetTimezone(familyID int64) (string, error)` method to `FamilyStore` in `backend/internal/store/family.go` — returns timezone string for given family, defaults handled by DB column default
- [x] T009 Add `UpdateTimezone(familyID int64, timezone string) error` method to `FamilyStore` in `backend/internal/store/family.go` — `UPDATE families SET timezone = $1 WHERE id = $2`
- [x] T010 Create `backend/internal/settings/handlers.go` — define `Handlers` struct with `familyStore` dependency, `NewHandlers` constructor, and `writeJSON` helper (follow existing handler patterns from `family/handlers.go`)
- [x] T011 Implement `HandleGetSettings` in `backend/internal/settings/handlers.go` — get `familyID` from JWT context via `middleware.GetFamilyID(r)`, call `familyStore.GetTimezone(familyID)`, return `{"timezone": "..."}` as JSON
- [x] T012 Implement `HandleUpdateTimezone` in `backend/internal/settings/handlers.go` — decode request body `{"timezone": "..."}`, validate with `time.LoadLocation()`, call `familyStore.UpdateTimezone(familyID, timezone)`, return `{"message": "Timezone updated", "timezone": "..."}` on success or 400 on invalid timezone
- [x] T013 Register settings routes in `backend/main.go` — create `settingsHandlers := settings.NewHandlers(familyStore)`, add `GET /api/settings` and `PUT /api/settings/timezone` with `requireParent` middleware
- [x] T014 Run `go test -p 1 ./...` from `backend/` to verify all store and handler tests pass

**Checkpoint**: Backend API fully functional — `GET /api/settings` returns timezone, `PUT /api/settings/timezone` validates and persists. All tests green.

---

## Phase 3: User Story 1 — Access Settings Page (Priority: P1) 🎯 MVP

**Goal**: Parent can navigate to `/settings` and see a settings page with category-based navigation showing "General" selected

**Independent Test**: Log in as parent, navigate to `/settings`, verify page loads with "General" category active and content area visible. Log in as child, navigate to `/settings`, verify redirect.

### Implementation

- [x] T015 [P] [US1] Add `SettingsResponse` interface to `frontend/src/types.ts` — `{ timezone: string }`
- [x] T016 [P] [US1] Add `getSettings()` and `updateTimezone(timezone: string)` API functions to `frontend/src/api.ts` — `get<SettingsResponse>("/settings")` and `put<{message: string, timezone: string}>("/settings/timezone", { timezone })`
- [x] T017 [US1] Create `frontend/src/pages/SettingsPage.tsx` — parent-only page with `Layout` wrapper, category sidebar/tab navigation (desktop: sidebar, mobile: tabs), "General" category selected by default. Define categories as a static config array `{ key: string, label: string, icon: LucideIcon, component: React.ComponentType }[]` with only "General" populated. Redirect non-parent users. Load settings via `getSettings()` on mount.
- [x] T018 [US1] Add `/settings` route to `frontend/src/App.tsx` — `<Route path="/settings" element={<SettingsPage />} />` (place before the `/:familySlug` catch-all route)

**Checkpoint**: Settings page loads at `/settings` with "General" category. Category nav architecture is extensible. Non-parents are redirected.

---

## Phase 4: User Story 2 — Set Family Timezone (Priority: P1)

**Goal**: Parent can view current timezone, search/select a new one, save it, and see confirmation

**Independent Test**: Navigate to settings, verify default timezone shows "US Eastern Time", change to "US Pacific Time", save, refresh page, verify Pacific persists.

### Implementation

- [x] T019 [P] [US2] Create `frontend/src/components/TimezoneSelect.tsx` — searchable timezone selector component. Embed a curated list of ~40 common IANA timezones with human-friendly labels (e.g., `{ value: "America/New_York", label: "US Eastern Time (America/New_York)" }`). Include a text filter input that searches both label and IANA value. Props: `value: string`, `onChange: (tz: string) => void`. Use existing UI styling patterns (rounded-xl, border-sand, text-bark, etc.).
- [x] T020 [US2] Integrate timezone setting into `SettingsPage.tsx` General section — display `TimezoneSelect` with current timezone from `getSettings()`, add a "Save" button, call `updateTimezone()` on save, show success toast/message on 200, show error message on failure. Disable save button when no changes detected.

**Checkpoint**: Full timezone read/update flow works end-to-end. Default is US Eastern, changes persist, success/error feedback shown.

---

## Phase 5: User Story 3 — Settings Navigation Entry Point (Priority: P2)

**Goal**: Parent sees a settings icon in the nav that links to `/settings`. Children do not see it.

**Independent Test**: Log in as parent, verify gear icon visible in both desktop and mobile nav, click it, arrive at `/settings`. Log in as child, verify no gear icon.

### Implementation

- [x] T021 [US3] Add settings navigation to `frontend/src/components/Layout.tsx` — import `Settings` icon from `lucide-react`. In desktop nav: add a settings button between the display name and logout button (gear icon, same styling as logout link). In mobile bottom nav: add a settings tab between Dashboard and Log out (gear icon + "Settings" label, same styling as other tabs). Both only rendered when `user.user_type === "parent"`. Navigate to `/settings` on click.

**Checkpoint**: Settings is discoverable from any page. Parent-only visibility confirmed.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [x] T022 Run `go test -p 1 ./...` from `backend/` — confirm all tests pass including new store and handler tests
- [ ] T023 Run manual end-to-end validation per `specs/013-parent-settings/quickstart.md` — verify full flow: nav entry → settings page → timezone change → persist → reload
- [x] T024 Verify child user cannot access `/settings` route or see settings nav icon

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (migration must exist before tests run)
- **US1 (Phase 3)**: Depends on Phase 2 (backend API must be functional)
- **US2 (Phase 4)**: Depends on Phase 3 (settings page must exist to add timezone to it)
- **US3 (Phase 5)**: Depends on Phase 3 (settings page must exist to link to it) — can run in parallel with Phase 4
- **Polish (Phase 6)**: Depends on all previous phases

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational only — creates the settings page shell
- **US2 (P1)**: Depends on US1 — adds timezone selector to the General category on the settings page
- **US3 (P2)**: Depends on US1 — adds nav link pointing to the settings page. Can run in parallel with US2.

### Within Each Phase

- Tests written and verified to FAIL before implementation (Phase 2)
- Store layer before handler layer (Phase 2)
- Types/API before page component (Phase 3)
- Component before integration (Phase 4)

### Parallel Opportunities

- **Phase 1**: T001 and T002 can run in parallel (different files)
- **Phase 2 tests**: T005 and T006 can run in parallel (same file but independent test functions — write together)
- **Phase 3**: T015 and T016 can run in parallel (different files)
- **Phase 4 + 5**: T019 (TimezoneSelect) and T021 (Layout nav) can run in parallel (different files, no dependency)

---

## Parallel Example: Phase 3 (US1)

```
# These can run in parallel (different files):
Task: T015 — Add SettingsResponse to frontend/src/types.ts
Task: T016 — Add settings API functions to frontend/src/api.ts

# Then sequentially (depends on T015 + T016):
Task: T017 — Create SettingsPage.tsx
Task: T018 — Add /settings route to App.tsx
```

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Complete Phase 1: Setup (migration)
2. Complete Phase 2: Foundational (backend API + tests)
3. Complete Phase 3: US1 — Settings page with category nav
4. Complete Phase 4: US2 — Timezone selection
5. **STOP and VALIDATE**: Full timezone read/update flow works
6. Deploy/demo if ready

### Full Delivery

7. Complete Phase 5: US3 — Nav entry point
8. Complete Phase 6: Polish & validation
9. Feature complete

---

## Notes

- Total tasks: **24**
- Tasks per user story: US1 = 4, US2 = 2, US3 = 1
- Foundational tasks: 11 (setup + backend)
- Polish tasks: 3
- Parallel opportunities: 4 groups identified
- MVP scope: Phases 1-4 (setup + backend + settings page + timezone)
- All tasks follow `- [ ] [ID] [P?] [Story?] Description with file path` format
