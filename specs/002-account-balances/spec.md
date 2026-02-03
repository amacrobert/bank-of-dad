# Feature Specification: Account Balances

**Feature Branch**: `002-account-balances`
**Created**: 2026-02-03
**Status**: Draft
**Input**: User description: "Add account balances. Each child has an account balance. Account balances are read-only for children. Parents can add or remove money from their childrens' accounts. On the parents' page, they can see all account balances for accounts at their bank (aka their children)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parent Views All Children's Balances (Priority: P1)

A parent logs into their Bank of Dad account and navigates to their dashboard. They see an overview displaying the current account balance for each of their children, allowing them to quickly understand the financial state of their family "bank."

**Why this priority**: This is the core value proposition - parents need visibility into all accounts to manage family finances effectively. Without this view, the parent cannot fulfill their role as the "banker."

**Independent Test**: Can be fully tested by logging in as a parent and verifying all linked children's balances appear on the parent dashboard. Delivers immediate value by providing financial oversight.

**Acceptance Scenarios**:

1. **Given** a parent with 3 children registered, **When** the parent views their dashboard, **Then** they see all 3 children's names and current balances displayed
2. **Given** a parent with no children registered, **When** the parent views their dashboard, **Then** they see an empty state with guidance on how to add children
3. **Given** a parent viewing the dashboard, **When** a child's balance changes, **Then** the dashboard reflects the updated balance on next page load

---

### User Story 2 - Parent Adds Money to Child's Account (Priority: P1)

A parent wants to deposit money into their child's account (e.g., weekly allowance, reward for chores, birthday money from grandparents). They select the child, enter an amount and optional note, and confirm the deposit.

**Why this priority**: Depositing money is a fundamental operation - without it, children would never have any balance. This enables the core "bank" functionality.

**Independent Test**: Can be fully tested by a parent selecting a child, entering an amount, and confirming the deposit appears in the child's balance. Delivers immediate value by allowing money to flow into child accounts.

**Acceptance Scenarios**:

1. **Given** a parent on the dashboard, **When** they add $10.00 to a child's account with note "Weekly allowance", **Then** the child's balance increases by $10.00 and the transaction is recorded
2. **Given** a parent attempting to add money, **When** they enter an invalid amount (negative, zero, or non-numeric), **Then** the system displays an appropriate error message and does not process the transaction
3. **Given** a parent adding money, **When** they leave the note field empty, **Then** the transaction proceeds successfully without a note

---

### User Story 3 - Parent Removes Money from Child's Account (Priority: P2)

A parent needs to withdraw money from a child's account (e.g., child made a purchase that the parent paid for, or a correction is needed). They select the child, enter an amount to remove, add an optional note, and confirm the withdrawal.

**Why this priority**: While less frequent than deposits, withdrawals are essential for a complete banking experience. Parents need to deduct money when children spend or to correct errors.

**Independent Test**: Can be fully tested by a parent selecting a child with existing balance, entering a withdrawal amount, and confirming the balance decreases appropriately.

**Acceptance Scenarios**:

1. **Given** a child with a $50.00 balance, **When** the parent removes $15.00 with note "Bought a book", **Then** the child's balance becomes $35.00 and the transaction is recorded
2. **Given** a child with a $20.00 balance, **When** the parent attempts to remove $25.00, **Then** the system prevents the transaction and displays an error indicating insufficient funds
3. **Given** a parent removing money, **When** the amount would result in exactly $0.00 balance, **Then** the transaction proceeds successfully

---

### User Story 4 - Child Views Their Balance and Transaction History (Priority: P2)

A child logs into their Bank of Dad account and sees their current account balance displayed prominently along with a history of all transactions on their account. Each transaction shows the date, amount, type (deposit or withdrawal), and any note/description provided by the parent. The balance and history are read-only - children cannot modify anything.

**Why this priority**: Children need visibility into their own finances to understand their savings, track where money came from, and see where it went. This supports financial literacy goals by teaching children to review their transaction history.

**Independent Test**: Can be fully tested by logging in as a child and verifying their balance and transaction history are displayed, and no modification controls are available.

**Acceptance Scenarios**:

1. **Given** a child with a $75.50 balance and 5 transactions, **When** they view their dashboard, **Then** they see "$75.50" as their current balance and a list of all 5 transactions
2. **Given** a child viewing their transaction history, **When** they look at a transaction, **Then** they see the date, amount, type (deposit/withdrawal), and the note/description if one was provided
3. **Given** a child viewing their balance, **When** they attempt to find ways to modify the balance or transactions, **Then** no edit, add, or remove controls are available to them
4. **Given** a child with a $0.00 balance and no transactions, **When** they view their dashboard, **Then** they see "$0.00" displayed and an empty transaction history (not an error state)
5. **Given** a child with many transactions, **When** they view their history, **Then** transactions are displayed in reverse chronological order (newest first)

---

### Edge Cases

- What happens when a parent tries to add an extremely large amount (e.g., $999,999.99)? The system should accept it if within configured limits.
- How does the system handle concurrent modifications (parent A and parent B both modifying the same child's balance)? The system should process transactions sequentially and maintain data integrity.
- What happens if a child is removed from the family? Their balance data should be retained for record-keeping but become inaccessible.
- How are decimal amounts handled? The system should support cents (2 decimal places) and round appropriately.
- What if a child has hundreds of transactions? The system should paginate or load more as needed to maintain performance.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST maintain an account balance for each child user
- **FR-002**: System MUST display account balances in the local currency format (e.g., $XX.XX for USD)
- **FR-003**: System MUST allow parents to add money to any of their children's accounts
- **FR-004**: System MUST allow parents to remove money from any of their children's accounts
- **FR-005**: System MUST prevent withdrawals that would result in a negative balance
- **FR-006**: System MUST display all children's balances on the parent dashboard
- **FR-007**: System MUST display the child's own balance on their dashboard (read-only)
- **FR-008**: System MUST display the child's transaction history on their dashboard (read-only)
- **FR-009**: System MUST NOT provide any balance or transaction modification controls to child users
- **FR-010**: System MUST record each balance modification as a transaction with timestamp, amount, type (deposit/withdrawal), and optional note
- **FR-011**: System MUST display transaction history in reverse chronological order (newest first)
- **FR-012**: System MUST ensure a parent can only view and modify balances for their own children
- **FR-013**: System MUST ensure a child can only view their own balance and transaction history
- **FR-014**: System MUST support monetary amounts with up to 2 decimal places (cents)
- **FR-015**: System MUST validate that deposit/withdrawal amounts are positive numbers greater than zero

### Key Entities

- **Account Balance**: The current monetary amount held in a child's account. Attributes include current balance, currency, and last modified timestamp.
- **Transaction**: A record of money added or removed from an account. Attributes include amount, type (deposit/withdrawal), note/description, timestamp, parent who made the change, and child whose account was affected.
- **Parent-Child Relationship**: The link that authorizes a parent to view and modify a child's balance. A parent can have multiple children; a child belongs to one family.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Parents can view all their children's balances within 2 seconds of loading the dashboard
- **SC-002**: Parents can complete a deposit or withdrawal transaction in under 30 seconds
- **SC-003**: 100% of balance modifications are accurately reflected in both parent and child views
- **SC-004**: Children can view their balance and transaction history within 2 seconds of loading their dashboard
- **SC-005**: Zero unauthorized balance modifications occur (children cannot modify their own or others' balances, parents cannot access other families' accounts)
- **SC-006**: Transaction history displays all transactions with complete details (date, amount, type, description)

## Assumptions

- Users (parents and children) are already authenticated via the existing user authentication system (feature 001-user-auth)
- The parent-child relationship has already been established in the system
- Currency is assumed to be USD; multi-currency support is out of scope for this feature
- There is no maximum balance limit (balances can grow indefinitely within system constraints)
- Notifications for balance changes are out of scope for this feature

## Out of Scope

- Transaction history viewing for parents (parents see current balances only, not detailed history per child)
- Scheduled/recurring deposits (e.g., automatic weekly allowance)
- Interest calculations
- Multiple currencies
- Transfer between children's accounts
- Spending categories or budgeting features
- Push notifications for balance changes
- Exporting transaction history
