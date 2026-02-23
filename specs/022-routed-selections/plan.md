# Implementation Plan: Routed Selections

**Branch**: `022-routed-selections` | **Date**: 2026-02-23 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/022-routed-selections/spec.md`

## Summary

Convert in-memory UI selection state (settings category tabs, child selector on dashboard/growth/children-settings) into URL path parameters. This enables page refresh persistence, browser back/forward navigation, and deep linking. Frontend-only changes to React Router routes and page components. No backend modifications.

## Technical Context

**Language/Version**: TypeScript 5.3.3, React 18.2.0
**Primary Dependencies**: react-router-dom 7.13.0 (already installed)
**Storage**: N/A (frontend-only, no DB changes)
**Testing**: `npx tsc --noEmit` + `npm run build` + manual browser testing
**Target Platform**: Web browser (desktop + mobile)
**Project Type**: Web application (frontend)
**Performance Goals**: N/A (no performance-sensitive changes)
**Constraints**: Must maintain existing React Router v7 patterns used in codebase
**Scale/Scope**: 5 files modified, 0 new files, 0 new dependencies

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | Frontend-only routing changes; verified via TypeScript compilation, build check, and manual browser testing. No financial logic or API changes. |
| II. Security-First Design | PASS | No new endpoints, no auth changes, no data exposure. Child names in URLs are already visible in the UI. |
| III. Simplicity | PASS | Uses existing react-router-dom patterns (useParams, useNavigate, Navigate). No new dependencies. Minimal changes to existing components. |

**Post-Phase 1 Re-check**: All principles still pass. No data model or API changes introduced.

## Project Structure

### Documentation (this feature)

```text
specs/022-routed-selections/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 research
├── data-model.md        # URL schema (no DB changes)
├── quickstart.md        # Implementation guide
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (files to modify)

```text
frontend/src/
├── App.tsx                          # Add parameterized routes
├── pages/
│   ├── ParentDashboard.tsx          # URL-driven child selection
│   ├── GrowthPage.tsx               # URL-driven child selection
│   └── SettingsPage.tsx             # URL-driven category + child selection
└── components/
    └── ChildrenSettings.tsx         # Controlled child selection via props
```

**Structure Decision**: No new files or directories. All changes are modifications to existing components within the established `frontend/src/` structure.

## Implementation Approach

### Phase 1: Route Definitions (App.tsx)

Replace static routes with parameterized routes:

**Current**:
```tsx
<Route path="/dashboard" element={<ParentDashboard />} />
<Route path="/growth" element={<GrowthPage />} />
<Route path="/settings" element={<SettingsPage />} />
```

**New**:
```tsx
<Route path="/dashboard/:childName?" element={<ParentDashboard />} />
<Route path="/growth/:childName?" element={<GrowthPage />} />
<Route path="/settings" element={<Navigate to="/settings/general" replace />} />
<Route path="/settings/:category" element={<SettingsPage />} />
<Route path="/settings/children/:childName" element={<SettingsPage />} />
```

The `?` suffix on `:childName?` makes the parameter optional (React Router v7 feature), so `/dashboard` and `/dashboard/bruce` both match. Settings uses a separate route for the nested child name since the category must be validated.

### Phase 2: Settings Page (SettingsPage.tsx)

- Replace `useSearchParams` with `useParams<{ category: string; childName?: string }>()`
- Derive `activeCategory` from `category` param (with validation)
- On category click: `navigate(\`/settings/${cat.key}\`)` instead of `setActiveCategory(cat.key)`
- Remove `useState` for `activeCategory` — URL is the source of truth
- Pass `selectedChildName` and `onChildSelect` callback to `ChildrenSettings`
- Redirect invalid categories to `/settings/general` with `replace: true`

### Phase 3: ChildrenSettings Component

- Accept new props: `selectedChildName?: string` and `onChildSelect: (child: Child | null) => void`
- Remove internal `selectedChild` state — becomes controlled
- Resolve `selectedChildName` → `Child` object after children list loads
- On child click in ChildSelectorBar: call `onChildSelect` (which triggers URL update in SettingsPage)
- On invalid `selectedChildName`: call `onChildSelect(null)` which triggers redirect

### Phase 4: Dashboard (ParentDashboard.tsx)

- Add `useParams<{ childName?: string }>()`
- Replace `setSelectedChild` state with URL-driven selection
- On child click: `navigate(\`/dashboard/${child.first_name.toLowerCase()}\`)` or `navigate("/dashboard")` for deselect
- On mount/param change: resolve `childName` → `Child` from loaded children list
- Invalid name → `navigate("/dashboard", { replace: true })`
- Update "Go to Settings → Children" link: `navigate("/settings/children")`

### Phase 5: Growth Page (GrowthPage.tsx)

- Same pattern as Dashboard: `useParams<{ childName?: string }>()`
- On child click: `navigate(\`/growth/${child.first_name.toLowerCase()}\`)` or `navigate("/growth")`
- Resolve `childName` → `Child` from children list
- Invalid name → `navigate("/growth", { replace: true })`
- Update "Go to Settings → Children" link: `navigate("/settings/children")`

### Key Design Decisions

1. **URL is source of truth**: Components derive selection from URL params, not internal state. This ensures refresh and back/forward always work correctly.

2. **Optional params for child routes**: `/dashboard/:childName?` means both `/dashboard` and `/dashboard/bruce` are valid, avoiding the need for separate routes per page.

3. **Separate routes for settings**: Settings needs `/settings/:category` and `/settings/children/:childName` as distinct routes because the category must be validated first. An index redirect handles bare `/settings`.

4. **Replace navigation for invalid URLs**: Using `navigate(..., { replace: true })` prevents invalid URLs from appearing in browser history.

5. **ChildrenSettings becomes controlled**: Selection state is lifted from ChildrenSettings to SettingsPage so the URL param can drive it. ChildSelectorBar already supports this pattern via its `onSelectChild` prop.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
