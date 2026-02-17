# Feature Specification: Combine Transaction Cards

**Feature Branch**: `014-combine-transactions`
**Created**: 2026-02-16
**Status**: Draft
**Input**: User description: "The Upcoming allowance/interest cards and Recent transactions cards both display transactions -- just upcoming ones vs past ones. Combine them into one Transactions card with sub-headings Upcoming and Recent. Apply this change to both places these are displayed: in the parent dashboard with a child selected, and on the child dashboard. Once done, remove newly obsolete code/components."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parent Views Combined Transactions for a Child (Priority: P1)

A parent navigates to their dashboard and selects a child. Instead of seeing two separate cards (one for upcoming allowance/interest and one for recent transactions), they see a single "Transactions" card. This card has two clearly labeled sections: "Upcoming" and "Recent". The information displayed is identical to what was previously spread across two cards.

**Why this priority**: The parent dashboard is the primary management view. Consolidating transaction information here reduces visual clutter and provides a more cohesive view of a child's financial activity.

**Independent Test**: Can be fully tested by logging in as a parent, selecting a child, and verifying the combined Transactions card displays both upcoming and recent transactions with correct section headings.

**Acceptance Scenarios**:

1. **Given** a parent is logged in and has selected a child with both upcoming and recent transactions, **When** the parent views the dashboard, **Then** they see a single "Transactions" card with an "Upcoming" section listing future allowances/interest and a "Recent" section listing past transactions.
2. **Given** a parent has selected a child with upcoming transactions but no recent transactions, **When** the parent views the dashboard, **Then** the "Transactions" card shows the "Upcoming" section with entries and the "Recent" section with an appropriate empty state message.
3. **Given** a parent has selected a child with recent transactions but no upcoming transactions, **When** the parent views the dashboard, **Then** the "Transactions" card shows the "Upcoming" section with an appropriate empty state message and the "Recent" section with entries.
4. **Given** a parent has selected a child with no transactions of either type, **When** the parent views the dashboard, **Then** the "Transactions" card shows both sections with appropriate empty state messages.

---

### User Story 2 - Child Views Combined Transactions on Their Dashboard (Priority: P1)

A child logs into their dashboard and sees a single "Transactions" card replacing the previous two separate cards. The card contains the same "Upcoming" and "Recent" sub-headings with the same data that was previously split across the two cards.

**Why this priority**: Equal priority since the child dashboard is equally important and the change must be consistent across both views.

**Independent Test**: Can be fully tested by logging in as a child and verifying the combined Transactions card displays both upcoming and recent transactions with correct section headings.

**Acceptance Scenarios**:

1. **Given** a child is logged in and has both upcoming and recent transactions, **When** the child views their dashboard, **Then** they see a single "Transactions" card with "Upcoming" and "Recent" sections displaying the appropriate transactions.
2. **Given** a child has no transactions, **When** the child views their dashboard, **Then** the "Transactions" card shows appropriate empty state messages in both sections.

---

### User Story 3 - Obsolete Components Are Removed (Priority: P2)

After the combined card is in place, the previously separate "Upcoming allowance/interest" card and "Recent transactions" card are removed from the codebase. No dead or orphaned code remains.

**Why this priority**: Cleanup is secondary to delivering the user-facing change but is important for codebase maintainability.

**Independent Test**: Can be verified by confirming that the old separate card components no longer exist in the codebase and that no references to them remain.

**Acceptance Scenarios**:

1. **Given** the combined Transactions card is implemented on both dashboards, **When** a developer reviews the codebase, **Then** the old separate upcoming and recent transaction card components have been removed.
2. **Given** the old components have been removed, **When** a developer searches the codebase for references to the old components, **Then** no orphaned imports, references, or dead code related to them exists.

---

### Edge Cases

- What happens when a child has many upcoming or recent transactions? Existing display limits and scrolling behavior should be preserved.
- What happens when one section has data but the other is empty? Both sections should always be visible with appropriate empty state messaging.
- What happens when transaction data is loading? The card should show a loading state consistent with the existing pattern.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display a single "Transactions" card in place of the two separate transaction cards on both the parent dashboard (with child selected) and the child dashboard.
- **FR-002**: The "Transactions" card MUST contain an "Upcoming" sub-heading section displaying future scheduled allowances and interest entries (the same data previously shown in the "Upcoming allowance/interest" card).
- **FR-003**: The "Transactions" card MUST contain a "Recent" sub-heading section displaying past completed transactions (the same data previously shown in the "Recent transactions" card).
- **FR-004**: The "Upcoming" section MUST appear before the "Recent" section within the card.
- **FR-005**: Each section MUST show an appropriate empty state message when there are no transactions of that type.
- **FR-006**: All existing transaction data, formatting, and display behavior MUST be preserved in the combined card with no information loss.
- **FR-007**: The combined card MUST behave consistently on both the parent dashboard (child selected view) and the child dashboard.
- **FR-008**: All code related to the previously separate "Upcoming allowance/interest" and "Recent transactions" cards MUST be removed after the combined card is in place, with no orphaned or dead code remaining.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Both dashboards (parent with child selected, and child) display exactly one "Transactions" card instead of two separate transaction cards.
- **SC-002**: 100% of transaction data previously visible across the two cards remains visible in the combined card with no information loss.
- **SC-003**: Users can distinguish upcoming from recent transactions via clearly labeled "Upcoming" and "Recent" section headings within the card.
- **SC-004**: Zero orphaned or dead code related to the old separate transaction cards remains in the codebase after cleanup.
- **SC-005**: The combined card loads and displays transaction data within the same performance envelope as the previous two separate cards.

## Assumptions

- The "Upcoming" section heading replaces what was previously titled "Upcoming allowance/interest" — the shorter label is appropriate since context within a "Transactions" card makes the full description unnecessary.
- The "Recent" section heading replaces what was previously titled "Recent transactions" — similarly shortened.
- Existing sorting, display limits, and formatting for each transaction type are preserved as-is.
- No changes to underlying data or business logic are required — this is a purely presentational consolidation.
- Empty state messages will follow the existing empty state patterns used elsewhere in the application.
