# Data Model: Withdrawal Requests

**Feature**: 032-withdrawal-requests | **Date**: 2026-03-28

## New Entities

### withdrawal_requests

Represents a child's request to withdraw funds, subject to parent approval.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Unique identifier |
| child_id | BIGINT | NOT NULL, FK → children(id) | Requesting child |
| family_id | BIGINT | NOT NULL, FK → families(id) | Family (for parent queries) |
| amount_cents | INTEGER | NOT NULL, CHECK > 0 | Requested amount in cents |
| reason | VARCHAR(500) | NOT NULL | Child's reason for the request |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | pending / approved / denied / cancelled |
| denial_reason | VARCHAR(500) | NULLABLE | Parent's reason for denial |
| reviewed_by_parent_id | BIGINT | NULLABLE, FK → parents(id) | Parent who approved/denied |
| reviewed_at | TIMESTAMPTZ | NULLABLE | When parent acted |
| transaction_id | BIGINT | NULLABLE, FK → transactions(id) | Resulting transaction (if approved) |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | When request was submitted |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | Last update timestamp |

**Indexes**:
- `idx_withdrawal_requests_child_id` on `(child_id)` — child's request history
- `idx_withdrawal_requests_family_id_status` on `(family_id, status)` — parent pending queries
- `uq_withdrawal_requests_child_pending` UNIQUE on `(child_id) WHERE status = 'pending'` — enforce one pending per child

**Constraints**:
- `chk_status_valid`: `status IN ('pending', 'approved', 'denied', 'cancelled')`
- `chk_amount_positive`: `amount_cents > 0`

## Modified Entities

### transactions

**Change**: Add `withdrawal_request` to the `transaction_type` check constraint.

Current allowed values: `'deposit', 'withdrawal', 'allowance', 'interest', 'chore'`
New allowed values: `'deposit', 'withdrawal', 'allowance', 'interest', 'chore', 'withdrawal_request'`

No new columns added to the transactions table.

## State Transitions

```
                  ┌──────────┐
                  │ pending  │
                  └────┬─────┘
                       │
            ┌──────────┼──────────┐
            │          │          │
            ▼          ▼          ▼
       ┌─────────┐ ┌────────┐ ┌───────────┐
       │approved │ │ denied │ │ cancelled │
       └─────────┘ └────────┘ └───────────┘
```

| From | To | Trigger | Side Effects |
|------|----|---------|--------------|
| pending | approved | Parent approves | Create withdrawal_request transaction; deduct from child balance |
| pending | denied | Parent denies | None (balance unchanged); denial_reason recorded |
| pending | cancelled | Child cancels | None (balance unchanged) |

All terminal states (approved, denied, cancelled) are final — no further transitions allowed.

## Go Model

```go
type WithdrawalRequestStatus string

const (
    WithdrawalRequestStatusPending   WithdrawalRequestStatus = "pending"
    WithdrawalRequestStatusApproved  WithdrawalRequestStatus = "approved"
    WithdrawalRequestStatusDenied    WithdrawalRequestStatus = "denied"
    WithdrawalRequestStatusCancelled WithdrawalRequestStatus = "cancelled"
)

type WithdrawalRequest struct {
    ID                 int64                   `gorm:"primaryKey" json:"id"`
    ChildID            int64                   `gorm:"not null" json:"child_id"`
    FamilyID           int64                   `gorm:"not null" json:"family_id"`
    AmountCents        int                     `gorm:"not null" json:"amount_cents"`
    Reason             string                  `gorm:"not null;size:500" json:"reason"`
    Status             WithdrawalRequestStatus `gorm:"not null;default:pending" json:"status"`
    DenialReason       *string                 `gorm:"size:500" json:"denial_reason,omitempty"`
    ReviewedByParentID *int64                  `json:"reviewed_by_parent_id,omitempty"`
    ReviewedAt         *time.Time              `json:"reviewed_at,omitempty"`
    TransactionID      *int64                  `json:"transaction_id,omitempty"`
    CreatedAt          time.Time               `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt          time.Time               `gorm:"autoUpdateTime" json:"updated_at"`
}
```

## TypeScript Types

```typescript
type WithdrawalRequestStatus = 'pending' | 'approved' | 'denied' | 'cancelled';

interface WithdrawalRequest {
    id: number;
    child_id: number;
    family_id: number;
    amount_cents: number;
    reason: string;
    status: WithdrawalRequestStatus;
    denial_reason?: string;
    reviewed_by_parent_id?: number;
    reviewed_at?: string;
    transaction_id?: number;
    created_at: string;
    updated_at: string;
}
```
