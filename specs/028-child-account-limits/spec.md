# Feature Specification: Free Tier Child Account Limits

**Feature Branch**: `028-child-account-limits`
**Created**: 2026-03-06
**Status**: Draft
**Input**: User description: "Limit free tier to 2 active children per bank. Children beyond the 2nd are added in a disabled state. Disabled accounts are restricted from transactions, allowance, interest, parent dashboard management, and login. Upgrading to Plus enables all children; downgrading disables children beyond the earliest 2."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Free Tier Parent Adds a Third Child (Priority: P1)

A parent on the free tier already has 2 active children in their bank. They navigate to add a third child. The system allows the child to be created but places the account in a "disabled" state. The parent sees a visual indicator on the disabled child in the ChildSelectorBar with an explanation that upgrading to Plus removes the limit.

**Why this priority**: This is the core monetization gate — the moment a parent hits the free tier limit is the primary conversion opportunity.

**Independent Test**: Can be fully tested by creating a free-tier family with 2 children, adding a third, and verifying the third child is created in a disabled state with appropriate UI indicators.

**Acceptance Scenarios**:

1. **Given** a free-tier parent with 2 active children, **When** they add a third child, **Then** the child is created successfully in a disabled state.
2. **Given** a free-tier parent with 2 active children, **When** they add a third child, **Then** the ChildSelectorBar shows the disabled child with a visually distinct appearance (muted/grayed out).
3. **Given** a disabled child appears in the ChildSelectorBar, **When** the parent interacts with it, **Then** a tooltip or overlay explains the free tier limit and provides a link to `/settings/subscription` to upgrade.
4. **Given** a free-tier parent with 2 active children, **When** they add a 4th or 5th child, **Then** each additional child is also created in a disabled state.

---

### User Story 2 - Disabled Account Restrictions (Priority: P1)

Disabled child accounts are fully restricted: they cannot receive transactions (deposits, withdrawals, allowance, interest), cannot be selected or managed from the parent dashboard (but remain manageable in `/settings/children` for editing name, password, avatar, or deletion), and cannot log in.

**Why this priority**: Enforcing restrictions is essential for the limit to have meaning — without enforcement, there's no incentive to upgrade.

**Independent Test**: Can be tested by attempting each restricted action against a disabled child and verifying all are blocked.

**Acceptance Scenarios**:

1. **Given** a disabled child account, **When** the allowance scheduler runs, **Then** the child's allowance schedule is skipped.
2. **Given** a disabled child account, **When** the interest scheduler runs, **Then** the child's interest accrual is skipped.
3. **Given** a disabled child account, **When** a parent attempts to deposit or withdraw from the parent dashboard, **Then** the action is prevented (the child is not selectable for transactions).
4. **Given** a disabled child account, **When** the child attempts to log in, **Then** login is rejected with a message indicating the account is disabled.
5. **Given** a disabled child account, **When** a parent visits `/settings/children`, **Then** the child appears and can be edited (name, password, avatar) or deleted.
6. **Given** a disabled child in the ChildSelectorBar, **When** the parent clicks on the disabled child, **Then** the child is NOT selected for dashboard management (deposits, withdrawals, growth projector, etc.).

---

### User Story 3 - Upgrade Enables All Children (Priority: P2)

When a parent upgrades from free to Plus, all disabled child accounts in their family are automatically enabled. The children immediately become fully functional — they can log in, receive transactions, collect allowance and interest, and appear as active in the parent dashboard.

**Why this priority**: This is the payoff for upgrading — the parent should see immediate value.

**Independent Test**: Can be tested by creating a free-tier family with disabled children, upgrading to Plus, and verifying all children become enabled.

**Acceptance Scenarios**:

1. **Given** a free-tier family with 2 active and 2 disabled children, **When** the parent upgrades to Plus, **Then** the 2 disabled children become enabled.
2. **Given** children just enabled by upgrade, **When** the parent views the dashboard, **Then** all 4 children appear as active and selectable in the ChildSelectorBar.
3. **Given** a Plus family, **When** the parent adds more children (up to the hard limit), **Then** all new children are created in an enabled state.

---

### User Story 4 - Downgrade Disables Excess Children (Priority: P2)

When a family loses its Plus status (subscription cancelled, expired, or payment failed), any children beyond the earliest 2 added are disabled. The earliest 2 children (by creation date) remain active.

**Why this priority**: Ensures the free tier limit is re-enforced on downgrade and creates retention pressure.

**Independent Test**: Can be tested by creating a Plus family with 4 children, downgrading, and verifying only the first 2 created remain active.

**Acceptance Scenarios**:

1. **Given** a Plus family with 4 children, **When** the subscription is cancelled and the period ends, **Then** the 2 children added most recently (by creation order) are disabled.
2. **Given** a Plus family with 4 children, **When** the subscription is cancelled and the period ends, **Then** the 2 children added earliest (by creation order) remain enabled.
3. **Given** a downgraded family with disabled children, **When** the parent views the ChildSelectorBar, **Then** disabled children show the same visual treatment and upgrade CTA as in the free-tier add scenario.

---

### Edge Cases

- What happens when a free-tier parent has exactly 2 children and deletes one, then adds a new one? The new child should be active (they are back under the limit).
- What happens when a Plus parent with 5 children deletes 3, then downgrades? The remaining 2 children should both stay active (they are at or under the limit).
- What happens when a family downgrades and has exactly 2 children? No children are disabled.
- What happens when a disabled child's allowance or interest schedule fires? The scheduler skips it silently — no error, no transaction, no side effects.
- What happens to a disabled child's existing balance? It is preserved. The balance remains visible in `/settings/children` but not in the parent dashboard.
- What happens if a parent tries to create children beyond the hard cap (20)? The existing hard cap rejection applies before the disabled logic.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow free-tier parents to create more than 2 children, up to the existing hard cap (20).
- **FR-002**: System MUST automatically set children beyond the 2nd to a "disabled" state when the family is on the free tier.
- **FR-003**: Disabled child accounts MUST NOT receive deposits, withdrawals, or any manual transactions.
- **FR-004**: Disabled child accounts MUST be skipped by the allowance scheduler — no allowance is deposited.
- **FR-005**: Disabled child accounts MUST be skipped by the interest scheduler — no interest is accrued.
- **FR-006**: Disabled child accounts MUST NOT be able to log in. Login attempts MUST return a clear error indicating the account is disabled due to the free tier limit.
- **FR-007**: Disabled child accounts MUST still appear in `/settings/children` and be editable (name, password, avatar) and deletable.
- **FR-008**: Disabled child accounts MUST NOT be selectable in the ChildSelectorBar for dashboard operations (transactions, growth projector, etc.).
- **FR-009**: The ChildSelectorBar MUST display disabled children with a visually distinct appearance (muted/grayed out styling).
- **FR-010**: The ChildSelectorBar MUST show a tooltip or overlay on disabled children explaining the free tier limit, with a call-to-action linking to `/settings/subscription`.
- **FR-011**: When a family upgrades to Plus, the system MUST enable all disabled child accounts in that family.
- **FR-012**: When a family loses Plus status (subscription deleted or expired), the system MUST disable all children beyond the earliest 2 (by creation order).
- **FR-013**: The "earliest 2 children" determination MUST be based on the order children were added (by database record creation time or ID order).
- **FR-014**: A Plus-tier parent adding children MUST have all new children created in an enabled state (up to the hard cap).
- **FR-015**: When a free-tier parent deletes children bringing the active count to fewer than 2, the system MUST re-evaluate and enable the next earliest disabled child to reach 2 active children.

### Key Entities

- **Child Account**: Existing entity, gains a new "disabled" status attribute (distinct from the existing "locked" status which is for failed login attempts).
- **Family**: Existing entity, already has `account_type` ("free" or "plus") used to determine the active child limit.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Free-tier parents can add up to 20 children, with children beyond the 2nd created in a disabled state, within the same time it takes to add an enabled child.
- **SC-002**: 100% of disabled child accounts are blocked from transactions, allowance, interest, and login.
- **SC-003**: Upgrading to Plus enables all disabled children within the same request/webhook processing cycle — no manual intervention required.
- **SC-004**: Downgrading from Plus correctly disables the right children (beyond the earliest 2) within the same webhook processing cycle.
- **SC-005**: Disabled children in the ChildSelectorBar clearly communicate the upgrade path — the tooltip/CTA is visible and links to the subscription page.

## Assumptions

- The free tier child limit is 2 active children. This is a business constant that may be extracted to configuration in the future but is hardcoded for now.
- The existing hard cap of 20 children per family remains unchanged.
- "Earliest 2 children" is determined by the child's database ID (auto-incrementing), which reflects creation order.
- The disabled state is a new boolean field, separate from the existing `is_locked` field (which handles failed login lockouts).
- Existing families on the free tier with more than 2 children (if any exist from before this feature) will need their excess children disabled. This can be handled via a one-time migration or on next relevant action.
- The subscription webhook handlers (`customer.subscription.updated`, `customer.subscription.deleted`) are the trigger points for enabling/disabling children on plan changes.
