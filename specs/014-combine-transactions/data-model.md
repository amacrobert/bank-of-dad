# Data Model: Combine Transaction Cards

**Feature**: 014-combine-transactions
**Date**: 2026-02-16

## Overview

No data model changes are required. This feature is a purely presentational refactor. All existing entities, relationships, and API contracts remain unchanged.

## Existing Entities (Unchanged)

### Transaction (recent/past)
- **Source**: `GET /api/children/{id}/transactions`
- **Fields**: id, child_id, parent_id, amount_cents, type (deposit|withdrawal|allowance|interest), note, schedule_id, created_at
- **Frontend type**: `Transaction` in `types.ts`

### UpcomingAllowance
- **Source**: `GET /api/children/{childId}/upcoming-allowances`
- **Fields**: amount_cents, next_date, note
- **Frontend type**: `UpcomingAllowance` in `types.ts`

### InterestSchedule
- **Source**: `GET /api/children/{childId}/interest-schedule`
- **Fields**: id, child_id, parent_id, frequency, day_of_week, day_of_month, status, next_run_at, created_at, updated_at
- **Frontend type**: `InterestSchedule` in `types.ts`

### UpcomingPayment (frontend-only, derived)
- **Defined in**: `UpcomingPayments.tsx` (currently) → `TransactionsCard.tsx` (after refactor)
- **Fields**: type (allowance|interest), amountCents, estimated (boolean), date, note
- **Purpose**: Unified shape for rendering upcoming allowance and interest entries together
- **Not persisted**: Computed in the component from UpcomingAllowance + InterestSchedule data

## State Transitions

None. No new state is introduced and no existing state machines are modified.

## Migration

None required. No database or API changes.
