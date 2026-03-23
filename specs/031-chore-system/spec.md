# Feature Specification: Chore & Task System

**Feature Branch**: `031-chore-system`
**Created**: 2026-03-22
**Status**: Draft
**Input**: User description: "Chore & Task System — Let parents create chores (e.g., 'Mow the lawn — $5', 'Clean your room — $2') that children can mark as complete. Parents approve completions, and pay is deposited automatically. Chores can be one-time or recurring."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parent Creates a Chore (Priority: P1)

A parent navigates to a child's management view and creates a new chore with a name, reward amount, and optional description. The parent can assign the chore to one or more of their children. The chore appears immediately in the assigned children's chore lists.

**Why this priority**: Without the ability to create chores, nothing else in the feature works. This is the foundational action that enables the entire chore workflow.

**Independent Test**: Can be fully tested by a parent creating a chore and verifying it appears in the child's chore list. Delivers value by establishing the chore catalog for the family.

**Acceptance Scenarios**:

1. **Given** a parent is logged in, **When** they create a chore with name "Mow the lawn", reward $5.00, and assign it to their child, **Then** the chore appears in that child's chore list with the correct name and reward amount.
2. **Given** a parent is creating a chore, **When** they assign it to multiple children, **Then** each assigned child sees the chore in their own chore list independently.
3. **Given** a parent is creating a chore, **When** they leave the name or reward amount blank, **Then** the system prevents submission and shows a validation message.
4. **Given** a parent is logged in, **When** they view their chore management area, **Then** they see all chores they have created with current status for each child.

---

### User Story 2 - Child Completes a Chore (Priority: P1)

A child views their list of assigned chores and marks one as complete. The chore moves to a "pending approval" state. The child can see which chores are available, which are pending approval, and which have been paid.

**Why this priority**: The child marking a chore as done is the core interaction loop. Without it, chores are just a static list with no engagement.

**Independent Test**: Can be tested by a child viewing their chore list, marking a chore complete, and verifying it moves to pending approval status.

**Acceptance Scenarios**:

1. **Given** a child has an assigned chore, **When** they mark it as complete, **Then** the chore status changes to "pending approval" and the parent is able to see it needs review.
2. **Given** a child has a chore in "pending approval" state, **When** they view their chore list, **Then** they can see it is awaiting parent approval and cannot mark it complete again.
3. **Given** a child has no assigned chores, **When** they view their chore list, **Then** they see an empty state message.

---

### User Story 3 - Parent Approves or Rejects Completion (Priority: P1)

A parent sees a list of chores pending approval across all their children. They can approve or reject each one. Approving a chore automatically deposits the reward amount into the child's account as a transaction. Rejecting returns the chore to an available state.

**Why this priority**: Approval is what triggers payment and closes the chore loop. It also gives parents control over quality — the child must actually do the work.

**Independent Test**: Can be tested by a parent approving a pending chore and verifying the reward is deposited into the child's balance, or rejecting and verifying the chore returns to available.

**Acceptance Scenarios**:

1. **Given** a child has marked a chore as complete, **When** the parent approves it, **Then** the chore reward amount is deposited into the child's account as a transaction with a note referencing the chore name.
2. **Given** a child has marked a chore as complete, **When** the parent rejects it, **Then** the chore returns to "available" status and no payment is made.
3. **Given** a parent rejects a chore, **When** they optionally provide a reason, **Then** the child can see the rejection reason on their chore list.
4. **Given** multiple children have pending chores, **When** the parent views their approval queue, **Then** they see all pending chores grouped or labeled by child.

---

### User Story 4 - Parent Creates a Recurring Chore (Priority: P2)

A parent creates a chore that repeats on a schedule (daily, weekly, or monthly). After the chore is approved (or the period elapses), a new instance of the chore automatically becomes available for the child.

**Why this priority**: Recurring chores eliminate the need for parents to manually re-create routine tasks like "Make your bed" every day. This significantly reduces ongoing effort and is where most real-world chore usage lives.

**Independent Test**: Can be tested by creating a recurring weekly chore, approving one instance, and verifying a new instance becomes available at the next scheduled time.

**Acceptance Scenarios**:

1. **Given** a parent creates a weekly recurring chore, **When** the child completes it and the parent approves it, **Then** a new instance of the chore becomes available at the start of the next period.
2. **Given** a recurring chore exists, **When** the child does not complete it within its period, **Then** the missed instance expires and a new instance becomes available for the next period.
3. **Given** a parent wants to stop a recurring chore, **When** they deactivate it, **Then** no new instances are created, but any pending instance can still be completed and approved.
4. **Given** a parent creates a daily recurring chore, **When** the child views their chore list, **Then** they see only the current day's instance (not future instances).

---

### User Story 5 - Child Views Chore Earnings History (Priority: P3)

A child can see a summary of how much they have earned from chores over time. This appears alongside their existing transaction history, with chore payments clearly labeled.

**Why this priority**: Seeing cumulative earnings from chores reinforces the connection between work and money. However, this is viewable through existing transaction history already — this story adds a chore-specific view.

**Independent Test**: Can be tested by a child with several approved chores viewing their earnings summary and verifying the total matches the sum of chore rewards.

**Acceptance Scenarios**:

1. **Given** a child has earned rewards from multiple chores, **When** they view their chore earnings, **Then** they see a total amount earned from chores and a list of completed chores with dates and amounts.
2. **Given** a child has no completed chores, **When** they view their chore earnings, **Then** they see an empty state with an encouraging message.

---

### User Story 6 - Parent Edits or Deletes a Chore (Priority: P3)

A parent can edit a chore's name, description, reward amount, or recurrence schedule. They can also delete a chore entirely. Editing does not affect already-completed or pending instances.

**Why this priority**: Management capabilities are important for long-term use but not needed for the initial workflow.

**Independent Test**: Can be tested by a parent editing a chore's reward amount and verifying new instances use the updated amount, while previously approved instances retain the original amount.

**Acceptance Scenarios**:

1. **Given** a parent edits a chore's reward amount, **When** the child next completes the chore, **Then** the new reward amount is used for payment.
2. **Given** a parent deletes a chore, **When** the child views their chore list, **Then** the chore no longer appears (pending instances are cancelled).
3. **Given** a chore has already been approved and paid, **When** the parent edits or deletes the chore, **Then** the historical transaction is not affected.

---

### Edge Cases

- What happens when a parent approves a chore but the child's account is disabled (free tier limit)? The approval should be blocked with a message explaining the child's account is inactive.
- What happens if a recurring chore's schedule crosses a timezone boundary (e.g., daily chore near midnight)? The system should use the family's configured timezone for determining period boundaries.
- What happens if a parent changes a chore's reward amount while an instance is pending approval? The pending instance should be paid at the reward amount that was set when the instance was created; the new amount applies to future instances only.
- What happens if a parent wants to change which children are assigned to a chore? The parent should delete the chore and recreate it with the desired assignments. Pending instances are cancelled on deletion; completed transaction history is preserved.
- What happens if the reward amount is $0.00? The system should allow $0 chores (useful for non-monetary tracking) but clearly indicate no payment will be made.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow parents to create chores with a name (required), reward amount in dollars and cents (required), and optional description.
- **FR-002**: System MUST allow parents to assign a chore to one or more of their children within the same family.
- **FR-003**: System MUST allow children to view their assigned chores organized by status: available, pending approval, and completed.
- **FR-004**: System MUST allow children to mark an available chore as complete, changing its status to pending approval.
- **FR-005**: System MUST allow parents to approve a pending chore, automatically depositing the reward amount into the child's account as a labeled transaction.
- **FR-006**: System MUST allow parents to reject a pending chore with an optional reason, returning it to available status.
- **FR-007**: System MUST support recurring chores with daily, weekly, or monthly frequency.
- **FR-008**: System MUST automatically generate new chore instances for recurring chores based on the configured schedule and family timezone.
- **FR-009**: System MUST allow parents to edit a chore's name, description, reward amount, and recurrence settings. Changes apply to future instances only.
- **FR-010**: System MUST allow parents to delete a chore, cancelling any pending instances while preserving completed transaction history.
- **FR-011**: System MUST allow parents to deactivate a recurring chore to stop generating new instances without deleting the chore.
- **FR-012**: System MUST expire missed recurring chore instances when the period elapses without completion.
- **FR-013**: System MUST prevent approval of chores for children whose accounts are disabled.
- **FR-014**: System MUST record the reward amount at the time the instance is created, so mid-edit changes do not alter pending payouts.
- **FR-015**: Chore reward transactions MUST appear in the child's existing transaction history with a note identifying the chore name.
- **FR-016**: System MUST allow reward amounts of $0.00 for non-monetary chore tracking.

### Key Entities

- **Chore**: A task defined by a parent with a name, optional description, reward amount, recurrence type (one-time, daily, weekly, monthly), active/inactive status, and assigned children. Belongs to a family.
- **Chore Instance**: A specific occurrence of a chore for a specific child. Has a status (available, pending approval, approved, rejected, expired) and tracks the reward amount at time of creation. For one-time chores, there is exactly one instance per assigned child. For recurring chores, new instances are generated per schedule period.
- **Chore Assignment**: The relationship between a chore and a child, indicating which children are responsible for a given chore.

## Assumptions

- Reward amounts follow the existing money convention (stored as integer cents internally, displayed as dollars).
- Chore reward deposits use the same transaction system as allowances and manual deposits — no new transaction infrastructure is needed.
- The recurring chore scheduler follows the same pattern as existing schedulers (allowance processor, interest accruer) — a background process that checks periodically.
- Family timezone (already configured via parent settings) is used for all schedule boundary calculations.
- There is no limit on the number of chores a family can create (no free-tier gating on chore count).
- Only parents can create, edit, delete, approve, or reject chores. Children can only view and mark as complete.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Parents can create a chore and have it appear in a child's list within 5 seconds.
- **SC-002**: The full chore lifecycle (create, assign, complete, approve, deposit) can be completed in under 2 minutes.
- **SC-003**: Recurring chore instances are generated within one scheduling cycle of their due time.
- **SC-004**: 100% of approved chore rewards result in a matching transaction in the child's account with no manual intervention after approval.
- **SC-005**: Children can distinguish between available, pending, and completed chores at a glance without reading instructions.
- **SC-006**: Parents can review and act on all pending chore approvals from a single view.
