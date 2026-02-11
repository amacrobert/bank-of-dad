# 007 Frontend Design - Implementation Plan

## Context

The Bank of Dad frontend is a functional React 18 + TypeScript SPA with 7 pages and 16+ components, but has minimal styling (42 lines of basic CSS). This plan redesigns it to a "Soft Modernist / Friendly Fintech" aesthetic per `frontend-design-abstract.md` — warm earthy palette, Nunito font, Lucide icons, Tailwind CSS, mobile-first responsive layout, and smooth animations. All existing functionality is preserved; this is a visual-only redesign.

---

## Phase 0: Infrastructure Setup

Install dependencies and configure tooling. **No visual changes yet.**

### Steps
1. **Install npm packages** (in `frontend/`):
   - `tailwindcss @tailwindcss/vite` (dev deps)
   - `lucide-react` (prod dep)

2. **Modify `frontend/vite.config.ts`** — add `@tailwindcss/vite` plugin alongside `react()`

3. **Modify `frontend/index.html`** — add Google Fonts `<link>` for Nunito (400–800 weights), add `<meta name="theme-color" content="#2D5A3D">`

4. **Rewrite `frontend/src/index.css`** — replace existing 42 lines with:
   - `@import "tailwindcss";`
   - `@theme` block defining custom colors (forest, sage, cream, sand, terracotta, amber, bark) and `--font-display: "Nunito"`
   - Base body styles (Nunito font, cream background, bark text)
   - Keyframe animations (fade-in-up, gentle-sway) for later use

5. **Rebuild dev container**: `docker compose build frontend && docker compose up -d frontend`

6. **Verify**: Page loads at localhost:8000 with cream background and Nunito font

### Color Palette (`@theme` values)
| Token | Hex | Usage |
|-------|-----|-------|
| `forest` | `#2D5A3D` | Primary actions, headings |
| `forest-light` | `#3A7A52` | Hover states |
| `sage` | `#87A98F` | Muted accents |
| `sage-light` | `#B5C9BA` | Light backgrounds, badges |
| `cream` | `#FDF6EC` | Page background |
| `cream-dark` | `#F5EBD9` | Card hover, input prefixes |
| `sand` | `#E8DCC8` | Borders, dividers |
| `terracotta` | `#C4704B` | Warnings, withdrawals |
| `amber` | `#D4A84B` | Highlights, gold accents |
| `amber-light` | `#F0D078` | Light amber backgrounds |
| `bark` | `#3D2E1F` | Primary text |
| `bark-light` | `#6B5744` | Secondary text |

### Files
- `frontend/package.json` (modified via npm)
- `frontend/vite.config.ts` (add plugin)
- `frontend/index.html` (add font + meta)
- `frontend/src/index.css` (rewrite)

---

## Phase 1: Shared UI Components

Create reusable building blocks in `frontend/src/components/ui/`.

### New Files
| File | Purpose |
|------|---------|
| `ui/Button.tsx` | Variants: primary (forest), secondary (cream-dark), danger, ghost. Min 48px touch target, loading spinner state, `active:scale-[0.97]` micro-interaction |
| `ui/Card.tsx` | White rounded-2xl container with soft shadow and sand border. Padding variants sm/md/lg |
| `ui/Input.tsx` | Label + input + error display. Rounded-xl, large padding, forest focus ring |
| `ui/Select.tsx` | Same pattern as Input for `<select>` elements |
| `ui/LoadingSpinner.tsx` | Centered Lucide `Loader2` with `animate-spin` + optional message. Page vs inline variants |

### Modified Files
| File | Changes |
|------|---------|
| `components/BalanceDisplay.tsx` | Restyle with Tailwind. Large: split dollars (text-5xl) and cents (text-2xl superscript). Color-code with forest green. |

---

## Phase 2: Layout Component

**Rewrite `frontend/src/components/Layout.tsx`**

Props: `user: AuthUser`, `children`, `maxWidth: "narrow" | "wide"` (480px vs 960px)

- **Mobile**: Fixed bottom tab bar (2-3 large labeled icons), content with bottom padding
- **Desktop (md:+)**: Top nav with leaf logo + "Bank of Dad" left, user name + logout right
- Child tabs: Home, Log out
- Parent tabs: Dashboard, Log out
- Uses Lucide icons (`LayoutDashboard`, `LogOut`, `Leaf`)

---

## Phase 3–7: Page Redesigns (simplest to most complex)

### Phase 3: NotFound (`pages/NotFound.tsx`)
- Full-screen centered on cream background
- Decorative Lucide tree icon with subtle CSS sway animation
- "Oops! This page wandered off." heading + "Go Home" primary button

### Phase 4: GoogleCallback (`pages/GoogleCallback.tsx`)
- Centered Lucide leaf icon with `animate-pulse`
- "Signing you in..." text. Error state with alert icon + back button
- Keep all existing redirect logic

### Phase 5: HomePage (`pages/HomePage.tsx`)
- Hero: Leaf icon + "Bank of Dad" heading + tagline + Google sign-in button (with inline Google "G" SVG)
- Decorative background blobs (sage/amber, blurred circles)
- 3 value prop cards (Shield, Coins, TrendingUp icons)
- Minimal footer

### Phase 6: FamilyLogin (`pages/FamilyLogin.tsx`)
- Card centered on cream background with Users icon avatar circle
- Large touch-friendly Input fields for name + password
- Styled error messages in red-tinted card
- Keep existing logic (family existence check, form submit, redirect)
- Note: No avatar grid — backend doesn't expose family member list to unauthenticated users

### Phase 7: SetupPage (`pages/SetupPage.tsx`) + SlugPicker restyle
- 2-step wizard: (1) Choose family URL, (2) "You're all set!" celebration
- Progress dots at top (forest green = done, sand = upcoming)
- SlugPicker: prefix label `bankofdad.com/` + input, CheckCircle/XCircle feedback, pill-shaped suggestions
- Celebration step: PartyPopper icon + "Go to Dashboard" button
- Keep existing slug validation logic

---

## Phase 8: ChildDashboard

**Rewrite `pages/ChildDashboard.tsx`** — wrap in Layout (narrow)

- Welcome greeting + child name
- Hero balance card (large BalanceDisplay, interest rate badge if > 0)
- Upcoming allowances section
- Recent activity with restyled TransactionHistory

**Restyle `components/TransactionHistory.tsx`**
- Each row: color-coded Lucide icon (ArrowDownCircle=deposit, ArrowUpCircle=withdrawal, Calendar=allowance, TrendingUp=interest) + description + amount
- Green text for income, terracotta for withdrawals
- Dividers with sand color

**Restyle `components/UpcomingAllowances.tsx`**
- Card with Calendar header, clean list items with date + amount

---

## Phase 9: ParentDashboard (most complex)

**Rewrite `pages/ParentDashboard.tsx`** — wrap in Layout (wide)

- Family info header with slug display
- **Desktop (md:+)**: Two-column grid — left: child list + add child, right: ManageChild detail panel (or empty state placeholder)
- **Mobile**: Stacked — child list at top, ManageChild expands below

### Components to restyle (8 files)
| Component | Key changes |
|-----------|-------------|
| `AddChildForm.tsx` | Card wrapper, Input/Button UI components, success state with credentials card |
| `ChildList.tsx` | Child cards with avatar circle (first letter), balance, chevron. Selected = forest ring. `selectedChildId` prop |
| `ManageChild.tsx` | Card-based sections instead of flat layout. Hero balance + deposit/withdraw buttons at top. Collapsible sections for settings |
| `DepositForm.tsx` | Input with $ prefix, Button UI, styled feedback |
| `WithdrawForm.tsx` | Same pattern as DepositForm |
| `ChildAllowanceForm.tsx` | Card with Calendar icon, Select/Input UI, status badge (active=sage, paused=sand), action button group |
| `InterestRateForm.tsx` | Card with TrendingUp icon, percentage input |
| `InterestScheduleForm.tsx` | Card, Select UI for frequency/day, status display |

---

## Phase 10: Polish & Animations

- Staggered `fade-in-up` animations on card lists (via `animationDelay` per item)
- Loading skeleton placeholders (pulse-animated sand rectangles) in ChildList, TransactionHistory, BalanceDisplay
- Responsive QA at 375px, 768px, 1024px+
- Accessibility: ARIA labels on nav, `aria-hidden` on decorative icons, verify contrast ratios (forest on cream = ~7.5:1, passes WCAG AA)
- Ensure bottom nav doesn't overlap content on mobile

---

## Files NOT Modified
- `frontend/src/api.ts` — API layer preserved
- `frontend/src/types.ts` — Types unchanged
- `frontend/src/App.tsx` — Routes unchanged
- `frontend/src/main.tsx` — Entry point unchanged
- `frontend/src/components/ProtectedRoute.tsx` — Unused, leave as-is
- `frontend/src/components/ScheduleForm.tsx` — Unused, leave as-is
- `frontend/src/components/ScheduleList.tsx` — Unused, leave as-is

---

## Key Architectural Decisions

1. **Tailwind v4 with `@theme`**: Tailwind v4 uses CSS-first config via `@theme` directive — no `tailwind.config.js` needed. Custom earthy colors defined directly in `index.css`, making `bg-forest`, `text-bark`, etc. available as native utility classes.

2. **Layout component with responsive nav**: Mobile bottom tab bar + desktop top nav in a single Layout component. Two max-width variants (480px child, 960px parent).

3. **FamilyLogin keeps form approach**: The spec's Netflix-style avatar grid requires a backend endpoint to list family members (doesn't exist). The form-based login is restyled to be warm and child-friendly instead.

4. **SetupPage is 2 steps, not 4**: Steps 2-3 from the spec (add children, set allowances) require backend changes. We implement: (1) choose slug, (2) celebration — matching current backend capabilities.

5. **ManageChild becomes card-based panels**: Converts from flat modal to sectioned Card layout that integrates into the parent dashboard's two-column grid on desktop.

---

## Verification

1. `docker compose build frontend && docker compose up -d`
2. Navigate to `http://localhost:8000`
3. Check each page: HomePage, SetupPage, FamilyLogin, ChildDashboard, ParentDashboard, NotFound
4. Test at mobile (375px) and desktop (1024px+) widths
5. Verify no console errors
6. Verify all existing functionality works (login, deposits, withdrawals, allowances, interest)
