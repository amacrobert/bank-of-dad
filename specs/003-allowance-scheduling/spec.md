# Feature Specification: Allowance Scheduling

**Feature Branch**: `003-allowance-scheduling`
**Created**: 2026-02-04
**Status**: Draft
**Input**: User description: "Automatic recurring deposits for children on a schedule (allowance scheduling)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parent Sets Up Weekly Allowance (Priority: P1)

A parent wants to automate their child's weekly allowance so they don't have to remember to make manual deposits each week. They configure a recurring deposit that automatically adds money to their child's account on a specific day each week.

**Why this priority**: This is the core value proposition of the feature. Without the ability to create scheduled allowances, the feature has no purpose. Weekly is the most common allowance frequency.

**Independent Test**: Parent creates a weekly allowance schedule for a child, system automatically deposits the specified amount on the scheduled day, and both parent and child can see the transaction in their history.

**Acceptance Scenarios**:

1. **Given** a parent with at least one child in their family, **When** the parent creates a weekly allowance of $10 for their child scheduled for Fridays, **Then** the system saves the schedule and shows it in the parent's allowance management view.

2. **Given** an active weekly allowance schedule, **When** the scheduled day arrives, **Then** the system automatically deposits the specified amount into the child's account with a note indicating it's a scheduled allowance.

3. **Given** an active allowance schedule, **When** the deposit is made, **Then** both the parent and child can see the transaction in the transaction history with clear indication it was an automatic allowance deposit.

---

### User Story 2 - Parent Views and Manages Allowance Schedules (Priority: P1)

A parent needs to see all active allowance schedules for their children and be able to modify or cancel them as family circumstances change.

**Why this priority**: Essential for parents to maintain control over automated deposits. Without management capabilities, parents cannot correct mistakes or adapt to changing needs.

**Independent Test**: Parent can view a list of all allowance schedules, edit the amount or frequency of an existing schedule, and pause or delete schedules they no longer need.

**Acceptance Scenarios**:

1. **Given** a parent with multiple allowance schedules, **When** they view the allowance management page, **Then** they see all schedules listed with child name, amount, frequency, and next scheduled date.

2. **Given** an active allowance schedule, **When** the parent edits the amount from $10 to $15, **Then** the new amount is used for all future deposits starting from the next scheduled occurrence.

3. **Given** an active allowance schedule, **When** the parent pauses the schedule, **Then** no deposits are made until the parent resumes it.

4. **Given** an active allowance schedule, **When** the parent deletes the schedule, **Then** no future deposits are made and the schedule is removed from their view.

---

### User Story 3 - Parent Creates Biweekly or Monthly Allowance (Priority: P2)

Some parents prefer to give allowances on a biweekly or monthly basis rather than weekly. The system should support these common scheduling frequencies.

**Why this priority**: Extends the core functionality to accommodate different family preferences. While weekly is most common, biweekly and monthly are also popular options.

**Independent Test**: Parent creates a monthly allowance schedule, and the system correctly deposits on the same day each month.

**Acceptance Scenarios**:

1. **Given** a parent creating a new allowance schedule, **When** they select "Monthly" frequency and choose the 1st of each month, **Then** the system schedules deposits for the 1st of each month.

2. **Given** a monthly schedule set for the 31st, **When** a month has fewer than 31 days, **Then** the deposit is made on the last day of that month.

3. **Given** a parent creating a biweekly schedule, **When** they set it to start on a specific date, **Then** the system deposits every 14 days from that start date.

---

### User Story 4 - Child Sees Upcoming Allowance (Priority: P3)

Children want to know when their next allowance is coming so they can plan their spending or saving.

**Why this priority**: Enhances the child experience but is not essential for the core functionality. Parents can operate the system without this feature.

**Independent Test**: Child logs in and sees when their next scheduled allowance will arrive and for how much.

**Acceptance Scenarios**:

1. **Given** a child with an active allowance schedule, **When** they view their dashboard, **Then** they see "Next allowance: $10 on Friday" (or similar messaging).

2. **Given** a child with no active allowance schedule, **When** they view their dashboard, **Then** no upcoming allowance information is displayed.

---

### Edge Cases

- What happens when a scheduled deposit would occur but there's a system outage? The system processes missed deposits when it recovers, ensuring no allowances are lost.
- What happens if a parent deletes a child who has an active allowance schedule? The schedule is automatically deleted along with the child.
- What happens if a schedule is set for February 29th in a non-leap year? The deposit is made on February 28th instead.
- What happens if multiple schedules for the same child trigger on the same day? Each schedule processes independently, resulting in multiple deposits.
- What happens if a schedule is created mid-week? The first deposit occurs on the next matching scheduled day.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow parents to create recurring deposit schedules for their children
- **FR-002**: System MUST support weekly, biweekly, and monthly frequencies for allowance schedules
- **FR-003**: System MUST automatically execute scheduled deposits at the appropriate time without manual intervention
- **FR-004**: System MUST record automated deposits as transactions with a clear indicator that they were scheduled allowances
- **FR-005**: Parents MUST be able to view all active allowance schedules for their family
- **FR-006**: Parents MUST be able to edit the amount of an existing allowance schedule
- **FR-007**: Parents MUST be able to pause and resume allowance schedules
- **FR-008**: Parents MUST be able to delete allowance schedules
- **FR-009**: System MUST handle end-of-month edge cases for monthly schedules (e.g., 31st falls back to last day of shorter months)
- **FR-010**: System MUST process any missed scheduled deposits when recovering from downtime
- **FR-011**: Children MUST be able to see when their next scheduled allowance will arrive
- **FR-012**: System MUST automatically delete allowance schedules when the associated child is deleted
- **FR-013**: Each allowance schedule MUST have an optional note/description that appears on each transaction

### Key Entities

- **AllowanceSchedule**: Represents a recurring deposit configuration. Contains the child it applies to, amount, frequency (weekly/biweekly/monthly), day of week or day of month, optional note, status (active/paused), and timestamps for creation and last modification.
- **ScheduledTransaction**: An extension of the existing Transaction entity that includes a reference to the originating AllowanceSchedule for audit purposes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Parents can create an allowance schedule in under 1 minute
- **SC-002**: Scheduled deposits execute within 1 hour of their scheduled time under normal operation
- **SC-003**: 100% of scheduled deposits are eventually processed (none are silently missed, even after system recovery)
- **SC-004**: Parents can view and modify all their allowance schedules from a single page
- **SC-005**: Children can see their upcoming allowance information on their dashboard

## Scope & Boundaries

### In Scope

- Creating, editing, pausing, resuming, and deleting allowance schedules
- Weekly, biweekly, and monthly frequencies
- Automatic deposit execution
- Parent management interface
- Child visibility of upcoming allowances
- Transaction history integration (scheduled deposits appear in history)

### Out of Scope

- Daily allowances (minimum frequency is weekly)
- One-time scheduled deposits (use manual deposit for these)
- Allowance schedules with end dates (schedules continue until manually stopped)
- Multiple allowance amounts that vary by schedule occurrence
- Notifications/reminders (email, push, SMS) about upcoming or completed deposits
- Child requesting changes to their allowance

## Assumptions

- The existing transaction system (from 002-account-balances) will be used to record scheduled deposits
- Parents can have multiple allowance schedules per child (e.g., weekly allowance + monthly bonus)
- Schedules use the server's timezone for determining when days start/end
- Paused schedules do not accumulate missed deposits - when resumed, they start fresh from the next scheduled occurrence
- The minimum allowance amount is $0.01 (1 cent), maximum is $999,999.99 (same as manual deposits)

## Dependencies

- **002-account-balances**: This feature depends on the balance and transaction system to record deposits
