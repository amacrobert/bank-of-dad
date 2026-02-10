# Feature Specification: Account Management Enhancements

**Feature Branch**: `006-account-management-enhancements`
**Created**: 2026-02-09
**Status**: Draft
**Input**: User description: "Tweaks and features related to interest accrual, allowances, and transaction visibility: parent transaction history, interest rate form improvements, unified child management with single allowance, scheduled interest accrual, child interest visibility."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parent Views Child Transaction History (Priority: P1)

A parent wants to see the full transaction history for each of their children, including deposits, withdrawals, allowances, and interest earnings. When managing a child, the parent can view a chronological list of all transactions.

**Why this priority**: Transaction visibility is fundamental for parents to understand account activity and verify that automated features (allowances, interest) are working correctly. Without this, parents are operating blind.

**Independent Test**: Parent opens the manage child view, sees the child's transaction history with all transaction types displayed with dates, amounts, and labels.

**Acceptance Scenarios**:

1. **Given** a parent managing a child with existing transactions, **When** they view the child management screen, **Then** they see a chronological list of all transactions including deposits, withdrawals, allowances, and interest.
2. **Given** a parent managing a child with no transactions, **When** they view the child management screen, **Then** they see an empty state message like "No transactions yet."
3. **Given** a child with transactions of different types, **When** the parent views the transaction list, **Then** each transaction shows its date, amount (positive for deposits/allowances/interest, negative for withdrawals), type label, and any associated note.

---

### User Story 2 - Parent Configures Interest Rate with Pre-populated Form (Priority: P1)

When a parent manages a child, they see the child's current interest rate pre-populated in the interest rate form. They can update it and save. The interest rate is visible at a glance on the child management screen.

**Why this priority**: Parents need clear visibility into configured interest rates to avoid confusion or accidental changes. A pre-populated form prevents parents from accidentally overwriting an existing rate with a default value.

**Independent Test**: Parent opens the manage child view for a child with a 5% interest rate already set. The form displays "5.00" pre-populated. Parent changes it to 7% and saves successfully.

**Acceptance Scenarios**:

1. **Given** a child with a 5% interest rate already configured, **When** the parent opens the manage child view, **Then** the interest rate form shows "5.00" pre-populated.
2. **Given** a child with no interest rate configured (0%), **When** the parent opens the manage child view, **Then** the interest rate form shows "0.00" pre-populated.
3. **Given** a parent viewing the interest rate form, **When** they change the rate to 7.5% and save, **Then** the system confirms the update and the form now shows "7.50".

---

### User Story 3 - Unified Child Management with Single Allowance (Priority: P1)

Each child can have at most one allowance schedule. Instead of managing allowances in a separate section, the parent configures the child's allowance directly within the child management form. The form shows the current allowance configuration (amount, frequency, day) and allows the parent to create, edit, pause/resume, or remove it.

**Why this priority**: Simplifies the parent experience by consolidating child configuration into one place. The one-allowance-per-child constraint makes allowance management a natural part of child settings rather than a separate workflow.

**Independent Test**: Parent opens manage child, sees the allowance section showing "No allowance configured." Parent sets up a $10 weekly allowance on Fridays. Returning to the form, they see the allowance pre-populated and can edit or remove it.

**Acceptance Scenarios**:

1. **Given** a child with no allowance, **When** the parent opens the manage child view, **Then** they see an allowance section indicating no allowance is configured, with an option to set one up.
2. **Given** a child with no allowance, **When** the parent fills in amount ($10), frequency (weekly), and day (Friday) and saves, **Then** the allowance is created and the form shows the saved configuration.
3. **Given** a child with an existing weekly $10 allowance, **When** the parent opens the manage child view, **Then** the allowance form is pre-populated with $10, weekly, Friday.
4. **Given** a child with an active allowance, **When** the parent pauses it, **Then** the allowance shows as paused and no further deposits are made until resumed.
5. **Given** a child with an active allowance, **When** the parent removes it, **Then** the allowance schedule is deleted and the form returns to the "no allowance" state.
6. **Given** a child already has an allowance, **When** another allowance is attempted to be created for the same child, **Then** the system prevents it (only one allowance per child is allowed).

---

### User Story 4 - Scheduled Interest Accrual (Priority: P2)

Parents can configure when interest is accrued and paid out for each child, using a schedule with the same capabilities as allowance scheduling (weekly, biweekly, or monthly on a specific day). This replaces the current fixed monthly accrual with a parent-controlled schedule.

**Why this priority**: Gives parents flexibility to match interest payouts to their family's preferences (e.g., monthly on the 1st, or weekly). However, the default monthly behavior already works, so this enhances rather than enables the feature.

**Independent Test**: Parent sets interest accrual to "monthly on the 15th" for a child. On the 15th, the system calculates and credits interest. The parent can later change it to "weekly on Sundays."

**Acceptance Scenarios**:

1. **Given** a child with an interest rate set and no accrual schedule, **When** the parent views the manage child screen, **Then** they see an option to configure when interest is paid out (defaulting to monthly on the 1st).
2. **Given** a parent configuring interest accrual, **When** they select "weekly" and "Sunday," **Then** interest is calculated and credited every Sunday.
3. **Given** a parent configuring interest accrual, **When** they select "monthly" and the 15th, **Then** interest is calculated and credited on the 15th of each month.
4. **Given** a child with a biweekly interest schedule, **When** the scheduled day arrives, **Then** the system calculates interest based on the current balance and configured annual rate, prorated for the accrual period.
5. **Given** an interest accrual schedule exists, **When** the parent changes the interest rate to 0%, **Then** the accrual schedule remains but produces $0 interest (no transaction is created).

---

### User Story 5 - Child Sees Interest Rate and Next Payment (Priority: P3)

Children can see their account's interest rate and when their next interest payment will arrive, helping them understand how their savings grow.

**Why this priority**: Educational value for children, but not essential for the system to function. Parents and the scheduler operate independently of this visibility.

**Independent Test**: Child logs in, views their dashboard, and sees "Interest rate: 5.00% annually. Next interest payment: March 15."

**Acceptance Scenarios**:

1. **Given** a child with a 5% interest rate and a monthly accrual schedule on the 15th, **When** they view their dashboard, **Then** they see "5.00% annual interest" and "Next interest payment: [date]."
2. **Given** a child with no interest rate configured, **When** they view their dashboard, **Then** no interest information is displayed.
3. **Given** a child whose interest accrual just occurred today, **When** they view their dashboard, **Then** the next interest date reflects the next scheduled occurrence.

---

### Edge Cases

- What happens if the interest accrual schedule fires but the child's balance is $0? No interest transaction is created (interest on $0 is $0).
- What happens if the interest rate is changed between accrual periods? The new rate applies from the next accrual. Previously credited interest is not recalculated.
- What happens if a child has an interest accrual schedule but the rate is set to 0%? The schedule exists but produces no interest transactions.
- What happens if the existing standalone allowance section is still bookmarked or linked? The old standalone allowance routes should be removed, with allowance management only accessible through the child management view.
- What happens if a child currently has multiple allowances from the old system? Existing multiple allowances continue to function, but the UI only allows editing the first one or creating one if none exists. The system enforces the one-allowance constraint for new schedules going forward.
- What happens if a parent sets interest accrual to weekly but the annual rate is very low (e.g., 0.5%)? The system calculates the prorated amount per period. If it rounds to $0.00, no interest transaction is created.
- What happens to interest proration when switching from monthly to weekly schedule? Interest is calculated based on the current balance at the time of each accrual, using the annual rate prorated for the period length (1/52 for weekly, 1/26 for biweekly, 1/12 for monthly).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Parents MUST be able to view the full transaction history of any of their children from the child management screen
- **FR-002**: Transaction history MUST display date, amount, type (deposit, withdrawal, allowance, interest), and note for each transaction
- **FR-003**: The interest rate form MUST be pre-populated with the child's current interest rate when a parent opens the manage child view
- **FR-004**: The system MUST enforce a maximum of one allowance schedule per child
- **FR-005**: The child's allowance configuration MUST be part of the child management form, not a separate section
- **FR-006**: The allowance form within child management MUST support creating, editing, pausing, resuming, and removing the child's allowance
- **FR-007**: The allowance form MUST be pre-populated with existing allowance configuration when one exists
- **FR-008**: Parents MUST be able to configure an interest accrual schedule for each child, choosing frequency (weekly, biweekly, monthly) and day
- **FR-009**: The interest accrual scheduler MUST use the configured schedule to determine when to calculate and credit interest
- **FR-010**: Interest calculations MUST prorate the annual rate based on the accrual period (1/52 for weekly, 1/26 for biweekly, 1/12 for monthly)
- **FR-011**: Children MUST be able to see their configured interest rate on their dashboard
- **FR-012**: Children MUST be able to see when their next scheduled interest payment will occur
- **FR-013**: The standalone allowance management section MUST be removed from the parent view, with allowance management consolidated into child management
- **FR-014**: Existing allowance schedules MUST continue to function after the migration to one-per-child

### Key Entities

- **AllowanceSchedule** (modified): Retains all existing fields. A uniqueness constraint is added so that each child can have at most one schedule.
- **InterestAccrualSchedule**: Represents when interest is calculated and credited for a child. Contains child reference, frequency (weekly/biweekly/monthly), day of week or day of month, and status. Follows the same structure as AllowanceSchedule.
- **Transaction** (unchanged): Continues to record all financial events including interest accrual, with transaction type distinguishing them.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Parents can view a child's full transaction history within the manage child view in under 2 seconds
- **SC-002**: 100% of existing interest rate configurations are pre-populated correctly when the manage child form loads
- **SC-003**: Parents can configure a child's allowance from the manage child view in under 1 minute
- **SC-004**: Interest accrues according to the parent-configured schedule within 1 hour of the scheduled time
- **SC-005**: Children can see their interest rate and next payment date on their dashboard
- **SC-006**: The one-allowance-per-child constraint is enforced with a clear error message if violated

## Scope & Boundaries

### In Scope

- Parent transaction history viewing within child management
- Interest rate form pre-population
- Consolidating allowance management into child management form
- One-allowance-per-child enforcement
- Scheduled interest accrual (weekly, biweekly, monthly) replacing the fixed monthly approach
- Child-facing interest rate and next payment visibility
- Removing the standalone allowance management section

### Out of Scope

- Bulk operations on multiple children at once
- Transaction filtering or search within the history view
- Export or download of transaction history
- Notifications about interest accrual or allowance deposits
- Child-initiated requests to change their allowance or interest settings
- Compound interest calculations (interest is always simple, calculated on current balance)
- Custom accrual frequencies beyond weekly, biweekly, monthly

## Assumptions

- The existing transaction listing and allowance scheduling infrastructure will be reused
- Interest proration uses simple division (annual rate / number of periods per year)
- Interest amounts that round to $0.00 are not recorded as transactions
- The "one allowance per child" constraint applies going forward; existing children with multiple allowances retain them but cannot create new ones
- Interest accrual schedules follow the same end-of-month handling as allowance schedules (e.g., 31st falls back to last day of shorter months)
- The standalone allowance section/routes will be removed from the parent view, not just hidden

## Dependencies

- **002-account-balances**: Transaction system for recording and displaying transactions
- **003-allowance-scheduling**: Existing allowance schedule system to be modified (one-per-child constraint, consolidated UI)
- **005-interest-accrual**: Interest rate storage and calculation logic to be extended with scheduling
