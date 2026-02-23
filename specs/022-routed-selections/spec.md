# Feature Specification: Routed Selections

**Feature Branch**: `022-routed-selections`
**Created**: 2026-02-23
**Status**: Draft
**Input**: User description: "Make sub-selections on routed pages part of the routing, so that refreshes maintain category and child selections, and I can use the browser navigation buttons to navigate back and forward in relation to selections I've made."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Settings Category Navigation via URL (Priority: P1)

A parent navigates to the settings page and selects a category. The URL updates to reflect the selected category (e.g., `/settings/account`). When the parent refreshes the page, they remain on the same category. When they use the browser back button, they return to the previous category.

**Why this priority**: Settings category routing is the most broadly impactful change, affecting all three categories (general, children, account) and replacing the existing `?tab=` query parameter system.

**Independent Test**: Can be fully tested by navigating between settings categories and verifying URL updates, page refreshes, and browser back/forward button behavior.

**Acceptance Scenarios**:

1. **Given** a parent is on any page, **When** they navigate to `/settings`, **Then** they are redirected to `/settings/general`.
2. **Given** a parent is on `/settings/general`, **When** they click the "Children" category, **Then** the URL updates to `/settings/children` and the children content is displayed.
3. **Given** a parent is on `/settings/account`, **When** they refresh the page, **Then** they remain on `/settings/account` with the Account content displayed.
4. **Given** a parent navigated from `/settings/general` to `/settings/children` to `/settings/account`, **When** they press the browser back button, **Then** they return to `/settings/children`.
5. **Given** a parent presses the browser back button on `/settings/children` (after arriving from `/settings/general`), **When** back is pressed again, **Then** they return to `/settings/general`.
6. **Given** a parent navigates to `/settings/nonexistent`, **When** the page loads, **Then** they are redirected to `/settings/general`.
7. **Given** any existing link in the app uses `?tab=` query parameters for settings navigation, **When** the feature is deployed, **Then** those links use the new `/settings/{category}` URL path pattern instead.

---

### User Story 2 - Child Selection on Dashboard via URL (Priority: P1)

A parent views the family dashboard and selects a child. The URL updates to include the child's first name (e.g., `/dashboard/bruce`). Refreshing the page preserves the child selection. Browser back/forward navigates between child selections.

**Why this priority**: The dashboard is the primary parent-facing page and benefits most from persistent child selection state.

**Independent Test**: Can be fully tested by selecting children on the dashboard and verifying URL updates, page refreshes, and browser navigation.

**Acceptance Scenarios**:

1. **Given** a parent is on `/dashboard`, **When** they select a child named "Bruce", **Then** the URL updates to `/dashboard/bruce`.
2. **Given** a parent is on `/dashboard/bruce`, **When** they refresh the page, **Then** Bruce remains selected and the dashboard shows Bruce's data.
3. **Given** a parent is on `/dashboard/bruce`, **When** they select a different child named "Diana", **Then** the URL updates to `/dashboard/diana`.
4. **Given** a parent navigated from `/dashboard` to `/dashboard/bruce` to `/dashboard/diana`, **When** they press the browser back button, **Then** they return to `/dashboard/bruce` with Bruce selected.
5. **Given** a parent is on `/dashboard/bruce`, **When** they deselect Bruce (click again to toggle off), **Then** the URL updates to `/dashboard` with no child selected.
6. **Given** a parent navigates to `/dashboard/nonexistent`, **When** the page loads, **Then** they are redirected to `/dashboard` with no child selected.
7. **Given** a parent is on `/dashboard`, **When** the page loads with no child name in the URL, **Then** no child is pre-selected (same as current behavior).

---

### User Story 3 - Child Selection on Settings Children Page via URL (Priority: P2)

A parent navigates to the children settings category and selects a child to manage. The URL updates to include the child's first name (e.g., `/settings/children/bruce`). Refreshing preserves the selection.

**Why this priority**: This extends the settings category routing (Story 1) with an additional child name segment, creating a deeper nested route.

**Independent Test**: Can be fully tested by navigating to children settings, selecting a child, and verifying URL, refresh, and back/forward behavior.

**Acceptance Scenarios**:

1. **Given** a parent is on `/settings/children`, **When** they select a child named "Bruce", **Then** the URL updates to `/settings/children/bruce`.
2. **Given** a parent is on `/settings/children/bruce`, **When** they refresh the page, **Then** Bruce remains selected and the child management panel displays Bruce's settings.
3. **Given** a parent navigates to `/settings/children/nonexistent`, **When** the page loads, **Then** they are redirected to `/settings/children` with no child pre-selected.
4. **Given** a parent is on `/settings/children/bruce`, **When** they press the browser back button (having come from `/settings/children`), **Then** they return to `/settings/children` with no child selected.
5. **Given** a parent is on `/settings/children/bruce`, **When** they deselect Bruce, **Then** the URL updates to `/settings/children`.

---

### User Story 4 - Child Selection on Growth Page via URL (Priority: P2)

A parent views the growth projector page and selects a child. The URL updates to include the child's first name (e.g., `/growth/bruce`). Refreshing preserves the selection.

**Why this priority**: Consistent URL-driven child selection across all pages with child selectors, following the same pattern as the dashboard.

**Independent Test**: Can be fully tested by selecting children on the growth page and verifying URL, refresh, and back/forward behavior.

**Acceptance Scenarios**:

1. **Given** a parent is on `/growth`, **When** they select a child named "Bruce", **Then** the URL updates to `/growth/bruce`.
2. **Given** a parent is on `/growth/bruce`, **When** they refresh the page, **Then** Bruce remains selected and the growth projector displays Bruce's data.
3. **Given** a parent navigates to `/growth/nonexistent`, **When** the page loads, **Then** they are redirected to `/growth` with no child selected.
4. **Given** a parent is on `/growth/bruce`, **When** they press the browser back button (having come from `/growth`), **Then** they return to `/growth` with no child selected.

---

### Edge Cases

- What happens when two children have the same first name? The system matches the first child with that name (case-insensitive). Families are unlikely to have duplicate first names, but the system handles it gracefully.
- What happens when a child's name contains special characters or spaces? The URL uses the lowercase version of the child's first name. Names with spaces or special characters are URL-encoded as needed.
- What happens when a child is deleted while another user has their URL bookmarked? The name becomes invalid and the user is redirected to the parent route (no selection).
- What happens if a parent navigates directly to `/settings/children/bruce` without first loading the children list? The page fetches the children list, matches "bruce" against loaded children, and selects the match (or redirects if no match once loading completes).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The settings page MUST use URL path segments (`/settings/{category}`) instead of query parameters (`?tab=`) for category navigation.
- **FR-002**: Navigating to `/settings` without a category MUST redirect to `/settings/general`.
- **FR-003**: Navigating to `/settings/{invalid}` (where the category doesn't exist) MUST redirect to `/settings/general`.
- **FR-004**: Selecting a child on the dashboard MUST update the URL to `/dashboard/{firstName}` using the child's first name in lowercase.
- **FR-005**: Selecting a child on the settings children page MUST update the URL to `/settings/children/{firstName}` using the child's first name in lowercase.
- **FR-006**: Selecting a child on the growth page MUST update the URL to `/growth/{firstName}` using the child's first name in lowercase.
- **FR-007**: Deselecting a child on any page MUST update the URL to remove the child name segment (e.g., `/dashboard/bruce` → `/dashboard`).
- **FR-008**: Navigating to a URL with an invalid child first name MUST redirect to the parent route with no child selected.
- **FR-009**: Child name matching in URLs MUST be case-insensitive (e.g., `/dashboard/Bruce` and `/dashboard/bruce` both select the same child).
- **FR-010**: Page refreshes MUST preserve the current category and child selection based on the URL.
- **FR-011**: Browser back and forward buttons MUST navigate between previous selection states as recorded in browser history.
- **FR-012**: All existing in-app navigation links that use the `?tab=` query parameter MUST be updated to use the new URL path pattern.
- **FR-013**: The `?tab=` query parameter MUST be removed from the settings page entirely.

### Key Entities

- **Settings Category**: One of "general", "children", or "account" — represents a section of the parent settings page, identified by its URL slug.
- **Child Selection**: A child identified by their lowercase first name in the URL path, used across dashboard, growth, and settings children pages.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Refreshing any page with a category or child selection in the URL restores the exact same view 100% of the time.
- **SC-002**: Browser back and forward buttons correctly navigate between all selection states the user visited during their session.
- **SC-003**: All pages with child selectors (dashboard, growth, settings children) consistently reflect the selected child in the URL.
- **SC-004**: No `?tab=` query parameters remain in the application's navigation or URL handling.
- **SC-005**: Invalid URLs (bad category names, non-existent child names) gracefully redirect to sensible defaults within 1 navigation cycle (no redirect loops).

## Assumptions

- Child first names within a family are unique. If duplicates exist, the first match is used.
- The growth page (`/growth`) follows the same child selection URL pattern as the dashboard, since it also has a child selector. This is inferred from the general requirement that "on pages where a child is able to be selected, the child's name should be appended to the URL."
- URL child name segments use lowercase (e.g., `/dashboard/bruce` not `/dashboard/Bruce`) for consistency, but matching is case-insensitive.
- Child-only routes (`/child/*`) are not affected by this feature since child users only see their own data.
