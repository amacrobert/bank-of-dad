# Tasks: Child Avatars

**Input**: Design documents from `/specs/010-child-avatars/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.md

**Tests**: Included — constitution requires test-first development for all store mutations.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Database + Type Changes)

**Purpose**: Schema migration and type updates that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T001 Add `avatar TEXT` column migration using `addColumnIfNotExists` in `backend/internal/store/sqlite.go`
- [X] T002 Add `Avatar *string` field to `Child` struct and update `GetByID`, `GetByFamilyAndName`, `ListByFamily` to SELECT and Scan avatar (using `sql.NullString`) in `backend/internal/store/child.go`
- [X] T003 Add `avatar` field to `HandleListChildren` response struct (`childResponse`) in `backend/internal/family/handlers.go`
- [X] T004 [P] Add `avatar?: string` to `Child` and `ChildCreateResponse` interfaces in `frontend/src/types.ts`

**Checkpoint**: Avatar column exists, backend reads/returns avatar, frontend types accept avatar

---

## Phase 2: User Story 1 — Set Avatar When Creating a Child (Priority: P1) 🎯 MVP

**Goal**: Parents can optionally select an emoji avatar when adding a new child

**Independent Test**: Create a child with an avatar selected and verify the avatar is stored and returned in the API response

### Tests for User Story 1

- [X] T005 [US1] Test `ChildStore.Create` with avatar: create child with avatar, verify avatar returned by `GetByID`; create child without avatar, verify `Avatar` is nil — in `backend/internal/store/child_test.go`

### Implementation for User Story 1

- [X] T006 [US1] Update `ChildStore.Create` to accept optional avatar parameter and include in INSERT statement in `backend/internal/store/child.go`
- [X] T007 [US1] Parse optional `avatar` field in `HandleCreateChild` request struct and pass to `Create`; include avatar in create response in `backend/internal/family/handlers.go`
- [X] T008 [US1] Create `AvatarPicker` component with emoji grid (16 emojis), tap-to-select, tap-again-to-deselect, selected highlight in `frontend/src/components/AvatarPicker.tsx`
- [X] T009 [US1] Integrate `AvatarPicker` into `AddChildForm` below the name input; send `avatar` field in POST request; reset avatar on form clear in `frontend/src/components/AddChildForm.tsx`

**Checkpoint**: Parent can create a child with an emoji avatar. Avatar is stored and returned by the API.

---

## Phase 3: User Story 2 — Display Avatar in Child List (Priority: P1)

**Goal**: Avatar emoji shown in child selection list instead of first-letter initial

**Independent Test**: View the child list with a child that has an avatar set — emoji is displayed. View a child without avatar — first letter is displayed.

### Implementation for User Story 2

- [X] T010 [US2] Update avatar circle in `ChildList` to conditionally show `child.avatar` emoji (if set) or `child.first_name.charAt(0)` fallback in `frontend/src/components/ChildList.tsx`

**Checkpoint**: Child list shows emoji avatars for children that have them, first-letter initials for those that don't.

---

## Phase 4: User Story 3 — Update Avatar in Account Settings (Priority: P2)

**Goal**: Parents can change or remove a child's avatar via the renamed "Update Name and Avatar" form

**Independent Test**: Open account settings for an existing child, change their avatar, submit, and verify the new avatar appears in the child list

### Tests for User Story 3

- [X] T011 [US3] Test avatar update in store: update avatar only (name unchanged), update name only (avatar preserved), clear avatar (set to nil) — in `backend/internal/store/child_test.go`

### Implementation for User Story 3

- [X] T012 [US3] Extend `ChildStore.UpdateName` to accept optional avatar parameter and update avatar column (nil = leave unchanged, empty string = clear, value = set) in `backend/internal/store/child.go`
- [X] T013 [US3] Parse optional `avatar` field in `HandleUpdateName` request struct; pass to store; include avatar in update response; update audit log details in `backend/internal/family/handlers.go`
- [X] T014 [US3] Rename "Update Name" to "Update Name and Avatar" in ManageChild; add `AvatarPicker` pre-selecting current avatar; send avatar in PUT request in `frontend/src/components/ManageChild.tsx`

**Checkpoint**: Parent can change or clear a child's avatar from account settings. All changes persist and display correctly.

---

## Phase 5: Polish & Verification

**Purpose**: Final validation across all stories

- [X] T015 Run full backend test suite (`go test ./...`) and verify all pass
- [X] T016 [P] Verify frontend TypeScript compilation (`npx tsc --noEmit`) passes with no errors

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — start immediately. BLOCKS all user stories.
- **User Story 1 (Phase 2)**: Depends on Phase 1 completion
- **User Story 2 (Phase 3)**: Depends on Phase 1 completion (independent of US1, but best done after US1 to have test data)
- **User Story 3 (Phase 4)**: Depends on Phase 1 completion (uses AvatarPicker from US1)
- **Polish (Phase 5)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after Phase 1. Creates the `AvatarPicker` component reused by US3.
- **US2 (P1)**: Can start after Phase 1. No dependency on US1 (reads avatar from backend which is set up in Phase 1).
- **US3 (P2)**: Best done after US1 (reuses `AvatarPicker` component from T008). Backend tasks (T011-T013) can start after Phase 1 independently.

### Within Each User Story

- Tests written before implementation (constitution: test-first)
- Store changes before handler changes
- Backend before frontend
- Shared components (AvatarPicker) before form integration

### Parallel Opportunities

- T003 and T004 can run in parallel (backend handler vs frontend types)
- T005 (test) and T008 (AvatarPicker) can run in parallel (different files, backend vs frontend)
- T010 (US2) can run in parallel with US3 tasks (different files, independent stories)
- T015 and T016 can run in parallel (backend tests vs frontend type check)

---

## Parallel Example: Phase 1

```bash
# After T001 and T002 (sequential — schema before struct):
# Launch in parallel:
Task: "T003 - Add avatar to HandleListChildren response in handlers.go"
Task: "T004 - Add avatar to TypeScript interfaces in types.ts"
```

## Parallel Example: After Phase 1

```bash
# US1 backend and US2 frontend can overlap:
Task: "T005 - Store test for Create with avatar"
Task: "T010 - ChildList avatar display"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Foundational (4 tasks)
2. Complete Phase 2: US1 — avatar selection on create (5 tasks)
3. Complete Phase 3: US2 — avatar display (1 task)
4. **STOP and VALIDATE**: Create a child with avatar, verify it displays in the list
5. Deploy/demo if ready — avatars work end-to-end

### Incremental Delivery

1. Phase 1 → Foundation ready
2. Phase 2 (US1) + Phase 3 (US2) → Avatars work for new children (MVP!)
3. Phase 4 (US3) → Existing children can get avatars too
4. Phase 5 → Final verification

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Login page (`FamilyLogin.tsx`) is NOT modified — no per-child avatars on that screen (see research.md R4)
- No new API endpoints — existing create, list, and update endpoints are extended
- AvatarPicker is created in US1 (T008) and reused in US3 (T014)
- Commit after each task or logical group
