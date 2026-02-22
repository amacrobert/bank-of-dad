# Research: Remember Family Login

## R1: localStorage for family slug storage

**Decision**: Use browser `localStorage` with a simple string value under key `family_slug`.

**Rationale**: The family slug is already a public URL component (anyone can visit `/{slug}`), so storing it in localStorage poses no security risk. `localStorage` persists across browser sessions, is synchronous (no flash on read), and requires zero dependencies. The value is a short string — no serialization overhead.

**Alternatives considered**:
- **`sessionStorage`**: Rejected — clears when the tab closes, which defeats the purpose of remembering across visits.
- **Cookie**: Rejected — unnecessary server-side visibility; adds complexity with domain/path scoping.
- **IndexedDB**: Rejected — overkill for a single string value; asynchronous reads would cause flash.

## R2: Redirect mechanism from `/`

**Decision**: Use React Router's `<Navigate>` component returned synchronously from `HomePage` before any content renders.

**Rationale**: `localStorage.getItem()` is synchronous, so the check completes before the first paint. Returning `<Navigate to={`/${slug}`} replace />` immediately prevents any flash of home page content. The `replace` flag ensures the home page doesn't appear in browser history.

**Alternatives considered**:
- **`useEffect` + `navigate()`**: Rejected — effect runs after first render, causing a visible flash of the home page.
- **Router loader/beforeEnter**: Rejected — react-router-dom v6 `loader` requires data router setup which the project doesn't use.
- **Middleware/wrapper component**: Rejected — unnecessary abstraction for a single-line check.

## R3: GoogleSignInButton extraction

**Decision**: Extract the existing `GoogleSignInButton` component from `HomePage.tsx` into `components/GoogleSignInButton.tsx`.

**Rationale**: The button (with Google SVG icon and styling) is currently defined inline in `HomePage.tsx`. Since it's now needed on the family login page too, it must be shared. The existing component API (`size` prop) is sufficient. The `GoogleSignInButtonDark` variant stays in `HomePage.tsx` since it's only used there (in the dark CTA section).

**Alternatives considered**:
- **Duplicate the button in FamilyLogin**: Rejected — violates DRY; styling/URL drift risk.
- **Extract both button variants**: Rejected — `GoogleSignInButtonDark` is only used on the home page; extracting it adds unnecessary complexity.

## R4: When to store the slug for parent login

**Decision**: Store in `AuthenticatedLayout` after `/auth/me` returns a valid `family_slug`.

**Rationale**: The parent OAuth flow is backend-initiated (`/api/auth/google/login` → Google → backend callback → redirect to `/auth/callback` with tokens). The frontend doesn't know the family slug until `/auth/me` is called. `AuthenticatedLayout` already calls this endpoint on every protected route load, making it the natural place to store the slug. For new parents without a family (`family_id === 0`), `family_slug` will be empty, so no preference is stored — satisfying FR-008.

**Alternatives considered**:
- **Store in GoogleCallback**: Rejected — the callback doesn't have the family slug; it only receives tokens.
- **Store in SetupPage after family creation**: Rejected — would only cover new parents; existing parents wouldn't benefit. `AuthenticatedLayout` covers both cases.

## R5: "Not your bank?" behavior

**Decision**: Clear localStorage and navigate to `/` using React Router's `navigate()`.

**Rationale**: The link must clear the stored slug *before* navigating to `/`, otherwise the `HomePage` component would immediately redirect back. Calling `clearFamilySlug()` then `navigate("/")` in the same click handler ensures correct ordering. Using `navigate()` (not `<Link>`) allows the clear operation to happen first.

**Alternatives considered**:
- **URL parameter to suppress redirect (e.g., `/?reset=1`)**: Rejected — adds URL complexity; leaks state into the URL.
- **React state flag**: Rejected — doesn't persist if the user refreshes; unnecessary complexity.
