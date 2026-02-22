# Quickstart: Remember Family Login

## Overview

Frontend-only feature. No backend changes, no database migrations, no new API endpoints.

## Prerequisites

- Node.js + npm (existing frontend dev environment)
- Running backend with at least one family set up (for manual testing)

## Files to Create

1. **`frontend/src/utils/familyPreference.ts`** — localStorage helpers (`getFamilySlug`, `setFamilySlug`, `clearFamilySlug`)
2. **`frontend/src/components/GoogleSignInButton.tsx`** — Extracted from `HomePage.tsx` for reuse on family login page

## Files to Modify

1. **`frontend/src/pages/HomePage.tsx`** — Add redirect logic; replace inline button with shared component
2. **`frontend/src/pages/FamilyLogin.tsx`** — Store slug on child login; add Google button + "Not your bank?" link
3. **`frontend/src/components/AuthenticatedLayout.tsx`** — Store slug after `/auth/me` returns valid family_slug

## Verification

```bash
cd frontend && npx tsc --noEmit && npm run build
```

### Manual test flows

1. **Child login stores slug**: Log in as child → check `localStorage.getItem("family_slug")` in browser console → should equal the family slug
2. **Auto-redirect works**: Close tab → open new tab → visit `/` → should redirect to family login page
3. **Parent Google login from family page**: Visit `/{slug}` → click "Sign in with Google" → complete OAuth → should reach parent dashboard
4. **"Not your bank?" clears preference**: On family login page → click "Not your bank?" → should reach home page → visiting `/` again stays on home page
5. **New parent (no family)**: Sign in with Google for first time → go through setup → after setup, `family_slug` should be stored
