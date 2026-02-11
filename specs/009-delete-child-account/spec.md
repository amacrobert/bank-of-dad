# Feature Specification: Delete Child Account

**Feature Branch**: `009-delete-child-account`
**Created**: 2026-02-11
**Status**: Draft
**Input**: User description: "As a parent, I want to be able to delete a child from my bank. Add a Delete account in the child management dashboard at the bottom of Account Settings. Include a safety measure where the parent must type the child's name to continue. Deleting a child's account will delete all data in the database for that child, including but not limited to auth events, transactions, allowance and interest schedules, and sessions."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Delete a child account (Priority: P1)

A parent decides to remove a child's account from their family bank. They navigate to the child's management view, expand Account Settings, and scroll to the bottom where they find a "Delete Account" section. The section warns that this action is permanent and all account data will be lost. To proceed, the parent must type the child's first name exactly as it appears. Once confirmed, the system permanently deletes the child's account and all associated data (transactions, schedules, sessions, and audit records). The parent is returned to the dashboard with the child no longer listed.

**Why this priority**: This is the core and only feature — the ability for a parent to permanently remove a child account and all related data.

**Independent Test**: Can be fully tested by creating a child account with transactions, schedules, and sessions, then deleting it and verifying all data is removed from the system.

**Acceptance Scenarios**:

1. **Given** a parent is viewing a child's Account Settings, **When** they scroll to the bottom, **Then** they see a "Delete Account" section with a warning and a name confirmation input.
2. **Given** the parent has typed the child's name correctly into the confirmation input, **When** they click the delete button, **Then** the child account and all associated data are permanently deleted.
3. **Given** the parent has typed the child's name incorrectly or left the field empty, **When** they attempt to click the delete button, **Then** the button remains disabled and deletion does not proceed.
4. **Given** the child account has been successfully deleted, **When** the parent views the dashboard, **Then** the child no longer appears in the child list.
5. **Given** a deleted child had an active session, **When** that session is used for any request, **Then** the request is rejected as unauthorized.

---

### Edge Cases

- What happens if the parent types the child's name with different capitalization? The comparison should be case-insensitive to reduce frustration.
- What happens if the child is currently logged in when their account is deleted? Their session is deleted, so any subsequent request will fail authentication — no special handling needed beyond session cleanup.
- What happens if deletion fails partway through? The operation should be atomic — either all data is deleted or none is. The parent sees an error message and can retry.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display a "Delete Account" section at the bottom of Account Settings in the child management view.
- **FR-002**: The delete section MUST display a prominent warning that the action is permanent and irreversible, and that all account data will be lost.
- **FR-003**: The parent MUST type the child's first name into a confirmation input field before the delete button becomes active.
- **FR-004**: The name confirmation MUST be case-insensitive (e.g., typing "alice" confirms deletion for "Alice").
- **FR-005**: The delete button MUST remain disabled until the typed name matches the child's first name.
- **FR-006**: Upon confirmed deletion, the system MUST permanently remove the child record and all associated data: transactions, allowance schedules, interest schedules, sessions, and authentication event records.
- **FR-007**: The deletion MUST be atomic — all data is removed in a single operation, or the operation fails entirely with no partial deletion.
- **FR-008**: After successful deletion, the parent MUST be returned to the dashboard with the child list refreshed and the deleted child no longer shown.
- **FR-009**: If the deleted child is the currently selected child in the management view, the selection MUST be cleared.
- **FR-010**: The system MUST log the deletion as an auditable event before removing the child's data.

### Key Entities

- **Child Account**: The primary record being deleted. Has relationships to transactions, allowance schedules, interest schedules, sessions, and auth events.
- **Associated Data**: All records linked to the child — financial transactions, recurring schedule configurations, active login sessions, and authentication audit logs.

## Assumptions

- Only a parent within the same family can delete a child account. The existing authorization middleware (family ownership check) already enforces this.
- There is no soft-delete or recovery mechanism. Deletion is permanent and immediate, as specified by the user.
- The confirmation input compares against the child's current first name as stored in the system.
- The delete section uses a visually distinct danger/destructive styling to communicate severity (red tones, warning icon).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A parent can delete a child account and all associated data within 30 seconds of deciding to do so.
- **SC-002**: 100% of associated data (transactions, schedules, sessions, auth events) is removed when a child account is deleted — no orphaned records remain.
- **SC-003**: The name-confirmation safety measure prevents accidental deletion — the delete action cannot be triggered without the parent deliberately typing the child's name.
- **SC-004**: The deletion operation completes atomically — on failure, no data is partially deleted.
