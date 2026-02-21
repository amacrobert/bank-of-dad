# Implementation Plan: Child Management in Settings with Parent Onboarding

**Branch**: `018-child-management-settings` | **Date**: 2026-02-20 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/018-child-management-settings/spec.md`

## Summary

Move all child account administration (add, edit name/avatar, reset password, delete) from the parent dashboard to a new "Children" sub-page in Settings. Extend the parent onboarding flow to include adding children before first dashboard visit. Clean up the dashboard to show financial management only. **This is a frontend-only change — no backend or database modifications needed.**

## Technical Context

**Language/Version**: TypeScript 5.3.3 + React 18.2.0
**Primary Dependencies**: react-router-dom, lucide-react, Vite
**Storage**: N/A — no schema changes; existing PostgreSQL backend + REST API unchanged
**Testing**: Manual testing (no frontend test framework currently in place)
**Target Platform**: Web (desktop + mobile responsive)
**Project Type**: Web application (frontend only for this feature)
**Performance Goals**: Standard web app responsiveness
**Constraints**: Must work on mobile (bottom tab bar) and desktop (sidebar layout)
**Scale/Scope**: 7 files modified/created, ~5 frontend components affected

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development
- **Status**: PASS (with note)
- This is a frontend-only UI reorganization. The project has no frontend test framework. All existing backend API endpoints and their tests remain unchanged. Validation is via manual testing of each acceptance scenario.

### II. Security-First Design
- **Status**: PASS
- No new API endpoints or auth changes. All child CRUD operations continue to use the same authenticated endpoints. The onboarding add-children step only works after family creation (authenticated parent with valid JWT).

### III. Simplicity
- **Status**: PASS
- This feature *increases* simplicity by:
  - Removing clutter from the dashboard (single responsibility: financial management)
  - Consolidating child management into one location (Settings → Children)
  - Reusing existing components (`AddChildForm`, `AvatarPicker`, `ChildList`) rather than creating new abstractions
  - Extending the existing Settings category pattern rather than introducing a new architecture

### Post-Phase 1 Re-check
- All decisions confirmed simple. No new dependencies. No new abstractions beyond extracting existing code into a component. No backend changes.

## Project Structure

### Documentation (this feature)

```text
specs/018-child-management-settings/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 research decisions
├── data-model.md        # Data model (no changes needed)
├── quickstart.md        # Development quickstart
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
frontend/src/
├── components/
│   ├── AddChildForm.tsx          # EXISTING — reused as-is in Settings + onboarding
│   ├── AvatarPicker.tsx          # EXISTING — reused as-is
│   ├── ChildList.tsx             # MODIFY  — update empty state message
│   ├── ChildAccountSettings.tsx  # NEW     — extracted from ManageChild (password, name/avatar, delete)
│   ├── ChildrenSettings.tsx      # NEW     — Settings sub-page composing child list + add + account settings
│   ├── ManageChild.tsx           # MODIFY  — remove account settings section
│   ├── AuthenticatedLayout.tsx   # EXISTING — unchanged
│   └── Layout.tsx                # EXISTING — unchanged
├── pages/
│   ├── ParentDashboard.tsx       # MODIFY  — remove add-child, add empty state
│   ├── SettingsPage.tsx          # MODIFY  — add "Children" category
│   └── SetupPage.tsx             # MODIFY  — add step 2 (add children)
├── api.ts                        # EXISTING — unchanged
└── types.ts                      # EXISTING — unchanged
```

**Structure Decision**: Frontend-only changes within existing `frontend/src/` structure. Two new components created, five existing files modified. Backend untouched.

## Design Decisions

### 1. Component Extraction: `ChildAccountSettings`

Extract lines 224-316 of `ManageChild.tsx` (the "Account Settings" collapsible section) into a new `ChildAccountSettings.tsx` component. This component accepts a `Child` prop and renders:
- Reset Password form
- Update Name and Avatar form (with `AvatarPicker`)
- Delete Account form (with confirmation)

Each sub-form calls the existing API endpoints (`PUT /children/{id}/password`, `PUT /children/{id}/name`, `DELETE /children/{id}`).

### 2. Settings Children Sub-page: `ChildrenSettings`

A new component that composes:
- `ChildList` — for selecting a child to manage
- `AddChildForm` — for creating new children
- `ChildAccountSettings` — for managing the selected child

Layout: Similar to the dashboard's two-column approach (child list on left, management on right) but containing only account settings (no financial data).

### 3. Settings Page Integration

Add to `CATEGORIES` array in `SettingsPage.tsx`:
```typescript
{ key: "children", label: "Children", icon: Users }
```

When `activeCategory === "children"`, render `<ChildrenSettings />` instead of the general settings form.

### 4. Onboarding Extension

Extend `SetupPage.tsx` from 2 steps to 3 steps:
- **Step 1**: Choose family slug (existing — unchanged)
- **Step 2**: Add children (NEW — uses `AddChildForm`, shows list of created children, has "Skip" option)
- **Step 3**: Confirmation "You're all set!" (existing — unchanged)

The progress dots update from 2 to 3. Step 2 is skippable at any time.

### 5. Dashboard Cleanup

- Remove `AddChildForm` import and usage from `ParentDashboard.tsx`
- Remove "Add a Child" collapsible button
- Remove the `showAddChild` and `addChildInitialized` state
- When `ChildList` reports 0 children, show empty state card: "No children yet. Go to Settings → Children to add your first child." with a link to `/settings` (and auto-select "children" category via URL param or state).

### 6. ManageChild Cleanup

- Remove the "Account Settings" collapsible section (lines 224-316)
- Remove related state: `newPassword`, `newName`, `newAvatar`, `passwordMsg`, `nameMsg`, `deleteConfirmName`, `deleting`, `showSettings`, `settingsRef`
- Remove imports: `AvatarPicker`, `Trash2`, `ChevronDown`
- The component becomes purely financial: balance, deposit, withdraw, transactions, allowance, interest

## Complexity Tracking

> No constitution violations. No complexity justifications needed.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | — | — |
