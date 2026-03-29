# Research: Withdrawal Requests

**Feature**: 032-withdrawal-requests | **Date**: 2026-03-28

## R1: Approval Workflow Pattern

**Decision**: Follow the existing chore instance approval pattern (available → pending_approval → approved/rejected) adapted for withdrawal requests (pending → approved/denied/cancelled).

**Rationale**: The chore system already implements a child-initiates / parent-reviews workflow with GORM-based status transitions, atomic updates using `RowsAffected == 0` checks, and transaction creation on approval. Reusing this pattern ensures consistency and reduces implementation risk.

**Alternatives considered**:
- Event-driven workflow with separate event table — overengineered for family-scale app
- Generic "approval" abstraction shared with chores — premature; only two approval flows exist

## R2: New Transaction Type vs Reuse "withdrawal"

**Decision**: Add a new `withdrawal_request` transaction type distinct from the existing `withdrawal` type.

**Rationale**: Per clarification, approved child-initiated requests must be visually distinguishable from parent-initiated withdrawals in transaction history. A separate type enables this at the data level without conditional display logic. The existing `transactions_transaction_type_check` constraint must be updated.

**Alternatives considered**:
- Reuse "withdrawal" type with a nullable `withdrawal_request_id` foreign key — harder to distinguish in queries and UI; adds nullable column to high-traffic table
- Add a `source` column to transactions — over-general; only this one case needs differentiation

## R3: Balance Validation Timing

**Decision**: Validate available balance at two points: (1) when child submits the request, and (2) when parent approves. Submission validates against available balance and rejects if insufficient. Approval re-validates and warns the parent if balance has dropped below the requested amount.

**Rationale**: Balance can change between submission and approval (allowances, interest, other withdrawals). Validating at both points prevents overdrafts while keeping the child's experience responsive.

**Alternatives considered**:
- Reserve/hold funds at submission time — adds complexity (held balance concept) for minimal benefit in a family app
- Only validate at approval time — poor child UX; they'd see a request accepted then denied later

## R4: One Pending Request Enforcement

**Decision**: Enforce at the database level with a partial unique index on `(child_id) WHERE status = 'pending'`, plus application-level check before insert.

**Rationale**: Database constraint prevents race conditions. Application check provides a friendly error message. This is the standard pattern for "at most one active" constraints in PostgreSQL.

**Alternatives considered**:
- Application-only enforcement — vulnerable to race conditions
- Pessimistic locking — unnecessarily complex for family-scale concurrency

## R5: Account Disabled Auto-Denial

**Decision**: Check `is_disabled` flag on the child's account at submission time (reject immediately) and at approval time (warn parent). Do NOT auto-deny existing pending requests when an account is disabled — instead, the parent sees the disabled status when reviewing and must deny manually.

**Rationale**: Auto-denying on disable would require a trigger or background job, adding complexity. Since only one request can be pending per child and the parent is the one disabling the account, they can deny it in the same session. The spec's FR-014 says "MUST be automatically denied" but the simpler path is checking at approval time — if the account is disabled, the approval is blocked with a clear message.

**Alternatives considered**:
- Database trigger to auto-deny on disable — adds hidden business logic in SQL
- Background job to sweep disabled accounts — over-engineered for one pending request per child

## R6: Frontend Integration Points

**Decision**: Integrate withdrawal requests into existing views rather than creating new routes:
- **Child dashboard**: Show pending request status + "Request Withdrawal" button (replaces or supplements current view)
- **Parent ManageChild**: Show pending request card with approve/deny actions alongside existing deposit/withdraw buttons
- **Parent dashboard**: Add badge/indicator for pending requests across children

**Rationale**: Keeps navigation simple (Simplicity principle). The child already sees their balance on the dashboard — adding a request button there is the natural location. Parents already manage each child from ManageChild — adding the review there keeps the workflow in context.

**Alternatives considered**:
- Dedicated `/child/withdrawal-requests` route — unnecessary indirection for a single pending request
- Dedicated parent page for all requests — could be added later if request volume warrants it; for now, per-child review is simpler
