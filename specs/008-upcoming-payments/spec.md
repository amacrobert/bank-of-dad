# Feature Specification: Unified Upcoming Payments Card

**Feature Branch**: `008-upcoming-payments`
**Created**: 2026-02-11
**Status**: Draft
**Input**: User description: "As a child, I want to see estimated upcoming interest payments combined with upcoming allowances in a single unified card on my dashboard, with an estimated interest amount based on current balance."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Child sees unified upcoming payments card (Priority: P1)

A child with both an active allowance schedule and active interest schedule opens their dashboard. Instead of seeing interest information crammed into the balance card and allowances in a separate card, they see a single "Upcoming allowance and interest" card that lists each upcoming payment — showing the amount, reason (allowance or interest), and when it will arrive.

**Why this priority**: This is the core feature — consolidating two separate displays into one clear, unified view that a child can easily understand.

**Independent Test**: Can be fully tested by logging in as a child with both allowance and interest configured, and verifying the unified card appears with both payment types listed.

**Acceptance Scenarios**:

1. **Given** a child has an active allowance schedule (e.g., $10/week) and an active interest schedule (e.g., 5% annual on a $100 balance), **When** the child views their dashboard, **Then** they see a single card titled "Upcoming allowance and interest" listing each upcoming payment with its amount, type label, and date.
2. **Given** a child has an active allowance and active interest, **When** the child views the card, **Then** the allowance entry shows the exact scheduled amount (e.g., "$10.00") and the interest entry shows an estimated amount (e.g., "~$0.10") calculated from the current balance.
3. **Given** a child has both upcoming payments, **When** the child views the card, **Then** each entry clearly indicates whether it is an allowance payment or an interest payment.

---

### User Story 2 - Card adapts to allowance-only or interest-only (Priority: P1)

A child who has only one type of scheduled payment (allowance or interest, not both) sees the card with an appropriate title and only the relevant entries.

**Why this priority**: Equal priority to User Story 1 — the card must adapt correctly to any configuration.

**Independent Test**: Can be tested by logging in as a child with only one payment type active and verifying the title and entries match.

**Acceptance Scenarios**:

1. **Given** a child has an active allowance schedule but no interest (rate is 0% or no schedule), **When** the child views their dashboard, **Then** the card is titled "Upcoming allowance" and shows only allowance payment(s).
2. **Given** a child has an active interest schedule (rate > 0%) but no allowance, **When** the child views their dashboard, **Then** the card is titled "Upcoming interest payment" and shows only the estimated interest amount and date.
3. **Given** a child has an active allowance but their interest schedule is paused, **When** the child views their dashboard, **Then** the card is titled "Upcoming allowance" and does not show interest.

---

### User Story 3 - Interest info removed from balance card (Priority: P1)

The interest rate badge and "Next interest" date line are removed from the "Your Balance" card. Interest information now lives exclusively in the upcoming payments card.

**Why this priority**: Essential cleanup — without this, interest info would be displayed in two places.

**Independent Test**: Can be tested by logging in as any child with interest configured and verifying the balance card shows only the balance amount.

**Acceptance Scenarios**:

1. **Given** a child has an active interest schedule, **When** the child views their dashboard, **Then** the balance card shows only the balance amount without an interest rate badge or next interest date.

---

### User Story 4 - No card when nothing is scheduled (Priority: P2)

A child with no active allowance and no active interest sees no upcoming payments card — the dashboard shows the balance card and transaction history without an empty placeholder.

**Why this priority**: Edge case, but important for a clean experience.

**Independent Test**: Can be tested by logging in as a child with no allowance or interest configured and verifying no upcoming payments card appears.

**Acceptance Scenarios**:

1. **Given** a child has no active allowance and no active interest, **When** the child views their dashboard, **Then** no upcoming payments card is displayed.

---

### Edge Cases

- What happens when the child's balance is $0.00 and interest is active? The estimated interest amount should display as "~$0.00".
- What happens when the interest estimate rounds to a very small fraction of a cent? Display should round to the nearest cent.
- What happens when the allowance has a note attached? The note should still be displayed alongside the allowance entry.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST display a unified "upcoming payments" card on the child dashboard that combines upcoming allowance and interest payments into a single list.
- **FR-002**: The card title MUST adapt based on which payment types are active for the child: "Upcoming allowance" (allowance only), "Upcoming interest payment" (interest only), or "Upcoming allowance and interest" (both).
- **FR-003**: Each entry in the card MUST display the payment amount, a label indicating the payment type (allowance or interest), and the date of the next payment.
- **FR-004**: The estimated interest payment amount MUST be calculated using the child's current balance and interest rate, without accounting for intermediate transactions before the interest date.
- **FR-005**: The interest estimate MUST be clearly marked as approximate (e.g., prefixed with "~" or labeled "Est.").
- **FR-006**: The interest estimate calculation MUST account for the payout frequency: weekly divides the annual rate by 52, biweekly by 26, and monthly by 12.
- **FR-007**: Interest rate and next interest date information MUST be removed from the "Your Balance" card on the child dashboard.
- **FR-008**: The card MUST NOT be displayed when the child has no active allowance and no active interest.
- **FR-009**: Allowance entries MUST continue to display any associated note text.
- **FR-010**: The card MUST be designed for a child audience — using clear labels, readable date formats, and prominent amounts.

### Key Entities

- **Upcoming Payment**: A line item representing either an allowance or interest payment, with: an amount (exact for allowance, estimated for interest), a type label, a date, and an optional note.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A child can identify all upcoming payments (type, amount, and date) from a single card without navigating to another page.
- **SC-002**: The card title correctly reflects the combination of active payment types in all four scenarios (both, allowance only, interest only, neither/hidden).
- **SC-003**: The estimated interest amount displayed matches the formula: balance * rate_bps / periods_per_year / 10,000, rounded to the nearest cent.
- **SC-004**: The balance card no longer contains any interest-related display elements.
