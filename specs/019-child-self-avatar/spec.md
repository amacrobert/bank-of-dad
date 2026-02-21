# Feature Specification: Child Self-Avatar Update

**Feature Branch**: `019-child-self-avatar`
**Created**: 2026-02-21
**Status**: Draft
**Input**: User description: "As a child, I want to be able to update my own avatar. Add a setting on the child's Settings → Appearance page at the top to update their avatar."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Child Updates Their Avatar (Priority: P1)

A child navigates to their Settings → Appearance page and sees an avatar picker at the top of the page (above the existing theme picker). The child selects an emoji avatar from the grid of available options. The system saves the selection and shows a confirmation message. The new avatar is immediately reflected in the UI wherever their avatar is displayed.

**Why this priority**: This is the core feature — giving children autonomy to personalize their identity within the app.

**Independent Test**: Can be fully tested by logging in as a child, navigating to Settings → Appearance, selecting an avatar, and verifying it persists after page reload.

**Acceptance Scenarios**:

1. **Given** a child is on the Settings → Appearance page, **When** they select an emoji from the avatar picker, **Then** the system saves the avatar and displays a success message.
2. **Given** a child has previously set an avatar, **When** they visit Settings → Appearance, **Then** the avatar picker shows their current avatar as selected.
3. **Given** a child has a selected avatar, **When** they click the currently selected avatar to deselect it, **Then** their avatar is cleared (set to none) and a success message is displayed.

---

### User Story 2 - Avatar Change Reflected Across the App (Priority: P1)

After a child updates their avatar, the change is visible everywhere their avatar appears (dashboard header, parent's child list, family login page, etc.) without requiring a full app refresh.

**Why this priority**: The avatar change must propagate to be meaningful — if the child updates it but it doesn't appear elsewhere, the feature feels broken.

**Independent Test**: Update avatar as child, then verify it appears in the dashboard header and on the parent's view of the child list.

**Acceptance Scenarios**:

1. **Given** a child updates their avatar on the Appearance page, **When** they navigate to the dashboard, **Then** the new avatar is displayed in the header.
2. **Given** a child has updated their avatar, **When** a parent views the child list, **Then** the parent sees the child's updated avatar.

---

### Edge Cases

- What happens when the child's session expires while saving the avatar? The system returns an authentication error and the child must log in again.
- What happens if a network error occurs during save? The system displays an error message and the avatar picker reverts to the previous selection.
- Can a child set the same avatar they already have? Yes — this is a no-op that still returns success.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST display an avatar picker section at the top of the child's Settings → Appearance page, above the existing theme picker.
- **FR-002**: The avatar picker MUST show the same set of emoji avatars available to parents (the existing 16-emoji grid).
- **FR-003**: The child's currently selected avatar MUST be visually highlighted when the page loads.
- **FR-004**: The system MUST allow a child to select a new avatar by clicking an emoji in the picker.
- **FR-005**: The system MUST allow a child to clear their avatar by clicking the currently selected emoji (deselect).
- **FR-006**: The system MUST save the avatar selection immediately on click and display a success or error message.
- **FR-007**: The system MUST provide a child-scoped endpoint that only accepts requests from authenticated child users.
- **FR-008**: The system MUST NOT allow parents or unauthenticated users to call the child avatar self-update endpoint.
- **FR-009**: After a successful avatar update, the child's updated avatar MUST be reflected in the app without a full page reload.
- **FR-010**: The existing AvatarPicker component MUST be reused (no new picker component).

### Key Entities

- **Child**: Existing entity. The `avatar` field (nullable text, storing an emoji string) is the field being updated. No schema changes needed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A child can update their avatar from the Settings → Appearance page in under 5 seconds (select + confirmation).
- **SC-002**: The avatar picker correctly displays the child's current avatar as selected on page load 100% of the time.
- **SC-003**: Avatar changes are immediately visible in the dashboard header after saving, without a full page reload.
- **SC-004**: Unauthorized users (parents, unauthenticated) are rejected from the child avatar endpoint with appropriate error responses.

## Assumptions

- The existing 16-emoji avatar set is sufficient — no new emoji options need to be added.
- The existing AvatarPicker component can be reused without modification.
- The save-on-click pattern (no separate "Save" button) matches the existing theme update UX on the same page.
- No database migration is needed — the `children.avatar` column already exists and is a free-text nullable field.
