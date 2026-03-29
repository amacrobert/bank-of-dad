# Feature Specification: Withdrawal Requests

**Feature Branch**: `032-withdrawal-requests`
**Created**: 2026-03-28
**Status**: Draft
**Input**: User description: "Withdrawal requests"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Child Requests a Withdrawal (Priority: P1)

A child wants to spend some of their saved money. They open their account, tap a "Request Withdrawal" button, enter an amount and a reason (e.g., "New video game"), and submit the request. The child sees the pending request on their dashboard and knows their parent will review it.

**Why this priority**: This is the core interaction — without children being able to initiate withdrawal requests, the feature has no purpose. Today, withdrawals are entirely parent-initiated, meaning children have no agency over their own money.

**Independent Test**: Can be fully tested by logging in as a child, submitting a withdrawal request, and verifying the request appears in a pending state with the correct amount and reason.

**Acceptance Scenarios**:

1. **Given** a child with a balance of $50.00, **When** they request a withdrawal of $20.00 with reason "Birthday present for mom", **Then** a pending withdrawal request is created and visible to the child.
2. **Given** a child with a balance of $10.00, **When** they try to request a withdrawal of $25.00, **Then** the request is rejected with a message indicating insufficient funds.
3. **Given** a child with $0.00 balance, **When** they view the withdrawal request form, **Then** the submit button is disabled and they see a message that they have no funds available.
4. **Given** a child with an existing pending request, **When** they view their account, **Then** they can see the pending request and its status.

---

### User Story 2 - Parent Reviews and Approves/Denies a Request (Priority: P1)

A parent sees a visual indicator in the app that a child has submitted a withdrawal request. The parent reviews the request details — amount, reason, and child's current balance — and chooses to approve or deny it. If approved, the funds are withdrawn from the child's account. If denied, the parent can optionally provide a reason.

**Why this priority**: Equally critical to Story 1 — requests without a review mechanism are useless. The parent approval step is what makes this a teaching tool for financial responsibility.

**Independent Test**: Can be fully tested by having a pending request in the system, logging in as a parent, viewing the request, and approving or denying it — then verifying the child's balance and request status update accordingly.

**Acceptance Scenarios**:

1. **Given** a pending withdrawal request from a child, **When** the parent approves it, **Then** the requested amount is withdrawn from the child's balance and a transaction record is created.
2. **Given** a pending withdrawal request, **When** the parent denies it with reason "Save up a bit more first", **Then** the request is marked as denied, the child's balance is unchanged, and the denial reason is visible to the child.
3. **Given** a pending withdrawal request that would now exceed the child's available balance (due to other transactions since the request), **When** the parent tries to approve it, **Then** the parent sees a warning that the child no longer has sufficient funds and must deny or the child must submit a new request.
4. **Given** multiple pending requests from different children, **When** the parent views their dashboard, **Then** they see all pending requests grouped by child.

---

### User Story 3 - Child Cancels a Pending Request (Priority: P2)

A child changes their mind about a withdrawal request that hasn't been reviewed yet. They can cancel the pending request themselves.

**Why this priority**: Gives children control and reduces unnecessary parent review work. Lower priority because the core flow works without it.

**Independent Test**: Can be fully tested by creating a pending request, then cancelling it as the child, and verifying it no longer appears as pending.

**Acceptance Scenarios**:

1. **Given** a child with a pending withdrawal request, **When** they cancel it, **Then** the request status changes to cancelled and it no longer appears in the parent's pending queue.
2. **Given** a child with an approved or denied request, **When** they view it, **Then** there is no option to cancel it.

---

### User Story 4 - Request History (Priority: P2)

Both parents and children can view past withdrawal requests and their outcomes. This provides a record of financial conversations and decisions.

**Why this priority**: Important for the educational value of the app but not required for the core request/approve flow to function.

**Independent Test**: Can be fully tested by having a mix of approved, denied, and cancelled requests, then verifying both parent and child can see the history with correct statuses and details.

**Acceptance Scenarios**:

1. **Given** a child with several past requests (approved, denied, cancelled), **When** they view their request history, **Then** they see each request with its amount, reason, status, and date.
2. **Given** a parent, **When** they view request history, **Then** they see past requests across all their children with the outcome and any denial reason they provided.

---

### Edge Cases

- What happens when a child submits a request and their account is subsequently disabled by the parent? The request is automatically denied.
- What happens if a parent approves a request but the withdrawal would impact savings goal allocations? The existing goal-impact warning flow applies.
- What happens if a child submits multiple requests that collectively exceed their balance? Each request is evaluated independently against the available balance at submission time; the parent sees a warning if approving one would make another unfulfillable.
- What happens if the child's available balance (excluding goal allocations) is less than the requested amount? The request validates against available balance, not total balance.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Children MUST be able to submit a withdrawal request specifying an amount and a reason.
- **FR-002**: The system MUST validate that the requested amount does not exceed the child's available balance at the time of submission.
- **FR-003**: Parents MUST be able to view all pending withdrawal requests from their children.
- **FR-004**: Parents MUST be able to approve or deny each withdrawal request individually.
- **FR-005**: When a parent approves a request, the system MUST withdraw the funds from the child's account and create a transaction record.
- **FR-006**: When a parent denies a request, the system MUST allow the parent to provide an optional denial reason.
- **FR-007**: Children MUST be able to cancel their own pending requests.
- **FR-008**: The system MUST prevent approval of a request if the child's available balance has fallen below the requested amount since submission.
- **FR-009**: Parents MUST see a visual indicator when pending withdrawal requests exist.
- **FR-010**: Both parents and children MUST be able to view a history of past withdrawal requests and their outcomes.
- **FR-011**: The reason field on a withdrawal request MUST be required and limited to 500 characters.
- **FR-012**: The withdrawal amount MUST follow existing constraints (minimum $0.01, maximum $999,999.99).
- **FR-013**: If an approved withdrawal would impact savings goal allocations, the existing goal-impact confirmation flow MUST apply.
- **FR-014**: If a child's account is disabled while they have pending requests, those requests MUST be automatically denied.
- **FR-015**: The system MUST allow only one pending withdrawal request per child at a time. A child must wait for their current request to be resolved or cancel it before submitting a new one.
- **FR-016**: Approved withdrawal requests MUST appear as a distinct transaction type (e.g., "Withdrawal Request") in the transaction history, distinguishable from parent-initiated withdrawals and linked to the originating request.

### Key Entities

- **Withdrawal Request**: Represents a child's request to withdraw funds. Key attributes: requesting child, amount, reason, status (pending/approved/denied/cancelled), denial reason, timestamps for creation and resolution.
- **Transaction**: Existing entity — an approved withdrawal request results in a new withdrawal transaction linked back to the originating request.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Children can submit a withdrawal request in under 30 seconds.
- **SC-002**: Parents can review and act on a withdrawal request in under 15 seconds.
- **SC-003**: 100% of approved requests result in an accurate balance deduction with no discrepancies.
- **SC-004**: Pending requests are visible to parents immediately after submission.
- **SC-005**: Children see updated request status immediately after parent action.
- **SC-006**: Request history is accessible and shows complete records for both parent and child views.

## Assumptions

- Children already have accounts with balances in the system.
- The existing parent-initiated direct withdrawal capability remains available alongside the new request-based flow.
- Withdrawal requests validate against the child's **available** balance (total balance minus savings goal allocations), consistent with existing withdrawal behavior.
- A reason/note is required for withdrawal requests (unlike parent-initiated withdrawals where the note is optional) to encourage children to think about their spending.
- No push notifications or email alerts — the visual indicator in the app is sufficient for the parent to notice pending requests.
- Pending withdrawal requests do not expire — they remain pending indefinitely until the parent approves/denies or the child cancels.
- Only one pending request per child at a time is allowed, to keep things simple and encourage children to prioritize. A child must wait for their current request to be resolved (or cancel it) before submitting a new one.

## Clarifications

### Session 2026-03-28

- Q: Should children be allowed to have multiple pending requests simultaneously, or limited to one at a time? → A: One pending request per child at a time.
- Q: Should approved withdrawal requests be distinguishable from parent-initiated withdrawals in transaction history? → A: Yes, show as a distinct label (e.g., "Withdrawal Request") linked to the originating request.
- Q: Should pending withdrawal requests expire automatically after a period of inactivity? → A: No expiration — requests stay pending until the parent acts or the child cancels.
