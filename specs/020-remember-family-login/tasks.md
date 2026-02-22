# Tasks: Remember Family Login

**Input**: Design documents from `/specs/020-remember-family-login/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, quickstart.md

**Tests**: No frontend test framework exists. Verification is manual browser testing + TypeScript type checking.

**Organization**: Tasks are grouped by user story. US4 (store slug) precedes US1 (redirect) since redirect depends on the slug being stored.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the localStorage utility that all user stories depend on

- [x] T001 Create family preference localStorage utility in `frontend/src/utils/familyPreference.ts` — export `getFamilySlug(): string | null`, `setFamilySlug(slug: string): void`, `clearFamilySlug(): void` using localStorage key `family_slug`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extract the GoogleSignInButton into a shared component so it can be used on both HomePage and FamilyLogin

**⚠️ CRITICAL**: US2 (parent login on family page) depends on this extraction being complete

- [x] T002 Extract `GoogleSignInButton` component and Google SVG icon from `frontend/src/pages/HomePage.tsx` into new file `frontend/src/components/GoogleSignInButton.tsx` — preserve existing `size` prop (`"default" | "compact" | "large"`), Google login URL construction (`` `${import.meta.env.VITE_API_URL || ""}/api/auth/google/login` ``), and all styling
- [x] T003 Update `frontend/src/pages/HomePage.tsx` to import `GoogleSignInButton` from `../components/GoogleSignInButton` — remove the inline `GoogleSignInButton` component definition and `GoogleIcon` SVG. Keep `GoogleSignInButtonDark` inline (only used on this page). Verify all three button instances (header, hero, hero large) still render correctly

**Checkpoint**: GoogleSignInButton is reusable. HomePage renders identically to before.

---

## Phase 3: User Story 4 — Store Family on Successful Login (Priority: P1)

**Goal**: After any successful login (parent or child), persist the family slug in localStorage so the auto-redirect (US1) can use it later.

**Independent Test**: Log in as a child → open browser console → run `localStorage.getItem("family_slug")` → should return the family slug. Log in as parent → same check. New parent with no family → should NOT have a value stored.

### Implementation for User Story 4

- [x] T004 [P] [US4] Store family slug on child login in `frontend/src/pages/FamilyLogin.tsx` — after successful `POST /auth/child/login` (after `setTokens()` call), call `setFamilySlug(familySlug)` where `familySlug` is the existing URL param. Import `setFamilySlug` from `../utils/familyPreference`
- [x] T005 [P] [US4] Store family slug on parent auth in `frontend/src/components/AuthenticatedLayout.tsx` — after `/auth/me` returns successfully, if `data.family_slug` is truthy, call `setFamilySlug(data.family_slug)`. Import `setFamilySlug` from `../utils/familyPreference`. This handles parent OAuth login (slug not known until /auth/me) and also serves as fallback for child login. Satisfies FR-008: new parents with `family_id === 0` have empty `family_slug`, so nothing is stored

**Checkpoint**: Family slug is stored in localStorage after both child and parent login flows.

---

## Phase 4: User Story 1 — Returning User Auto-Redirect (Priority: P1)

**Goal**: Visiting `/` automatically redirects to the stored family login page, with no visible flash of home page content.

**Independent Test**: Log in as any user (slug gets stored from Phase 3). Close tab. Open new tab, visit `/`. Should redirect to `/{family-slug}` immediately with no flash. Clear localStorage manually → visit `/` → should see normal home page.

### Implementation for User Story 1

- [x] T006 [US1] Add auto-redirect to `frontend/src/pages/HomePage.tsx` — at the top of the `HomePage` component function (before any hooks or JSX), call `getFamilySlug()`. If it returns a non-null value, return `<Navigate to={`/${slug}`} replace />` immediately. Import `getFamilySlug` from `../utils/familyPreference` and `Navigate` from `react-router-dom`. The synchronous localStorage read prevents any flash of home page content (FR-006)

**Checkpoint**: Full store-and-redirect loop works. Child logs in → slug stored → visit `/` → redirected to family login page.

---

## Phase 5: User Story 2 — Parent Login from Family Page (Priority: P1)

**Goal**: Parents can sign in with Google directly from the family login page, without navigating to the home page.

**Independent Test**: Visit `/{family-slug}` → see child picker grid → below it, see a divider, "Parent login" heading, and "Sign in with Google" button → click the button → complete Google OAuth → arrive at parent dashboard.

### Implementation for User Story 2

- [x] T007 [US2] Replace "Are you a parent?" section with Google sign-in button in `frontend/src/pages/FamilyLogin.tsx` — remove the existing "Are you a parent? Log in here" `<p>` + `<Link>` block (currently below the child picker grid). Replace with: (1) a horizontal divider styled with `border-t border-sand`, (2) a "Parent login" text label styled as `text-sm text-bark-light text-center`, (3) the `<GoogleSignInButton size="default" />` component centered below the label. Import `GoogleSignInButton` from `../components/GoogleSignInButton`. This section should only appear in the child picker view (not the password entry view), matching the existing placement of the "Are you a parent?" text (FR-003, FR-004)

**Checkpoint**: Parents can complete Google OAuth from the family login page.

---

## Phase 6: User Story 3 — "Not Your Bank?" Escape Hatch (Priority: P2)

**Goal**: Users on shared devices can clear the stored family preference and return to the home page.

**Independent Test**: Visit family login page → click "Not your bank?" → arrive at home page. Visit `/` again → stay on home page (no redirect).

### Implementation for User Story 3

- [x] T008 [US3] Add "Not your bank?" link to `frontend/src/pages/FamilyLogin.tsx` — below the new "Parent login" section (added in T007), add a text link: "Not your bank?" styled as `text-xs text-bark-light hover:underline cursor-pointer text-center`. On click: call `clearFamilySlug()` then `navigate("/")` (use `useNavigate` hook). Import `clearFamilySlug` from `../utils/familyPreference`. Must clear storage before navigating to prevent the HomePage redirect from sending the user back (FR-005, FR-007). The link should appear in the child picker view only (not the password entry view)

**Checkpoint**: Full escape hatch works — clear preference, land on home page, no re-redirect.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Verify build integrity across all changes

- [x] T009 Run TypeScript check (`cd frontend && npx tsc --noEmit`) and production build (`npm run build`) — fix any type errors or build failures introduced by the feature

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (imports familyPreference indirectly) — BLOCKS US2
- **US4 (Phase 3)**: Depends on Phase 1 (imports familyPreference)
- **US1 (Phase 4)**: Depends on Phase 1 (imports familyPreference) + Phase 3 (slug must be stored to test redirect)
- **US2 (Phase 5)**: Depends on Phase 2 (uses extracted GoogleSignInButton)
- **US3 (Phase 6)**: Depends on Phase 1 (imports familyPreference) + Phase 5 (adds below the parent login section from T007)
- **Polish (Phase 7)**: Depends on all phases complete

### User Story Dependencies

- **US4 (P1)**: Can start after Phase 1 — No dependencies on other stories
- **US1 (P1)**: Can start after Phase 1, but requires US4 to be functional for end-to-end testing
- **US2 (P1)**: Can start after Phase 2 — Independent of US1 and US4
- **US3 (P2)**: Can start after Phase 1, but should follow US2 (T007) since T008 adds content below T007's section

### Parallel Opportunities

- **T004 and T005** can run in parallel (different files: FamilyLogin.tsx vs AuthenticatedLayout.tsx)
- **Phase 3 (US4) and Phase 2 (Foundational)** can run in parallel after Phase 1 completes — US4 modifies FamilyLogin.tsx and AuthenticatedLayout.tsx while Phase 2 modifies HomePage.tsx and creates GoogleSignInButton.tsx

```
Phase 1 (T001)
    ├── Phase 2 (T002 → T003)  ──→ Phase 5: US2 (T007) ──→ Phase 6: US3 (T008)
    └── Phase 3: US4 (T004 ∥ T005) ──→ Phase 4: US1 (T006)
                                                                    ↓
                                                              Phase 7 (T009)
```

---

## Implementation Strategy

### MVP First (US4 + US1)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 3: US4 — store slug on login (T004, T005)
3. Complete Phase 4: US1 — auto-redirect from `/` (T006)
4. **STOP and VALIDATE**: Log in → close browser → visit `/` → confirm redirect works
5. This delivers the core user value with minimal changes

### Full Delivery

1. Setup (T001) → then Foundational (T002, T003) and US4 (T004, T005) in parallel
2. US1 (T006) after US4
3. US2 (T007) after Foundational
4. US3 (T008) after US2
5. Polish (T009) — verify everything builds

---

## Notes

- All changes are frontend-only — no backend modifications
- No existing frontend test framework — verification is manual + TypeScript type checking
- The `GoogleSignInButtonDark` variant stays inline in `HomePage.tsx` (only used there)
- Family slug is a public URL component — no security concern with localStorage storage
- Logout does NOT clear the family preference (per spec assumptions)
