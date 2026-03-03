# Feature Specification: Stripe Subscription Integration

**Feature Branch**: `024-stripe-subscription`
**Created**: 2026-02-26
**Status**: Draft
**Input**: User description: "Add Stripe integration. Parents will be able to upgrade their free account to a Plus account, for a $2/month (or $20/year) subscription via Stripe. Accounts remain free, upgrading is not required to access and use the application. Add a new /settings/subscription page for parents to upgrade or downgrade their account. No features will change at this time for a Plus vs a Free account. In the future, there will be changes. Store the account type in the family table."

## User Scenarios & Testing

### User Story 1 - Parent Upgrades to Plus (Priority: P1)

A parent navigates to the subscription settings page and chooses to upgrade their family's free account to Plus. They select either a monthly ($2/month) or annual ($20/year) billing plan and complete payment through Stripe's checkout flow. After successful payment, their account type changes to Plus immediately.

**Why this priority**: This is the core revenue-generating flow. Without the ability to upgrade, the feature has no value.

**Independent Test**: Can be fully tested by signing in as a parent, navigating to /settings/subscription, selecting a plan, completing Stripe checkout, and verifying the account shows as Plus.

**Acceptance Scenarios**:

1. **Given** a parent with a free account, **When** they navigate to the subscription page and select the monthly plan, **Then** they are redirected to Stripe's checkout page with the $2/month plan pre-selected.
2. **Given** a parent is on Stripe's checkout page, **When** they complete payment successfully, **Then** they are redirected back to the subscription page and their account type shows "Plus".
3. **Given** a parent with a free account, **When** they select the annual plan, **Then** they are redirected to Stripe's checkout page with the $20/year plan pre-selected.
4. **Given** a parent is on Stripe's checkout page, **When** they cancel or abandon the checkout, **Then** they are returned to the subscription page and their account remains free.

---

### User Story 2 - Parent Views Subscription Status (Priority: P2)

A parent navigates to the subscription settings page and sees their current account type (Free or Plus), billing period if subscribed, and options to change their plan.

**Why this priority**: Users need to understand their current subscription state and have clear options for managing it.

**Independent Test**: Can be fully tested by signing in as a parent and navigating to /settings/subscription, verifying the correct account type and available actions are displayed.

**Acceptance Scenarios**:

1. **Given** a parent with a free account, **When** they visit the subscription page, **Then** they see their current plan is "Free" and are presented with options to upgrade to Plus (monthly or annual).
2. **Given** a parent with an active Plus subscription, **When** they visit the subscription page, **Then** they see their current plan is "Plus", their billing period (monthly/annual), and an option to manage or cancel their subscription.

---

### User Story 3 - Parent Cancels Plus Subscription (Priority: P3)

A parent with a Plus subscription decides to cancel. They navigate to the subscription page and choose to cancel. Their Plus access continues until the end of the current billing period, then reverts to Free.

**Why this priority**: Cancellation is essential for user trust and is required by most payment platform policies and consumer protection laws.

**Independent Test**: Can be fully tested by signing in as a Plus parent, cancelling the subscription, and verifying access continues until period end before reverting to Free.

**Acceptance Scenarios**:

1. **Given** a parent with an active Plus subscription, **When** they choose to cancel, **Then** they are shown a confirmation prompt explaining that Plus will remain active until the end of the current billing period.
2. **Given** a parent confirms cancellation, **When** the cancellation is processed, **Then** their subscription is marked as cancelled but Plus access continues until the billing period ends.
3. **Given** a cancelled subscription reaches the end of its billing period, **When** the period expires, **Then** the family's account type reverts to Free.

---

### User Story 4 - Parent Switches Billing Period (Priority: P4)

A parent with an active Plus subscription wants to switch between monthly and annual billing. They can do so from the subscription page.

**Why this priority**: Nice-to-have convenience for existing subscribers. Lower priority since initial subscription choice covers most users.

**Independent Test**: Can be fully tested by signing in as a Plus parent on monthly billing, switching to annual, and verifying the plan change takes effect at the next billing cycle.

**Acceptance Scenarios**:

1. **Given** a parent with a monthly Plus subscription, **When** they choose to switch to annual billing, **Then** the change is scheduled to take effect at the next billing cycle and the parent sees confirmation.
2. **Given** a parent with an annual Plus subscription, **When** they choose to switch to monthly billing, **Then** the change is scheduled to take effect at the next billing cycle.

---

### Edge Cases

- What happens if a Stripe webhook is received but the family cannot be found? The system logs the error and ignores the event.
- What happens if Stripe checkout succeeds but the webhook delivery fails? The system retries via Stripe's built-in webhook retry mechanism. The account type is only updated via webhook confirmation, not on redirect.
- What happens if a parent tries to upgrade when they already have an active Plus subscription? The subscription page shows their current status and does not offer a duplicate upgrade.
- What happens if a payment fails during subscription renewal? Stripe handles retry logic. If all retries fail, the subscription is cancelled and the account reverts to Free via webhook.
- What happens if two parents in the same family try to manage the subscription simultaneously? The subscription is per-family, and the most recent Stripe state (via webhooks) wins.

## Requirements

### Functional Requirements

- **FR-001**: System MUST allow parents to view their family's current subscription status (Free or Plus) on the subscription settings page.
- **FR-002**: System MUST allow parents to initiate an upgrade from Free to Plus by selecting either a monthly ($2/month) or annual ($20/year) plan.
- **FR-003**: System MUST redirect parents to Stripe's hosted checkout page for payment processing when upgrading.
- **FR-004**: System MUST update the family's account type to Plus upon receiving a confirmed payment event from Stripe.
- **FR-005**: System MUST allow parents to cancel their Plus subscription, with access continuing until the end of the current billing period.
- **FR-006**: System MUST revert the family's account type to Free when a subscription ends or is not renewed.
- **FR-007**: System MUST store the family's account type (Free or Plus) persistently.
- **FR-008**: System MUST process subscription lifecycle events from Stripe via webhooks (payment succeeded, subscription cancelled, subscription expired).
- **FR-009**: System MUST allow parents to switch between monthly and annual billing periods.
- **FR-010**: System MUST NOT gate any existing features behind the Plus account type at this time.
- **FR-011**: System MUST store Stripe customer and subscription identifiers to link families with their Stripe records.

### Key Entities

- **Family**: Existing entity representing a household. Extended with an account type (Free or Plus) and a link to the corresponding Stripe customer record.
- **Subscription**: Represents the billing relationship between a family and the Stripe subscription. Tracks billing period, status (active, cancelled, expired), and Stripe subscription identifier.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Parents can complete the full upgrade flow (from clicking "Upgrade" to seeing their Plus status) in under 2 minutes.
- **SC-002**: Subscription status updates (upgrade, cancellation, expiry) are reflected in the application within 30 seconds of the corresponding payment event.
- **SC-003**: 100% of subscription state changes are driven by confirmed payment events, not by client-side redirects alone.
- **SC-004**: Parents can cancel their subscription in 3 clicks or fewer from the subscription page.
- **SC-005**: The subscription page clearly communicates the current plan, billing period, renewal date, and available actions at a glance.

## Assumptions

- Stripe is the sole payment processor; no alternative payment methods are needed.
- Stripe Checkout (hosted payment page) is used rather than an embedded payment form, reducing PCI compliance scope.
- Subscription management (cancellation, plan changes) uses Stripe's Customer Portal or direct API calls, not a custom payment form.
- Only parents can manage subscriptions; children cannot access subscription settings.
- The subscription is per-family, not per-parent. Any parent in the family can manage it.
- Webhook verification (signature validation) is used to ensure events are genuinely from Stripe.
- Free accounts have no expiration or trial period; they are the permanent default.
- Prices are in USD.
