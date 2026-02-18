# Implementation Plan: Child Visual Themes

**Branch**: `017-child-visual-themes` | **Date**: 2026-02-18 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/017-child-visual-themes/spec.md`

## Summary

Add a child settings page with visual theme selection. Children choose from three themes (Sapling, Piggy Bank, Rainbow), each defining an accent color, background color, and subtle SVG background pattern. Theme preference is persisted per child in the database and applied via CSS custom property overrides on all child-facing pages.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: `jackc/pgx/v5`, `testify` (backend); react-router-dom, lucide-react, Vite (frontend)
**Storage**: PostgreSQL 17 — add `theme TEXT` column to existing `children` table
**Testing**: `go test -p 1 ./...` (backend), manual verification (frontend — no existing test framework)
**Target Platform**: Web application (desktop + mobile responsive)
**Project Type**: Web (Go backend + React frontend)
**Performance Goals**: Theme application < 1 second, no additional network requests for theme assets
**Constraints**: SVG backgrounds as inline data URIs (no external assets), themes affect only child-facing pages
**Scale/Scope**: 3 themes, 1 new API endpoint, 1 migration, ~6 modified files, ~3 new files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development

- **Backend**: TDD for store `UpdateTheme` method, handler validation, and `/auth/me` theme inclusion
- **Contract test**: New `PUT /api/child/settings/theme` endpoint
- **Frontend**: No existing test framework; manual testing per project convention
- **Status**: PASS — test plan covers all backend code paths

### II. Security-First Design

- **Authentication**: Theme endpoint requires `requireAuth` middleware
- **Authorization**: Handler validates `user_type === "child"` and updates only the authenticated child's own record
- **Input validation**: Theme value validated against allowlist (`sapling`, `piggybank`, `rainbow`)
- **No sensitive data**: Theme is a cosmetic preference, not financial or personal data
- **Status**: PASS — standard auth patterns, input validation, no sensitive data exposure

### III. Simplicity

- **Minimal schema change**: Single nullable column on existing table (no new tables)
- **CSS custom property override**: Zero changes to existing components — theme is applied at the root
- **No new dependencies**: All frontend theme logic uses built-in browser APIs
- **SVG data URIs**: No asset pipeline changes, no external requests
- **Status**: PASS — minimal footprint, leverages existing patterns

### Post-Design Re-check

- All constitution gates remain PASS after Phase 1 design
- No complexity violations identified

## Project Structure

### Documentation (this feature)

```text
specs/017-child-visual-themes/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── api.md           # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── migrations/
│   ├── 004_child_theme.up.sql       # NEW — ALTER TABLE children ADD COLUMN theme
│   └── 004_child_theme.down.sql     # NEW — rollback
├── internal/
│   ├── store/
│   │   └── child.go                 # MODIFY — add Theme field, update SELECTs, add UpdateTheme
│   ├── auth/
│   │   └── handlers.go              # MODIFY — include theme in /auth/me child response
│   ├── family/
│   │   └── handlers.go              # MODIFY — add HandleUpdateTheme
│   └── store/
│       └── child_test.go            # MODIFY — test UpdateTheme, test Theme in GetByID
└── main.go                          # MODIFY — register PUT /api/child/settings/theme

frontend/
├── src/
│   ├── pages/
│   │   └── ChildSettingsPage.tsx    # NEW — child settings page with theme selection
│   ├── context/
│   │   └── ThemeContext.tsx          # NEW — theme state management + CSS property overrides
│   ├── themes.ts                    # NEW — theme definitions (colors, SVG patterns)
│   ├── App.tsx                      # MODIFY — add route + ThemeProvider
│   ├── components/
│   │   └── Layout.tsx               # MODIFY — add Settings nav for child users
│   ├── types.ts                     # MODIFY — add theme to ChildUser
│   └── api.ts                       # MODIFY — add updateChildTheme function
```

**Structure Decision**: Follows existing web application structure. No new packages or directories beyond `frontend/src/context/` (which already exists for `TimezoneContext`).

## Design Decisions

### Theme Application via CSS Custom Properties

When a child is logged in with a non-default theme, the `ThemeProvider` overrides these CSS custom properties on `document.documentElement.style`:

| Property              | Sapling (default) | Piggy Bank  | Rainbow     |
|-----------------------|-------------------|-------------|-------------|
| `--color-forest`      | `#2D5A3D`         | `#8B4560`   | `#5B4BA0`   |
| `--color-forest-light`| `#3A7A52`         | `#A55A78`   | `#7366B8`   |
| `--color-cream`       | `#FDF6EC`         | `#FDF0F4`   | `#F3F0FD`   |
| `--color-cream-dark`  | `#F5EBD9`         | `#F5E0E8`   | `#E8E3F5`   |

All existing Tailwind utility classes (`text-forest`, `bg-cream`, etc.) automatically reflect theme colors. No component changes required.

### Background SVG Patterns

Each theme includes a subtle repeating SVG pattern applied as a `background-image` on the body element. Patterns use very low opacity (6-8%) so they add ambiance without competing with content. SVGs are embedded as data URIs — no external requests needed.

| Theme      | Pattern       | Opacity |
|------------|---------------|---------|
| Sapling    | Scattered leaves | ~8%  |
| Piggy Bank | Faint coins   | ~6%    |
| Rainbow    | Tiny stars    | ~6%    |

### Theme Cleanup on Logout

The `ThemeProvider` removes CSS overrides and background image when the user logs out or when a parent user is detected, ensuring parent-facing pages are never themed.

### API Design

Single new endpoint: `PUT /api/child/settings/theme`. Uses `requireAuth` (not `requireParent`) because children set their own themes. Handler extracts child ID from JWT auth context and validates ownership. Theme value validated against an allowlist.

### Database Design

Single new column: `children.theme TEXT` (nullable). NULL and `"sapling"` are treated identically in application code. Follows the exact pattern of the existing `avatar` column.

## Complexity Tracking

> No constitution violations. No complexity justifications needed.
