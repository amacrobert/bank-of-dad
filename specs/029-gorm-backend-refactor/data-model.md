# Data Model: GORM Backend Refactor

**Date**: 2026-03-16 | **Feature**: 029-gorm-backend-refactor

> This document defines the GORM model structs that will live in `backend/models/`. Each model maps 1:1 to an existing PostgreSQL table. No schema changes — models must match the current migration-defined schema exactly.

## Entity: Family

**Table**: `families`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| ID | int64 | `primaryKey` | id | SERIAL |
| Slug | string | `uniqueIndex;not null` | slug | TEXT UNIQUE |
| Timezone | string | `not null;default:America/New_York` | timezone | TEXT |
| AccountType | string | `not null;default:free` | account_type | CHECK: free, plus |
| StripeCustomerID | *string | `uniqueIndex` | stripe_customer_id | nullable |
| StripeSubscriptionID | *string | `uniqueIndex` | stripe_subscription_id | nullable |
| SubscriptionStatus | *string | | subscription_status | nullable |
| SubscriptionCurrentPeriodEnd | *time.Time | | subscription_current_period_end | nullable TIMESTAMPTZ |
| SubscriptionCancelAtPeriodEnd | bool | `not null;default:false` | subscription_cancel_at_period_end | |
| BankName | string | `not null;default:Dad` | bank_name | TEXT |
| CreatedAt | time.Time | `autoCreateTime` | created_at | TIMESTAMPTZ |

**Associations**:
- HasMany: Parents, Children

---

## Entity: Parent

**Table**: `parents`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| ID | int64 | `primaryKey` | id | SERIAL |
| GoogleID | string | `uniqueIndex;not null` | google_id | TEXT |
| Email | string | `not null` | email | TEXT |
| DisplayName | string | `not null` | display_name | TEXT |
| FamilyID | int64 | `not null;default:0` | family_id | FK to families |
| CreatedAt | time.Time | `autoCreateTime` | created_at | TIMESTAMPTZ |

**Associations**:
- BelongsTo: Family

---

## Entity: Child

**Table**: `children`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| ID | int64 | `primaryKey` | id | SERIAL |
| FamilyID | int64 | `not null;uniqueIndex:idx_family_child` | family_id | FK to families |
| FirstName | string | `not null;uniqueIndex:idx_family_child` | first_name | UNIQUE(family_id, first_name) |
| PasswordHash | string | `not null` | password_hash | TEXT |
| IsLocked | bool | `not null;default:false` | is_locked | |
| IsDisabled | bool | `not null;default:false` | is_disabled | |
| FailedLoginAttempts | int | `not null;default:0` | failed_login_attempts | |
| BalanceCents | int64 | `not null;default:0` | balance_cents | BIGINT, int64 cents |
| InterestRateBps | int | `not null;default:0` | interest_rate_bps | basis points |
| LastInterestAt | *time.Time | | last_interest_at | nullable TIMESTAMPTZ |
| Avatar | *string | | avatar | nullable TEXT |
| Theme | *string | | theme | nullable TEXT |
| CreatedAt | time.Time | `autoCreateTime` | created_at | TIMESTAMPTZ |
| UpdatedAt | time.Time | `autoUpdateTime` | updated_at | TIMESTAMPTZ |

**Associations**:
- BelongsTo: Family
- HasMany: Transactions, AllowanceSchedules, SavingsGoals, GoalAllocations

---

## Entity: Transaction

**Table**: `transactions`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| ID | int64 | `primaryKey` | id | SERIAL |
| ChildID | int64 | `not null` | child_id | FK to children (CASCADE) |
| ParentID | int64 | `not null` | parent_id | FK to parents (RESTRICT) |
| AmountCents | int64 | `not null` | amount_cents | BIGINT, int64 cents |
| TransactionType | string | `not null` | transaction_type | CHECK: deposit, withdrawal, allowance, interest |
| Note | *string | | note | nullable TEXT |
| ScheduleID | *int64 | | schedule_id | nullable FK to allowance_schedules (SET NULL) |
| CreatedAt | time.Time | `autoCreateTime` | created_at | TIMESTAMPTZ |

**Associations**:
- BelongsTo: Child, Parent
- BelongsTo: AllowanceSchedule (optional, via ScheduleID)

---

## Entity: AllowanceSchedule

**Table**: `allowance_schedules`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| ID | int64 | `primaryKey` | id | SERIAL |
| ChildID | int64 | `not null` | child_id | FK to children (CASCADE) |
| ParentID | int64 | `not null` | parent_id | FK to parents (RESTRICT) |
| AmountCents | int64 | `not null` | amount_cents | BIGINT |
| Frequency | string | `not null` | frequency | CHECK: weekly, biweekly, monthly |
| DayOfWeek | *int | | day_of_week | 0-6, nullable |
| DayOfMonth | *int | | day_of_month | 1-31, nullable |
| Note | *string | | note | nullable TEXT |
| Status | string | `not null;default:active` | status | CHECK: active, paused |
| NextRunAt | *time.Time | | next_run_at | nullable TIMESTAMPTZ |
| CreatedAt | time.Time | `autoCreateTime` | created_at | TIMESTAMPTZ |
| UpdatedAt | time.Time | `autoUpdateTime` | updated_at | TIMESTAMPTZ |

**Associations**:
- BelongsTo: Child, Parent
- HasMany: Transactions (via schedule_id FK)

---

## Entity: InterestSchedule

**Table**: `interest_schedules`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| ID | int64 | `primaryKey` | id | SERIAL |
| ChildID | int64 | `not null;uniqueIndex` | child_id | FK to children (CASCADE), one per child |
| ParentID | int64 | `not null` | parent_id | FK to parents (RESTRICT) |
| Frequency | string | `not null` | frequency | CHECK: weekly, biweekly, monthly |
| DayOfWeek | *int | | day_of_week | 0-6, nullable |
| DayOfMonth | *int | | day_of_month | 1-31, nullable |
| Status | string | `not null;default:active` | status | CHECK: active, paused |
| NextRunAt | *time.Time | | next_run_at | nullable TIMESTAMPTZ |
| CreatedAt | time.Time | `autoCreateTime` | created_at | TIMESTAMPTZ |
| UpdatedAt | time.Time | `autoUpdateTime` | updated_at | TIMESTAMPTZ |

**Associations**:
- BelongsTo: Child, Parent

---

## Entity: RefreshToken

**Table**: `refresh_tokens`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| ID | int64 | `primaryKey` | id | SERIAL |
| TokenHash | string | `uniqueIndex;not null` | token_hash | TEXT |
| UserType | string | `not null` | user_type | CHECK: parent, child |
| UserID | int64 | `not null` | user_id | polymorphic FK |
| FamilyID | int64 | `not null` | family_id | |
| ExpiresAt | time.Time | `not null` | expires_at | TIMESTAMPTZ |
| CreatedAt | time.Time | `autoCreateTime` | created_at | TIMESTAMPTZ |

**Associations**: None (polymorphic user reference, not a traditional FK)

---

## Entity: AuthEvent

**Table**: `auth_events`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| ID | int64 | `primaryKey` | id | SERIAL |
| EventType | string | `not null` | event_type | TEXT |
| UserType | string | `not null` | user_type | TEXT |
| UserID | *int64 | | user_id | nullable INTEGER |
| FamilyID | *int64 | | family_id | nullable INTEGER |
| IPAddress | string | `not null` | ip_address | TEXT |
| Details | *string | | details | nullable TEXT |
| CreatedAt | time.Time | `autoCreateTime` | created_at | TIMESTAMPTZ |

**Associations**: None (audit log, loosely coupled)

---

## Entity: StripeWebhookEvent

**Table**: `stripe_webhook_events`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| StripeEventID | string | `primaryKey` | stripe_event_id | TEXT, not SERIAL |
| EventType | string | `not null` | event_type | TEXT |
| ProcessedAt | time.Time | `autoCreateTime` | processed_at | TIMESTAMPTZ |

**Associations**: None

---

## Entity: SavingsGoal

**Table**: `savings_goals`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| ID | int64 | `primaryKey` | id | SERIAL |
| ChildID | int64 | `not null` | child_id | FK to children (CASCADE) |
| Name | string | `not null` | name | TEXT |
| TargetCents | int64 | `not null` | target_cents | CHECK > 0 |
| SavedCents | int64 | `not null;default:0` | saved_cents | CHECK >= 0 |
| Emoji | *string | | emoji | nullable TEXT |
| Status | string | `not null;default:active` | status | CHECK: active, completed |
| CompletedAt | *time.Time | | completed_at | nullable TIMESTAMPTZ |
| CreatedAt | time.Time | `autoCreateTime` | created_at | TIMESTAMPTZ |
| UpdatedAt | time.Time | `autoUpdateTime` | updated_at | TIMESTAMPTZ |

**Associations**:
- BelongsTo: Child
- HasMany: GoalAllocations

---

## Entity: GoalAllocation

**Table**: `goal_allocations`

| Field | Go Type | GORM Tag | DB Column | Notes |
|-------|---------|----------|-----------|-------|
| ID | int64 | `primaryKey` | id | SERIAL |
| GoalID | int64 | `not null` | goal_id | FK to savings_goals (CASCADE) |
| ChildID | int64 | `not null` | child_id | FK to children (CASCADE) |
| AmountCents | int64 | `not null` | amount_cents | CHECK != 0 |
| CreatedAt | time.Time | `autoCreateTime` | created_at | TIMESTAMPTZ |

**Associations**:
- BelongsTo: SavingsGoal, Child

---

## Relationship Diagram

```
Family (1) ──┬── (*) Parent
              └── (*) Child ──┬── (*) Transaction
                              ├── (*) AllowanceSchedule ── (*) Transaction (via schedule_id)
                              ├── (0..1) InterestSchedule
                              ├── (*) SavingsGoal ── (*) GoalAllocation
                              └── (*) GoalAllocation

RefreshToken (standalone, polymorphic user ref)
AuthEvent (standalone, audit log)
StripeWebhookEvent (standalone, idempotency)
```

## Key Design Decisions

1. **Pointer types for nullable columns**: `*string`, `*int64`, `*time.Time` — GORM handles nil ↔ NULL automatically.
2. **int64 for all money fields**: BalanceCents, AmountCents, TargetCents, SavedCents — no custom types.
3. **GORM autoCreateTime/autoUpdateTime**: Replaces `DEFAULT NOW()` for Go-side timestamp management while keeping DB defaults as fallback.
4. **Composite unique index**: `uniqueIndex:idx_family_child` on (FamilyID, FirstName) for Child.
5. **No GORM AutoMigrate**: Models are read-only reflections of migration-managed schema.
