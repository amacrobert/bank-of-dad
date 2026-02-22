# Feature Specification: Remember Family Login

**Feature Branch**: `020-remember-family-login`
**Created**: 2026-02-21
**Status**: Draft
**Input**: User description: "As a user of Bank of Dad (parent or child), I don't want to have to remember my family bank's login URL when I visit the site. Upon a successful login, the frontend should store the family login URL in local storage. Successive visits to / will redirect the user to the family bank login page based on the stored value. This will mean that parent login will have to be available directly from the family login page. Replace the 'Are you a parent? Log in here' text with a 'Sign in with Google' button, underneath a divider separating out a 'Parent login' section. Finally, there needs to be an option on the login page for users who do not belong to this bank. Make a 'Not your bank?' link that erases the local storage value and sends the user to /."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Returning User Auto-Redirect (Priority: P1)

A returning user (parent or child) who has previously logged in visits the site's root URL. Instead of seeing the generic home page, they are automatically redirected to their family's login page. This eliminates the need to remember or bookmark the family-specific URL.

**Why this priority**: This is the core value of the feature — reducing friction for returning users who visit the site repeatedly.

**Independent Test**: Can be fully tested by logging in once, closing the browser, and revisiting `/`. The user should land on the family login page without any extra steps.

**Acceptance Scenarios**:

1. **Given** a user has previously logged in successfully and visits `/`, **When** the page loads, **Then** they are redirected to their family's login page (e.g., `/smith-family`).
2. **Given** a user has never logged in on this browser and visits `/`, **When** the page loads, **Then** they see the standard home page with no redirect.
3. **Given** a user has previously logged in and the stored family no longer exists, **When** they are redirected to the family login page, **Then** the family login page shows its existing "family not found" state and the user can navigate back to `/`.

---

### User Story 2 - Parent Login from Family Page (Priority: P1)

A parent visiting the family login page can sign in with Google directly from that page, without needing to navigate back to the home page. The current "Are you a parent? Log in here" text is replaced with a dedicated "Parent login" section containing a "Sign in with Google" button.

**Why this priority**: Since returning users will be auto-redirected to the family login page, parents must be able to log in from there. Without this, the auto-redirect feature would break the parent login flow.

**Independent Test**: Can be tested by visiting a family login page directly and completing a parent Google sign-in from that page.

**Acceptance Scenarios**:

1. **Given** a parent visits a family login page, **When** they look below the child account picker, **Then** they see a visual divider followed by a "Parent login" section with a "Sign in with Google" button.
2. **Given** a parent clicks the "Sign in with Google" button on the family login page, **When** they complete the Google OAuth flow, **Then** they are authenticated and redirected to the parent dashboard.

---

### User Story 3 - "Not Your Bank?" Escape Hatch (Priority: P2)

A user who is redirected to a family login page they don't belong to (e.g., using a shared device) can clear the remembered family and return to the home page.

**Why this priority**: Essential for shared devices and for correcting a stored family that no longer applies to the current user, but secondary to the core redirect and login functionality.

**Independent Test**: Can be tested by verifying the "Not your bank?" link is visible on the family login page, clicking it, and confirming the user arrives at the home page with no further auto-redirect.

**Acceptance Scenarios**:

1. **Given** a user is on a family login page, **When** they click the "Not your bank?" link, **Then** the stored family preference is cleared and they are navigated to the home page (`/`).
2. **Given** a user has clicked "Not your bank?" and cleared their preference, **When** they visit `/` again, **Then** they see the standard home page with no redirect.

---

### User Story 4 - Store Family on Successful Login (Priority: P1)

When any user (parent or child) successfully logs in, the system remembers which family login page to use for future visits. This happens automatically with no user action required.

**Why this priority**: This is the mechanism that powers the auto-redirect. Without storing the family preference, no redirect can occur.

**Independent Test**: Can be tested by logging in as a child or parent, then checking that the family preference has been persisted in the browser.

**Acceptance Scenarios**:

1. **Given** a child successfully logs in on a family login page, **When** login completes, **Then** the family login URL path is stored in the browser for future visits.
2. **Given** a parent successfully logs in via Google (from any page), **When** login completes and the parent's family is known, **Then** the family login URL path is stored in the browser for future visits.
3. **Given** a parent who has not yet set up a family logs in, **When** login completes, **Then** no family preference is stored (there is no family to remember yet).

---

### Edge Cases

- What happens when a user clears their browser's local storage? They see the standard home page on next visit — no redirect occurs. This is expected behavior.
- What happens when a user logs out? The stored family preference is **not** cleared on logout. The user should still be redirected to their family login page on their next visit, since that's the page they'd want to reach.
- What happens on a brand-new device or browser? No redirect occurs — the user sees the home page and must navigate to their family URL manually the first time.
- What happens if the stored family slug becomes invalid (family deleted or slug changed)? The redirect sends the user to the family login page, which already handles the "family not found" case with a message and a link back to the home page.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST store the family login URL path in the browser's local storage upon any successful login (parent or child).
- **FR-002**: System MUST redirect users from `/` to their stored family login page if a family preference exists in local storage.
- **FR-003**: The family login page MUST display a "Parent login" section below the child account picker, visually separated by a divider, containing a "Sign in with Google" button.
- **FR-004**: The "Sign in with Google" button on the family login page MUST initiate the same Google OAuth flow as the existing home page button.
- **FR-005**: The family login page MUST display a "Not your bank?" link that clears the stored family preference from local storage and navigates the user to `/`.
- **FR-006**: The redirect from `/` MUST happen before the home page content is displayed to avoid a flash of the home page.
- **FR-007**: The "Not your bank?" link MUST prevent the auto-redirect from immediately sending the user back (i.e., clearing storage must happen before navigation to `/`).
- **FR-008**: System MUST NOT store a family preference when a parent logs in but has not yet set up a family (new parent in onboarding flow).

### Key Entities

- **Family Preference**: A browser-local record of the user's family login URL path (e.g., `/smith-family`). Stored per-browser, not per-user. Has no server-side component.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Returning users who have previously logged in reach the family login page in one step (visiting `/`) instead of two (visiting `/` then navigating to family URL).
- **SC-002**: Parents can complete Google sign-in from the family login page without needing to visit any other page.
- **SC-003**: Users on shared devices can clear the remembered family and reach the home page in one click.
- **SC-004**: The redirect from `/` to the family login page completes without a visible flash of the home page content.

## Assumptions

- Only one family preference is stored per browser. If a user belongs to multiple families, the most recently logged-in family is remembered.
- The family preference persists indefinitely in local storage — there is no expiration.
- Logout does not clear the family preference, since the most common post-logout action is logging back in to the same family.
