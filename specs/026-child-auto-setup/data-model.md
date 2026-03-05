# Data Model: Child Auto-Setup

## Overview

No new database tables or columns are needed. This feature orchestrates existing entities through existing API endpoints.

## Existing Entities Used

### Child (existing — no changes)
- Created via `POST /api/children`
- Fields: id, family_id, first_name, password_hash, balance_cents (starts at 0), avatar, theme

### Transaction (existing — no changes)
- Created via `POST /api/children/{id}/deposit`
- An "Initial deposit" transaction record with type="deposit"
- Atomically updates child.balance_cents

### AllowanceSchedule (existing — no changes)
- Created via `PUT /api/children/{id}/allowance`
- Fields: child_id, parent_id, amount_cents, frequency="weekly", day_of_week=(current day), note="Weekly allowance", status="active", next_run_at=(calculated)

### InterestSchedule (existing — no changes)
- Created via `PUT /api/children/{id}/interest`
- Fields: child_id, parent_id, frequency="monthly", day_of_month=1, status="active", next_run_at=(calculated)
- Also sets interest_rate_bps on the child's interest record

## Data Flow

```
Parent submits Add Child form
  │
  ├─ 1. POST /api/children { first_name, password, avatar? }
  │     → Creates Child record (balance_cents = 0)
  │     → Returns { id, first_name, family_slug, login_url }
  │
  ├─ 2. POST /api/children/{id}/deposit { amount_cents, note: "Initial deposit" }
  │     → Creates Transaction record (type="deposit")
  │     → Updates Child.balance_cents atomically
  │     (skipped if initial_deposit is 0 or empty)
  │
  ├─ 3. PUT /api/children/{id}/allowance { amount_cents, frequency: "weekly", day_of_week, note: "Weekly allowance" }
  │     → Creates AllowanceSchedule record (status="active")
  │     → Calculates next_run_at based on family timezone
  │     (skipped if weekly_allowance is 0 or empty)
  │
  └─ 4. PUT /api/children/{id}/interest { interest_rate_bps, frequency: "monthly", day_of_month: 1 }
        → Sets interest rate on child
        → Creates InterestSchedule record (status="active")
        → Calculates next_run_at based on family timezone
        (skipped if annual_interest is 0 or empty)
```

## Migrations

None required.
