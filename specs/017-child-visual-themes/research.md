# Research: Child Visual Themes

**Feature**: 017-child-visual-themes
**Date**: 2026-02-18

## Decision 1: Theme Persistence Strategy

**Decision**: Store theme as a nullable `TEXT` column on the `children` table, matching the existing `avatar` column pattern.

**Rationale**: The `avatar` column is already nullable `TEXT`, scanned via `sql.NullString`, and represented as `*string` in the Go `Child` struct. Theme is the same kind of per-child preference. NULL means "sapling" (default). This avoids a separate settings table and keeps reads efficient (theme comes back with every child query, including `/auth/me`).

**Alternatives considered**:
- Separate `child_settings` table: Overkill for a single column. Would require joins on every auth check. Rejected per Simplicity principle.
- Store in `families` table: Theme is per-child, not per-family. Children in the same family should have independent themes.
- Frontend-only (localStorage): Would not persist across devices or survive cache clears. Spec requires cross-session persistence.

## Decision 2: API Endpoint for Theme Update

**Decision**: `PUT /api/child/settings/theme` with `requireAuth` middleware. Handler validates the user is a child and extracts child ID from auth context.

**Rationale**: Children must be able to set their own theme (spec: "Theme selection does not require parent approval"). Existing child mutation routes (`PUT /api/children/{id}/name`) require parent auth. A child-self-service endpoint under `/api/child/settings/` is the cleanest path and establishes a namespace for future child settings.

**Alternatives considered**:
- Extend `PUT /api/children/{id}/name` to include theme: Mixes parent-authorized operations with child self-service. Rejected for clarity.
- `PUT /api/children/{id}/theme` with parent auth: Contradicts spec requirement that children set their own theme without parent approval.
- No new middleware (reuse `requireAuth` + handler check): Chosen. Adding a `requireChild` middleware is premature — only one child endpoint exists.

## Decision 3: Theme Application Mechanism (Frontend)

**Decision**: Override CSS custom properties (`--color-forest`, `--color-forest-light`, `--color-cream`, `--color-cream-dark`) on `document.documentElement.style` and set `background-image` on the body element. Use a `ThemeProvider` React context.

**Rationale**: All existing components use Tailwind classes that resolve to CSS custom properties defined in `@theme {}` (e.g., `text-forest` → `--color-forest`). Overriding these properties at the root element makes every component automatically theme-aware with zero changes to individual component files. The background SVG is applied to the body. A ThemeProvider context manages state and cleanup.

**Alternatives considered**:
- `data-theme` attribute with CSS selectors: Would require duplicating every color rule. More CSS, more maintenance.
- Per-component `className` props: Would require touching every child-facing component. Violates simplicity.
- Tailwind `@layer` variants: Tailwind v4 doesn't natively support runtime theme switching via layers.

## Decision 4: Theme Color Palettes

**Decision**: Three carefully chosen palettes that maintain readability and contrast while feeling distinct.

| Theme      | Accent (text/buttons) | Accent Light (hover) | Background       | Background Dark (hover) |
|------------|-----------------------|----------------------|------------------|-------------------------|
| Sapling    | `#2D5A3D`            | `#3A7A52`            | `#FDF6EC`        | `#F5EBD9`               |
| Piggy Bank | `#8B4560`            | `#A55A78`            | `#FDF0F4`        | `#F5E0E8`               |
| Rainbow    | `#5B4BA0`            | `#7366B8`            | `#F3F0FD`        | `#E8E3F5`               |

**Rationale**: Each palette maintains sufficient contrast for WCAG AA compliance on body text. The accent colors are deep and rich (forest green, rose mauve, indigo purple) while backgrounds are soft pastels. The light variants follow the same offset pattern as the existing Sapling theme.

## Decision 5: SVG Background Pattern Design

**Decision**: Inline SVG data URIs with subtle, low-opacity repeating patterns.

| Theme      | Pattern                        | Style                                      |
|------------|--------------------------------|--------------------------------------------|
| Sapling    | Small scattered leaves         | ~8% opacity, scattered randomly, muted green |
| Piggy Bank | Faint coin circles             | ~6% opacity, scattered, muted rose          |
| Rainbow    | Tiny stars                     | ~6% opacity, scattered, muted purple         |

**Rationale**: Data URIs avoid additional network requests and keep themes self-contained. Low opacity (6-8%) ensures patterns are ambient — visible but never competing with content. The patterns are simple shapes (leaves, circles, stars) that render crisply at any size.

## Decision 6: Theme Slug Format

**Decision**: Use lowercase slugs: `sapling`, `piggybank`, `rainbow`. Store these exact strings in the database.

**Rationale**: Simple, URL-safe, no spaces or special characters. The display labels ("Sapling", "Piggy Bank", "Rainbow") are frontend concerns only. NULL in the database maps to `sapling` as the default.

**Alternatives considered**:
- Numeric IDs: Less readable in the database and API responses.
- Display names with spaces: Require encoding in URLs and JSON.
