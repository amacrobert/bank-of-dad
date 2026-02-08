# Feature Specification: Interest Accrual

**Feature Branch**: `005-interest-accrual`
**Created**: 2026-02-07
**Status**: Draft
**Input**: User description: "Interest accrual on savings accounts"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parent Configures Interest Rate (Priority: P1)

As a parent, I want to set an interest rate on my child's account so that my child's savings grow over time and they learn about the value of saving money.

**Why this priority**: Without an interest rate configured, no interest can accrue. This is the foundational setup step that enables all other interest-related features.

**Independent Test**: Can be fully tested by a parent logging in, navigating to a child's account settings, setting an interest rate, and confirming the rate is saved and displayed.

**Acceptance Scenarios**:

1. **Given** a parent is viewing their child's account, **When** they set an interest rate (e.g., 5% per year), **Then** the rate is saved and displayed on the child's account.
2. **Given** a parent has set an interest rate, **When** they update the rate to a different value, **Then** the new rate takes effect and is displayed. The previous rate no longer applies to future accruals.
3. **Given** a parent has set an interest rate, **When** they set the rate to 0%, **Then** interest accrual stops for that child's account.
4. **Given** a parent has multiple children, **When** they configure interest rates, **Then** each child can have a different interest rate.

---

### User Story 2 - Automatic Interest Accrual (Priority: P1)

As a parent, I want interest to be automatically calculated and added to my child's account balance on a regular schedule so that savings grow without manual intervention.

**Why this priority**: This is the core value of the feature — automated interest growth. Without this, the interest rate is just a number with no effect.

**Independent Test**: Can be fully tested by setting an interest rate on a child's account, waiting for the accrual period to pass (or triggering it manually in a test), and verifying the balance increased by the correct amount.

**Acceptance Scenarios**:

1. **Given** a child has a balance of $100.00 and an annual interest rate of 10%, **When** the monthly accrual runs, **Then** $0.83 (10% / 12 months, rounded to the nearest cent) is added to the child's balance.
2. **Given** a child has a balance of $0.00, **When** the accrual runs, **Then** no interest is added and no transaction is recorded.
3. **Given** a child has no interest rate configured (or 0%), **When** the accrual runs, **Then** no interest is added to that child's account.
4. **Given** interest is accrued, **When** a parent or child views the account, **Then** the interest payment appears as a distinct transaction in the transaction history.

---

### User Story 3 - Child Sees Interest Earnings (Priority: P2)

As a child, I want to see how much interest I've earned so that I understand that saving money makes it grow, and I feel motivated to save more.

**Why this priority**: The educational value of interest is realized when children can see it happening. This story makes interest visible and meaningful to the child.

**Independent Test**: Can be fully tested by a child logging in and viewing their transaction history after interest has been accrued, confirming interest transactions are clearly labeled and the total interest earned is visible.

**Acceptance Scenarios**:

1. **Given** interest has been accrued on a child's account, **When** the child views their transaction history, **Then** interest payments are clearly labeled (e.g., "Interest earned") and distinguishable from deposits and withdrawals.
2. **Given** multiple interest payments have accrued over time, **When** the child views their account, **Then** they can see each individual interest payment with its date and amount.

---

### User Story 4 - Parent Views Interest Activity (Priority: P2)

As a parent, I want to see interest accrual activity across my children's accounts so that I can monitor the feature and understand how much interest has been paid out.

**Why this priority**: Parents need visibility into interest payments to maintain trust and oversight of the system they've configured.

**Independent Test**: Can be fully tested by a parent logging in and viewing transaction history for a child's account, filtering or identifying interest-related transactions.

**Acceptance Scenarios**:

1. **Given** interest has accrued on a child's account, **When** the parent views the child's transaction history, **Then** interest transactions are visible with date, amount, and the interest rate that was in effect.
2. **Given** a parent has multiple children with interest configured, **When** they view each child's account, **Then** they can see the interest history for each child independently.

---

### Edge Cases

- What happens when a child's balance is negative? Interest MUST NOT accrue on negative balances. Only positive balances earn interest.
- What happens when the calculated interest is less than one cent (e.g., $0.001)? The interest amount is rounded to the nearest cent. If the result rounds to $0.00, no interest transaction is created for that period.
- What happens when the interest rate is changed mid-month? The new rate applies starting from the next accrual period. The current period uses the rate that was in effect when the period began.
- What happens when a child's account is created mid-month? Interest accrues based on the balance at the time of the next scheduled accrual. No pro-rating for partial periods.
- What happens when the accrual process fails partway through (e.g., system error after processing some children)? Successfully accrued interest is preserved. Failed accounts are retried on the next accrual run. The same account MUST NOT receive duplicate interest for the same period.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Parents MUST be able to set an annual interest rate for each child's account, expressed as a percentage (e.g., 5%).
- **FR-002**: The interest rate MUST be configurable per child — different children in the same family can have different rates.
- **FR-003**: Parents MUST be able to update or disable (set to 0%) the interest rate at any time.
- **FR-004**: Interest MUST be calculated and added to child account balances automatically on a monthly schedule.
- **FR-005**: Interest MUST be calculated as: (current balance) × (annual rate / 12), rounded to the nearest cent.
- **FR-006**: Interest MUST NOT accrue on negative or zero balances.
- **FR-007**: Each interest payment MUST be recorded as a distinct transaction in the child's account history.
- **FR-008**: Interest transactions MUST be clearly distinguishable from deposits and withdrawals (labeled as "Interest earned" or similar).
- **FR-009**: Interest transactions MUST record the interest rate that was in effect when the interest was calculated.
- **FR-010**: The system MUST prevent duplicate interest accrual for the same child and the same accrual period.
- **FR-011**: The interest rate MUST be validated to be between 0% and 100% (inclusive).
- **FR-012**: Interest rate changes MUST take effect starting from the next accrual period, not retroactively.

### Key Entities

- **Interest Rate Configuration**: The annual interest rate assigned to a child's account by a parent. Attributes: child account, rate percentage, effective date.
- **Interest Transaction**: A record of interest credited to a child's account. Attributes: child account, amount, interest rate applied, accrual period, creation date. Related to: child account balance, transaction history.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Parents can configure an interest rate for a child's account in under 30 seconds.
- **SC-002**: Interest is automatically calculated and credited monthly with 100% accuracy (correct amount based on balance and rate).
- **SC-003**: Children can identify interest earnings in their transaction history within 5 seconds of viewing it.
- **SC-004**: Zero duplicate interest payments occur across all accounts over any time period.
- **SC-005**: Interest accrual completes for all accounts within 1 minute of the scheduled time, even with 100+ child accounts.

## Assumptions

- The existing account balance and transaction system (from feature 002) is in place and functional.
- The existing allowance scheduling system (from feature 003) provides a pattern for recurring background operations.
- Monthly accrual is the appropriate frequency for a family banking educational tool — it's frequent enough to be visible but not so frequent as to clutter transaction history.
- Simple interest (not compound) is used — interest is calculated on the current balance, not on accumulated interest. This is simpler to understand for children and simpler to implement.
- The interest rate is an annual percentage rate (APR) divided by 12 for monthly accrual — standard banking convention.
- No minimum balance is required to earn interest.
- Interest accrual runs as an automated background process, similar to how allowance scheduling already works in the system.
