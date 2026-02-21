# Quickstart: 018-child-management-settings

## Prerequisites

- Node.js and npm installed
- Frontend dev server running (`cd frontend && npm run dev`)
- Backend running with PostgreSQL (`cd backend && go run .`)
- A parent account with at least one child (for testing dashboard cleanup)

## What This Feature Changes

This is a **frontend-only** feature. No backend or database changes.

### Files to Create
1. `frontend/src/components/ChildAccountSettings.tsx` — Account admin for a single child (password reset, name/avatar edit, delete)
2. `frontend/src/components/ChildrenSettings.tsx` — Settings sub-page composing child list + add form + account settings

### Files to Modify
1. `frontend/src/components/ManageChild.tsx` — Remove account settings section
2. `frontend/src/pages/ParentDashboard.tsx` — Remove add-child form, add empty state
3. `frontend/src/pages/SettingsPage.tsx` — Add "Children" category
4. `frontend/src/pages/SetupPage.tsx` — Add step 2 (add children) to onboarding
5. `frontend/src/components/ChildList.tsx` — Update empty state message

## Development Order

1. Extract `ChildAccountSettings` from `ManageChild` (can be tested immediately)
2. Build `ChildrenSettings` component for Settings page
3. Wire into `SettingsPage` as new category
4. Strip account admin from `ManageChild` and "Add Child" from dashboard
5. Update dashboard empty state
6. Extend `SetupPage` with add-children step
7. Update `ChildList` empty state text

## Manual Testing

1. **Settings → Children**: Navigate to Settings, click Children tab, verify you can add/edit/delete children
2. **Dashboard cleanup**: Verify dashboard only shows financial controls (no add/edit/delete)
3. **Dashboard empty state**: Delete all children, verify dashboard shows link to Settings → Children
4. **Onboarding**: Create a new account, verify step 2 prompts to add children, verify skip works
5. **Onboarding with children**: Add children during onboarding, verify they appear on dashboard
