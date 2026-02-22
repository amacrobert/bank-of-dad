# Implementation Plan: Remember Family Login

**Branch**: `020-remember-family-login` | **Date**: 2026-02-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/020-remember-family-login/spec.md`

## Summary

Store the family login URL slug in `localStorage` after successful login so returning users visiting `/` are automatically redirected to their family's login page. Add a "Sign in with Google" button directly to the family login page (replacing the "Are you a parent? Log in here" link) and provide a "Not your bank?" escape hatch to clear the stored preference.

This is a **frontend-only** feature — no backend changes, no database changes, no new API endpoints.

## Technical Context

**Language/Version**: TypeScript 5.3.3 + React 18.2.0
**Primary Dependencies**: react-router-dom, lucide-react, Vite (existing — no new deps)
**Storage**: Browser `localStorage` (no server-side persistence)
**Testing**: Manual browser testing (no existing frontend test framework)
**Target Platform**: Web browser (desktop + mobile)
**Project Type**: Web application (frontend only)
**Performance Goals**: Redirect from `/` must complete without visible flash of home page content
**Constraints**: Must work with existing Google OAuth flow (backend-initiated redirect to `/api/auth/google/login`)
**Scale/Scope**: 4 files modified, 1 new utility module

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | N/A | Frontend-only, no existing frontend test framework; feature is UI behavior verified manually |
| II. Security-First Design | PASS | No sensitive data stored (family slug is public URL component); no auth changes; OAuth flow unchanged |
| III. Simplicity | PASS | Minimal localStorage read/write; no new dependencies; leverages existing components |

**Quality Gates**:
- No new linting errors: Will verify with `tsc --noEmit`
- Build succeeds: Will verify with `npm run build`
- No API contract changes: Confirmed — no backend modifications

## Project Structure

### Documentation (this feature)

```text
specs/020-remember-family-login/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── quickstart.md        # Phase 1 output
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (repository root)

```text
frontend/src/
├── utils/
│   └── familyPreference.ts    # NEW — localStorage get/set/clear helpers
├── components/
│   └── GoogleSignInButton.tsx  # NEW — extracted from HomePage.tsx for reuse
├── pages/
│   ├── HomePage.tsx            # MODIFIED — check localStorage, redirect if set; import shared button
│   ├── FamilyLogin.tsx         # MODIFIED — store slug on login, add Google button, add "Not your bank?"
│   └── GoogleCallback.tsx      # MODIFIED — store slug after parent OAuth (from /auth/me on next load)
└── components/
    └── AuthenticatedLayout.tsx # MODIFIED — store slug when /auth/me returns a valid family_slug
```

**Structure Decision**: Frontend-only changes following existing project conventions. One new utility module for localStorage operations and one extracted component for the Google sign-in button (currently defined inside `HomePage.tsx`, now needed in two places).

## Design Decisions

### D1: Where to store the family slug

**Decision**: Store in `AuthenticatedLayout` for both user types, plus immediately in `FamilyLogin` for child login.

**Rationale**: `AuthenticatedLayout` calls `/auth/me` on every protected route load, and both `ParentUser` and `ChildUser` responses include `family_slug`. This gives us a single, reliable storage point. For child login, we also store the slug immediately in `FamilyLogin.tsx` since we know it from the URL param — this covers the case before the child reaches a protected route.

**Where each user type gets stored**:
- **Child**: Stored in `FamilyLogin.tsx` right after successful `POST /auth/child/login` (slug is the URL param)
- **Parent**: Stored in `AuthenticatedLayout.tsx` when `/auth/me` returns a user with non-empty `family_slug`
- **New parent (no family yet)**: Not stored — `family_slug` is empty until setup completes; `AuthenticatedLayout` redirects to `/setup` when `family_id === 0`

### D2: How to prevent flash of home page on redirect

**Decision**: Check `localStorage` synchronously at the top of the `HomePage` component and return a redirect (`<Navigate>`) immediately, before rendering any home page content.

**Rationale**: `localStorage` is synchronous — reading a value takes microseconds. By checking before the first render, no home page content ever appears. No loading spinner needed.

### D3: Extracting GoogleSignInButton

**Decision**: Extract the `GoogleSignInButton` component from `HomePage.tsx` into a standalone `components/GoogleSignInButton.tsx` file.

**Rationale**: The "Sign in with Google" button is currently defined inside `HomePage.tsx`. Now that it's needed on both the home page and the family login page, it should be a shared component. The existing Google SVG icon and styling move with it.

### D4: localStorage key and value

**Decision**: Key = `family_slug`, Value = the raw slug string (e.g., `smith-family`).

**Rationale**: Simple string value. The redirect constructs the URL as `/${slug}`. No JSON serialization needed. A dedicated utility (`familyPreference.ts`) provides `getFamilySlug()`, `setFamilySlug(slug)`, `clearFamilySlug()` to centralize access.

### D5: "Not your bank?" placement

**Decision**: Place the "Not your bank?" link at the bottom of the family login page, below the new "Parent login" section.

**Rationale**: It should be visible but not prominent — most users won't need it. Placing it at the very bottom keeps it accessible without cluttering the main login flow.

## File-by-File Change Summary

### New Files

**`frontend/src/utils/familyPreference.ts`**
- `getFamilySlug(): string | null` — reads from localStorage
- `setFamilySlug(slug: string): void` — writes to localStorage
- `clearFamilySlug(): void` — removes from localStorage
- Key constant: `FAMILY_SLUG_KEY = "family_slug"`

**`frontend/src/components/GoogleSignInButton.tsx`**
- Extract `GoogleSignInButton` component and Google SVG icon from `HomePage.tsx`
- Accept props: `size?: "default" | "compact" | "large"` (existing pattern)
- Construct Google login URL: `` `${import.meta.env.VITE_API_URL || ""}/api/auth/google/login` ``

### Modified Files

**`frontend/src/pages/HomePage.tsx`**
- Import `getFamilySlug` from `utils/familyPreference`
- Import `Navigate` from `react-router-dom`
- At top of component: if `getFamilySlug()` returns a slug, return `<Navigate to={`/${slug}`} replace />`
- Replace inline `GoogleSignInButton` / `GoogleSignInButtonDark` definitions with import from shared component
- Remove the now-extracted Google SVG and button component code

**`frontend/src/pages/FamilyLogin.tsx`**
- Import `setFamilySlug`, `clearFamilySlug` from `utils/familyPreference`
- Import `GoogleSignInButton` from `components/GoogleSignInButton`
- After successful child login (line ~55, after `setTokens()`): call `setFamilySlug(familySlug)`
- Replace "Are you a parent? Log in here" section (lines 179-184) with:
  - A horizontal divider
  - "Parent login" heading
  - `<GoogleSignInButton />` component
- Add "Not your bank?" link below the parent login section:
  - On click: `clearFamilySlug()` then `navigate("/")`
  - Styled as a subtle text link

**`frontend/src/components/AuthenticatedLayout.tsx`**
- Import `setFamilySlug` from `utils/familyPreference`
- After `/auth/me` returns successfully, if `data.family_slug` is truthy: call `setFamilySlug(data.family_slug)`
- This handles parent login (slug available after OAuth + /auth/me) and serves as a fallback for child login

**`frontend/src/pages/GoogleCallback.tsx`**
- No changes needed. The slug is stored by `AuthenticatedLayout` when the parent reaches `/dashboard`.

## Complexity Tracking

No constitution violations to justify — all three principles are satisfied.
