# Feature Specification: Child Selector Redesign

**Feature Branch**: `021-child-selector-redesign`
**Created**: 2026-02-23
**Status**: Draft
**Input**: User description: "Redesign child selection pattern to eliminate wasted space from two-column layouts, support 0–12 children, and establish a reusable idiom for current and future parent-facing features."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parent selects a child on the Dashboard (Priority: P1)

A parent visits the Dashboard and sees a compact, horizontal child selector above the main content area. Each child is represented as a selectable chip showing their avatar and name. The parent taps a child's chip to select them, and the full-width content area below updates to show that child's financial management panel (balance, transactions, allowance, interest). The selected chip is visually highlighted. If no child is selected, the content area shows a prompt to select a child.

**Why this priority**: The Dashboard is the most-visited parent screen. Fixing its layout directly addresses the primary pain point of wasted vertical space when the right column is long but the left child list is short.

**Independent Test**: Can be fully tested by logging in as a parent with 1–3 children, selecting each child, and verifying the content area fills the full width with the selected child's financial data.

**Acceptance Scenarios**:

1. **Given** a parent with 3 children visits the Dashboard, **When** the page loads, **Then** a horizontal child selector bar appears above the content area showing all 3 children as chips with avatars and names, and no child is pre-selected.
2. **Given** the Dashboard is showing the child selector, **When** the parent clicks a child's chip, **Then** that chip becomes visually highlighted and the content area below shows the selected child's management panel at full width.
3. **Given** a child is selected on the Dashboard, **When** the parent clicks a different child's chip, **Then** the highlight moves to the new chip and the content area updates to show the newly selected child's data.
4. **Given** no child is selected, **When** the parent views the content area, **Then** a helpful empty state message is displayed prompting them to select a child.

---

### User Story 2 - Parent selects a child on Settings > Children (Priority: P2)

A parent navigates to Settings > Children and sees the "Add a new child" form at full width. Below it, the same horizontal child selector bar appears. When the parent selects a child, the account settings forms (reset password, update name/avatar, delete account) appear below the selector at full width, giving the forms ample space instead of being crammed into a narrow right column.

**Why this priority**: The Settings > Children page is the second page affected by the current layout problem, where child management forms are cramped in a narrow column. Fixing this provides immediate usability improvement for parents managing their children's accounts.

**Independent Test**: Can be fully tested by navigating to Settings > Children, selecting a child, and verifying the account settings forms render at full width with comfortable spacing.

**Acceptance Scenarios**:

1. **Given** a parent visits Settings > Children, **When** the page loads, **Then** the "Add a new child" form appears at full width, followed by a horizontal child selector bar showing all children.
2. **Given** the child selector is visible on Settings > Children, **When** the parent selects a child, **Then** the child's account settings (reset password, update name/avatar, delete account) appear below the selector at full width.
3. **Given** a child's settings are displayed, **When** the parent selects a different child, **Then** the settings update to reflect the newly selected child's data.

---

### User Story 3 - Child selector handles varying family sizes gracefully (Priority: P2)

The child selector pattern visually accommodates families of all sizes from 0 to 12 children without degrading the layout. For small families (1–4), chips are displayed inline. For larger families (5–12), the selector scrolls horizontally or wraps to a second row, ensuring all children remain accessible without overwhelming the page.

**Why this priority**: The selector must work well across all realistic family sizes. A pattern that only looks good for 2–3 children would fail to generalize and would need to be redesigned again.

**Independent Test**: Can be tested by creating families of varying sizes (1, 4, 8, 12 children) and verifying the selector remains usable and visually balanced at each size.

**Acceptance Scenarios**:

1. **Given** a parent has 1 child, **When** the selector renders, **Then** the single chip is displayed without awkward centering or excessive whitespace.
2. **Given** a parent has 6 children, **When** the selector renders, **Then** all 6 children are accessible via horizontal scrolling or wrapping, with a visual indicator that more children exist if scrolling is needed.
3. **Given** a parent has 12 children, **When** the selector renders, **Then** all 12 children remain accessible and the selector does not push the main content unreasonably far down the page.
4. **Given** a parent has 0 children, **When** the Dashboard loads, **Then** the child selector is not shown and a helpful empty state guides the parent to add children from Settings.

---

### User Story 4 - Reusable selector for future parent features (Priority: P3)

The child selector is implemented as a shared, reusable component that can be dropped into any parent-facing page. Future features (such as the Growth Projector viewed from a parent's perspective) can adopt the same selector without duplicating code or inventing a new selection idiom.

**Why this priority**: While not directly user-facing, establishing the selector as a reusable component ensures consistency across current and future pages and reduces future development effort.

**Independent Test**: Can be tested by verifying that both the Dashboard and Settings > Children pages use the same underlying selector component with identical interaction behavior.

**Acceptance Scenarios**:

1. **Given** the child selector component exists, **When** it is used on the Dashboard, **Then** it behaves identically to when it is used on Settings > Children (same visual style, same interaction pattern).
2. **Given** a developer wants to add child selection to a new page, **When** they use the shared component, **Then** it works without page-specific modifications beyond providing a callback for selection changes.

---

### Edge Cases

- What happens when a selected child is deleted (on Settings > Children)? The selector should deselect the removed child and collapse back to the unselected state.
- What happens when a new child is added while the selector is visible? The selector should refresh and include the new child without requiring a page reload.
- What happens on narrow/mobile screens? The chip selector should remain usable, potentially scrolling horizontally with touch-friendly sizing.
- What happens if a child's name is very long (e.g., 20+ characters)? The chip should truncate the name gracefully with an ellipsis.
- What happens when the parent's balance display for a child shows a very large number? The chip should only show the avatar and name (not the balance), keeping chips compact.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST replace the vertical child list sidebar on the Dashboard with a horizontal child selector bar positioned above the main content area.
- **FR-002**: System MUST replace the vertical child list sidebar on Settings > Children with a horizontal child selector bar positioned between the "Add a new child" form and the child account settings.
- **FR-003**: Each child in the selector MUST be represented as a compact chip showing their avatar (emoji or initial) and first name.
- **FR-004**: The selected child's chip MUST be visually distinct from unselected chips (e.g., highlighted background, border, or elevation change).
- **FR-005**: The child selector MUST support 0 to 12 children without layout degradation — using horizontal scrolling or row wrapping for larger families.
- **FR-006**: When a child is selected, the content area below the selector MUST use the full available width (no side-by-side columns for child list and detail).
- **FR-007**: The child selector MUST be a single, shared component reusable across any parent-facing page.
- **FR-008**: The selector MUST show a locked indicator for children whose accounts are locked.
- **FR-009**: The selector MUST update dynamically when children are added or removed without requiring a full page reload.
- **FR-010**: On mobile screens, the selector MUST remain usable with appropriately sized touch targets and horizontal scrolling when needed.
- **FR-011**: Child names exceeding the chip's display width MUST be truncated with an ellipsis.

### Key Entities

- **Child Chip**: A compact, selectable UI element representing one child — displays avatar and name. No new data entities are introduced; the selector operates on the existing Child data model.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The content area for a selected child's management panel or settings uses 100% of the available content width (no unused column beside it).
- **SC-002**: Parents can switch between children with a single click/tap on any page using the selector.
- **SC-003**: The selector renders correctly and remains usable for families with 1, 4, 8, and 12 children without requiring page scrolling just to see the selector itself.
- **SC-004**: Both the Dashboard and Settings > Children pages use the same shared selector component with identical visual treatment and interaction behavior.
- **SC-005**: On mobile viewports (< 768px), all child chips remain accessible via horizontal scroll with touch-friendly tap targets (minimum 44px height).

## Assumptions

- The child selector does not need to display the child's balance — balance information is shown in the detail content area after selection. This keeps chips compact.
- The current ChildList component and its data-fetching logic can be adapted into the new chip-based selector rather than built from scratch.
- The visual style of chips will follow the existing design language (cream/forest/sand color palette, rounded corners, subtle shadows).
- Pre-selecting the first child on page load is not required — the empty/unselected state with a prompt is acceptable and consistent with current behavior.
- The "Add a new child" form on Settings > Children remains in its current position above the child selector, not inside it.
