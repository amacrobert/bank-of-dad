# Feature Specification: Child Management in Settings with Parent Onboarding

**Feature Branch**: `018-child-management-settings`
**Created**: 2026-02-20
**Status**: Draft
**Input**: User description: "Move child account management to settings. Add parent onboarding to include adding children."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parent Manages Children from Settings (Priority: P1)

As a parent, I want all child account management (add, edit, delete, reset password) consolidated under a "Children" sub-page in Settings, so my dashboard stays focused on financial activity rather than account administration.

**Why this priority**: This is the core ask — removing child management clutter from the dashboard. It affects every parent's daily experience and must work before onboarding can build on it.

**Independent Test**: Can be fully tested by navigating to Settings → Children and performing all CRUD operations on child accounts. Delivers a cleaner dashboard and organized settings experience.

**Acceptance Scenarios**:

1. **Given** a parent is on the Settings page, **When** they click the "Children" tab/link, **Then** they see a list of all their children with options to add, edit, and delete.
2. **Given** a parent is on the Children settings sub-page, **When** they click "Add a Child", **Then** they see the add-child form (name, password, avatar) and can create a new child account.
3. **Given** a parent is on the Children settings sub-page with existing children, **When** they select a child, **Then** they can edit the child's name and avatar, reset their password, or delete the account.
4. **Given** a parent has successfully added a child from Settings, **When** the child is created, **Then** the success screen displays the child's login credentials (same as current behavior).
5. **Given** a parent is on the parent dashboard, **When** they view the page, **Then** they no longer see the add-child form, name/avatar editing, password reset, or delete-account controls. Only the financial management UI remains (balance, deposit, withdraw, transactions, allowance, interest).

---

### User Story 2 - Parent Onboarding Includes Adding Children (Priority: P2)

As a new parent, I want the setup flow to guide me through adding my first children so that when I arrive at the dashboard for the first time, my children's accounts are already there and I can immediately start managing their finances.

**Why this priority**: This closes the gap in the new-user experience. Currently, parents land on an empty dashboard after setup and must figure out how to add children. This is the second priority because it depends on the child creation form being functional (which it already is, just needs to be reused in the onboarding context).

**Independent Test**: Can be fully tested by creating a new parent account, going through setup, adding at least one child during onboarding, and verifying the children appear on the dashboard upon arrival.

**Acceptance Scenarios**:

1. **Given** a parent has just created their family (chosen a slug), **When** they proceed to the next step, **Then** they see a screen prompting them to add their first child.
2. **Given** a parent is on the onboarding add-children step, **When** they add a child, **Then** the child appears in a list on the same screen and they can add more children.
3. **Given** a parent has added at least one child during onboarding, **When** they click "Continue to Dashboard" (or equivalent), **Then** the dashboard loads with those children already visible and selectable.
4. **Given** a parent is on the onboarding add-children step, **When** they choose to skip without adding children, **Then** they proceed to the dashboard (empty state) and can add children later from Settings → Children.
5. **Given** a parent has added children during onboarding, **When** a child is successfully created, **Then** the login credentials are displayed so the parent can share them with the child.

---

### User Story 3 - Dashboard Shows Only Financial Management (Priority: P3)

As a parent, I want the dashboard to focus exclusively on financial activity for my children — viewing balances, making deposits/withdrawals, managing allowances, and configuring interest — without any account administration UI.

**Why this priority**: This is the natural consequence of P1 and P2. Once management moves to Settings and onboarding handles initial setup, the dashboard cleanup follows. It's lower priority because it's largely a removal task.

**Independent Test**: Can be tested by logging in as a parent with existing children and verifying the dashboard only shows financial controls (balance, deposit, withdraw, transactions, allowance, interest) with no child CRUD operations visible.

**Acceptance Scenarios**:

1. **Given** a parent is on the dashboard with children, **When** they select a child, **Then** they see balance, deposit/withdraw, transactions, allowance schedule, and interest configuration — but no name/avatar editing, password reset, or delete controls.
2. **Given** a parent is on the dashboard, **When** they look for a way to add a child, **Then** no "Add a Child" button or form is present on the dashboard.
3. **Given** a parent needs to manage a child's account settings, **When** they look for account administration, **Then** there is a clear path from the dashboard to Settings → Children (e.g., a link or the existing Settings navigation).

---

### Edge Cases

- What happens when a parent deletes their last child from Settings? The dashboard shows an appropriate empty state directing them to Settings → Children to add a child.
- What happens if a parent navigates directly to the dashboard URL during onboarding (before completing the add-children step)? The onboarding flow should still be shown if setup is incomplete, but the add-children step is skippable so parents are never blocked from reaching the dashboard.
- What happens when a parent has many children (e.g., 10+)? The children list in Settings should be scrollable and manageable at any reasonable count.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a "Children" sub-page within the Settings page that lists all children in the family.
- **FR-002**: System MUST allow parents to add a new child from the Settings → Children page, using the same fields as today (first name, password, optional avatar).
- **FR-003**: System MUST display child login credentials after successful creation (from both Settings and onboarding).
- **FR-004**: System MUST allow parents to edit a child's name and avatar from the Settings → Children page.
- **FR-005**: System MUST allow parents to reset a child's password from the Settings → Children page.
- **FR-006**: System MUST allow parents to delete a child's account from the Settings → Children page, with the same confirmation mechanism (type child's name).
- **FR-007**: System MUST remove all child account management controls (add, edit name/avatar, reset password, delete) from the parent dashboard.
- **FR-008**: The parent dashboard MUST retain all financial management controls: balance display, deposit, withdraw, transaction history, allowance scheduling, and interest configuration.
- **FR-009**: The parent onboarding flow MUST include an add-children step after family creation and before arriving at the dashboard.
- **FR-010**: The onboarding add-children step MUST allow parents to add multiple children sequentially.
- **FR-011**: The onboarding add-children step MUST allow parents to skip without adding any children.
- **FR-012**: The Settings → Children page MUST use the same sidebar/tab navigation pattern as the existing Settings page (currently has "General" tab for timezone).
- **FR-013**: When a parent has no children, the dashboard MUST display an empty state that directs them to Settings → Children to add their first child.

### Key Entities

- **Child Account**: Represents a child's account within a family. Key attributes: first name, password, avatar (emoji), associated family. Management operations: create, update name/avatar, reset password, delete.
- **Family Setup State**: Tracks whether a parent has completed the full onboarding flow (family creation + optional child creation). Determines whether to show the onboarding flow or the dashboard.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Parents can perform all child management operations (add, edit, delete, reset password) entirely from the Settings → Children page without visiting the dashboard.
- **SC-002**: The parent dashboard contains zero child account administration controls — only financial management UI.
- **SC-003**: New parents who complete onboarding with at least one child see that child on the dashboard immediately upon first visit.
- **SC-004**: Parents can complete the full onboarding flow (family creation + adding first child) in under 3 minutes.
- **SC-005**: The onboarding add-children step is skippable — parents who skip arrive at the dashboard within 2 clicks of family creation.

## Assumptions

- The existing add-child form component (`AddChildForm`) and child management UI can be extracted and reused in the new Settings context and onboarding flow with minimal changes.
- No new backend API endpoints are needed — the existing child CRUD endpoints remain unchanged.
- The Settings page's existing category navigation pattern (sidebar on desktop, tabs on mobile) extends naturally to include a "Children" category alongside the existing "General" category.
- The onboarding flow completion state is determined by the existing `family_id` check — once a family exists, the parent proceeds to the dashboard. The add-children step is purely UI-driven and does not gate dashboard access.
- The "Family URL" sharing section currently on the dashboard remains on the dashboard (it is informational, not child management).
