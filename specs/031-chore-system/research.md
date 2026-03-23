# Research: Chore & Task System

**Feature Branch**: `031-chore-system`
**Created**: 2026-03-22

## R1: Transaction Type Extension

**Decision**: Add `TransactionTypeChore TransactionType = "chore"` to `models/transaction.go` and a `DepositChore` method to `TransactionRepo`.

**Rationale**: The existing transaction system already supports typed deposits (deposit, withdrawal, allowance, interest). Adding a "chore" type follows the exact same pattern as `DepositAllowance` — atomic insert + balance update in a GORM transaction. No schema change needed on the `transactions` table since `transaction_type` is a VARCHAR. The `schedule_id` column can be repurposed or left null; a chore-specific reference can go in the `note` field.

**Alternatives considered**:
- Reuse "deposit" type with a note prefix: Rejected — loses the ability to filter/query chore earnings distinctly.
- Add a `chore_instance_id` FK column to transactions: Considered but deferred — the note field suffices for MVP and avoids a migration on the heavily-used transactions table.

## R2: Recurring Chore Scheduling

**Decision**: Follow the allowance scheduler pattern — a `ChoreScheduler` goroutine with `time.NewTicker` and a stop channel. It runs every 5 minutes, generates new instances for active recurring chores whose current period has started, and expires instances whose period has ended without completion.

**Rationale**: The allowance scheduler (`internal/allowance/scheduler.go`) is proven, handles timezone-aware scheduling via the family's IANA timezone setting, and integrates cleanly via `main.go`. The chore scheduler does two things per tick: (1) create new instances where none exist for the current period, (2) mark expired instances for periods that have passed.

**Alternatives considered**:
- Cron-based scheduling: Rejected — the app already uses in-process tickers; adding cron adds a dependency and operational complexity.
- Event-driven (create next instance on approval): Rejected — doesn't handle the "missed/expired" case when a child never completes the chore.

## R3: Chore Instance State Machine

**Decision**: Chore instances follow this state machine:

```
available → pending_approval → approved (terminal)
                             → rejected → available (resets)
available → expired (terminal, recurring only)
```

**Rationale**: The states map directly to the user stories. "Available" means the child can act. "Pending approval" means the parent needs to review. "Approved" triggers deposit and is final. "Rejected" returns to available so the child can retry. "Expired" handles the case where a recurring period elapses without completion.

**Alternatives considered**:
- Separate "rejected" terminal state: Rejected — the spec says rejection returns the chore to available, so the child gets another chance.
- No "expired" state (just delete old instances): Rejected — preserving expired instances lets parents see what was missed.

## R4: Multi-Child Assignment

**Decision**: Use a `chore_assignments` join table between `chores` and `children`. When a chore is created, one `chore_instance` is generated per assigned child (for one-time chores) or per assigned child per period (for recurring chores).

**Rationale**: A chore like "Clean your room" may apply to multiple children independently. Each child needs their own instance to track their own completion status. The join table keeps assignments queryable and editable without duplicating chore definitions.

**Alternatives considered**:
- Duplicate the chore row per child: Rejected — makes editing (rename, change reward) require updating N rows.
- Single instance shared by all children: Rejected — children complete independently; one child's completion shouldn't affect another's.

## R5: Reward Amount Snapshotting

**Decision**: The `chore_instance` records the `reward_cents` at creation time, copied from the parent `chore`. Approval deposits this snapshotted amount, not the chore's current reward.

**Rationale**: The spec explicitly requires (FR-014) that mid-edit changes do not alter pending payouts. Snapshotting at instance creation is the simplest approach — no need to track "reward at time of completion" vs "reward at time of approval."

**Alternatives considered**:
- Snapshot at completion time (when child marks done): Viable but creates a window where a parent edits between creation and completion, which is a more common scenario than editing between completion and approval.
- Always use current chore reward: Rejected — violates FR-014.

## R6: Disabled Account Handling

**Decision**: The `HandleApprove` handler checks `child.IsDisabled` before processing approval. If disabled, return 403 with a clear error message. The scheduler also skips instance generation for disabled children.

**Rationale**: Consistent with existing disabled-child enforcement (028-child-account-limits). The check happens at the action boundary (approval and instance generation), not at view time — disabled children can still see their chore list.

**Alternatives considered**:
- Block at completion time too: Rejected — letting children mark chores as done even when disabled is harmless and avoids confusion; the block at approval prevents any financial impact.
