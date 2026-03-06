# Research: Free Tier Child Account Limits

## R1: Disabled vs. Locked — Two Separate States

**Decision**: Add a new `is_disabled` boolean column to the `children` table, separate from the existing `is_locked` column.

**Rationale**: `is_locked` is for failed login lockouts (auto-triggered after 5 bad passwords, reset by parent password change). `is_disabled` is a business-level restriction tied to subscription tier. They have different triggers, different resolution paths, and different user-facing messaging. Conflating them would create confusing UX and fragile logic.

**Alternatives considered**:
- Single `status` enum (`active`, `locked`, `disabled`): Would break existing `is_locked` logic and create impossible states (locked AND disabled). Rejected.
- Reuse `is_locked` with a reason field: Over-engineered for two booleans. Rejected.

## R2: Where to Enforce the Free Tier Limit

**Decision**: Enforce at child creation time in `HandleCreateChild` (backend) and at subscription change time in webhook handlers. The frontend also filters disabled children from dashboard selection but the backend is the source of truth.

**Rationale**: The backend must be the enforcement point for security. The frontend provides UX by showing disabled state visually. Dual enforcement (backend guards + frontend UX) follows the existing pattern used for `is_locked`.

**Enforcement points**:
1. `HandleCreateChild` — determine if new child should be disabled based on family account_type and current enabled child count
2. `HandleChildLogin` — reject disabled children
3. `HandleDeposit` / `HandleWithdraw` — reject transactions for disabled children
4. Allowance `ListDue` query — filter out disabled children via SQL JOIN
5. Interest `ListDue` / `ListDueForInterest` queries — filter out disabled children via SQL JOIN
6. Webhook handlers — enable/disable children on subscription changes

## R3: Subscription Webhook Integration Points

**Decision**: Hook into three existing webhook handlers to manage child enable/disable:

1. **`handleCheckoutCompleted`** — After `UpdateSubscriptionFromCheckout()`, call `EnableAllChildren(familyID)` to enable all disabled children.
2. **`handleSubscriptionDeleted`** — After `ClearSubscription()`, look up the family by subscription ID (before clearing), then call `DisableExcessChildren(familyID, 2)`.
3. **`handleSubscriptionUpdated`** — No child state changes here. Subscription cancellation at period end doesn't immediately remove Plus — the family stays Plus until `subscription.deleted` fires.

**Rationale**: `checkout.session.completed` is the upgrade event. `customer.subscription.deleted` is the definitive downgrade event (Stripe fires it when the subscription actually ends). `customer.subscription.updated` with `cancel_at_period_end=true` is just a cancellation notice — the user retains Plus until period end.

**Key detail**: `handleSubscriptionDeleted` calls `ClearSubscription(subID)` which updates by `stripe_subscription_id`. We need to look up the family ID BEFORE clearing the subscription, because `ClearSubscription` NULLs the `stripe_subscription_id` field. Use `GetFamilyByStripeSubscriptionID(subID)` first.

## R4: Child Ordering for Disable-on-Downgrade

**Decision**: Use the `id` column (SERIAL auto-incrementing) to determine child creation order. The earliest 2 children by `id` remain enabled; all others are disabled.

**Rationale**: The `id` column is auto-incrementing and reflects insertion order. Using `created_at` would also work but `id` is simpler and avoids any timestamp precision concerns. The spec states "earliest 2 added" which maps directly to lowest IDs.

## R5: Re-evaluation on Child Deletion

**Decision**: When a child is deleted from a free-tier family, re-evaluate whether any disabled children should be enabled. If the count of enabled children drops below 2 and disabled children exist, enable the earliest disabled child(ren) to bring the count to 2.

**Rationale**: Without re-evaluation, deleting an enabled child would leave the family with 1 active + N disabled children even though they're entitled to 2 active. This is the behavior described in the spec's edge cases.

**Implementation**: Add a `ReconcileChildLimits(familyID int64, limit int)` method to the child store that enables/disables children as needed based on the current family account type and the limit. Call it after child deletion and after subscription changes.

## R6: Current Max Children Limit Bug

**Decision**: The current `HandleCreateChild` has `if count >= 2` which actually limits families to 2 children total. The error message says "up to 20 children." This feature changes the semantics: the hard cap remains at 20, but free-tier families get children 3+ created as disabled.

**Rationale**: The existing `count >= 2` was intentionally set as a temporary measure before subscription tiers existed. Now that we're implementing the tiered limit, we raise the hard cap check to 20 and add the disabled logic for free-tier families.

## R7: Frontend Disabled Child UX in ChildSelectorBar

**Decision**: Disabled children in ChildSelectorBar will be:
- Grayed out (reduced opacity, muted colors)
- Not selectable (clicking shows a tooltip/popover instead of selecting)
- Display a brief explanation with a CTA link to `/settings/subscription`

**Rationale**: The existing pattern shows a lock icon for `is_locked` children. For disabled children, we need a stronger visual treatment since it's a business limitation, not a temporary lockout. A tooltip on click (rather than hover) works better on mobile.
