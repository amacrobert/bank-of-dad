# Research: 018-child-management-settings

## Decision 1: Component Extraction Strategy

**Decision**: Extract the "Account Settings" section from `ManageChild.tsx` (reset password, update name/avatar, delete) into a standalone `ChildAccountSettings.tsx` component. Reuse the existing `AddChildForm.tsx` as-is in both Settings and onboarding contexts.

**Rationale**: `ManageChild.tsx` currently mixes financial management (balance, deposit, withdraw, transactions, allowance, interest) with account administration (password reset, name/avatar edit, delete). The cleanest split is to extract the account settings into their own component. The `AddChildForm` is already a standalone component with a simple `onChildAdded` callback — it needs no changes.

**Alternatives considered**:
- Duplicating the UI code in Settings: Rejected — violates DRY, causes maintenance burden.
- Moving the entire `ManageChild` to Settings: Rejected — it contains financial management that must stay on the dashboard.

## Decision 2: Settings Page Architecture

**Decision**: Add a "Children" category to the existing `CATEGORIES` array in `SettingsPage.tsx`. When the "Children" category is active, render a child management panel that shows the children list and allows CRUD operations.

**Rationale**: The Settings page already has a category navigation pattern (sidebar on desktop, tabs on mobile) with an extensible `CATEGORIES` array. Adding "Children" as a second category is the natural extension point — zero architectural changes needed.

**Alternatives considered**:
- Separate `/settings/children` route: Rejected — the current pattern uses client-side category switching, not route-based navigation. Adding routes would break the pattern.
- Accordion/expandable sections: Rejected — inconsistent with the existing category navigation UX.

## Decision 3: Onboarding Flow Extension

**Decision**: Extend the `SetupPage.tsx` from a 2-step flow (slug → confirmation) to a 3-step flow (slug → add children → confirmation). The add-children step reuses `AddChildForm` and shows a growing list of created children. A "Skip" link allows proceeding without adding children.

**Rationale**: The setup page already has a step-based flow with progress dots. Inserting a step between family creation and the confirmation screen is straightforward. The step is purely UI-driven — once the family is created (step 1), the parent has a valid `family_id` and all child creation API calls work normally.

**Alternatives considered**:
- Separate onboarding page: Rejected — the setup page already handles multi-step onboarding. Adding complexity with a new route is unnecessary.
- Modal-based child creation: Rejected — modals are harder to use on mobile and would break the clean step flow.

## Decision 4: Dashboard Empty State

**Decision**: When a parent has no children, the dashboard shows a friendly empty state card with a link to Settings → Children. The "Add a Child" collapsible section and form are removed entirely from the dashboard.

**Rationale**: The spec requires the dashboard to focus on financial management only. The empty state must direct parents to the correct location (Settings → Children) to add their first child.

**Alternatives considered**:
- Keep "Add Child" on dashboard as a convenience: Rejected — contradicts the spec's core requirement to move all child management to Settings.
- Redirect to Settings when no children: Rejected — confusing UX, user doesn't know why they left the dashboard.

## Decision 5: No Backend Changes Needed

**Decision**: This feature requires zero backend changes. All existing API endpoints are sufficient.

**Rationale**: The child CRUD endpoints (`POST /children`, `PUT /children/{id}/name`, `PUT /children/{id}/password`, `DELETE /children/{id}`) already exist and work correctly. The frontend is simply moving where these API calls are triggered from (dashboard → settings/onboarding). The `GET /children` endpoint for listing children is also already available.

**Alternatives considered**: None — the backend API surface is complete for this feature.
