# Feature Specification: Savings Growth Projector

**Feature Branch**: `016-savings-growth-projector`
**Created**: 2026-02-17
**Status**: Draft
**Input**: User description: "As a child, I want to learn about the power of savings and compound interest, and model my money's potential growth. Create a new page for child accounts that enables them to explore their account's potential/projected growth. This page is navigable via a new navigation item after Home. The main feature of this page is a graph showing projected growth of their account over time. It takes any allowance and compound interest schedules into account. The page will also have inputs that allow children to play with potential scenarios -- like if they spend $X per week, or contribute $Y more now, and things of that nature. Finally, there should be a simple english explanation of what it all means."

## User Scenarios & Testing

### User Story 1 - View Projected Account Growth (Priority: P1)

A child navigates to the new "Growth" page from the sidebar and sees a graph showing how their account balance is projected to grow over time. The projection is based on their current balance, their active allowance schedule, and their active compound interest schedule. The child can see at a glance how their money will grow if they simply leave everything as-is.

**Why this priority**: This is the core value of the feature — giving children a visual understanding of compound growth. Without this, no other scenario is meaningful.

**Independent Test**: Can be fully tested by navigating to the Growth page and verifying the graph renders with correct projections based on the child's current balance, allowance, and interest configuration. Delivers immediate educational value even without scenario controls.

**Acceptance Scenarios**:

1. **Given** a child with a $100 balance, $10/week allowance, and 5% annual interest compounding weekly, **When** they navigate to the Growth page, **Then** they see a graph showing projected balance growth over the default time horizon of 1 year.
2. **Given** a child with a balance but no allowance and no interest configured, **When** they view the Growth page, **Then** the graph shows a flat line at their current balance with a message explaining that no growth sources are configured.
3. **Given** a child with an allowance but no interest, **When** they view the Growth page, **Then** the graph shows linear growth from allowance deposits only.
4. **Given** a child with interest but no allowance, **When** they view the Growth page, **Then** the graph shows compound growth from interest only.

---

### User Story 2 - Explore What-If Scenarios (Priority: P2)

A child adjusts scenario inputs on the Growth page to explore how different behaviors affect their account's future. Available scenarios include: recurring weekly spending, a one-time extra deposit, and a one-time withdrawal. As the child changes these inputs, the graph and explanation update immediately to reflect the new projection.

**Why this priority**: This is the interactive, educational heart of the feature — letting children experiment with financial decisions and see consequences. It builds on the base projection from P1.

**Independent Test**: Can be tested by adjusting each scenario input and verifying that the graph and explanation update correctly. For example, adding $5/week spending should reduce the projected balance compared to the baseline.

**Acceptance Scenarios**:

1. **Given** a child viewing the Growth page with a baseline projection, **When** they enter $5 in the "Weekly spending" input, **Then** the graph updates to show a lower projected balance that accounts for $5 withdrawn each week.
2. **Given** a child viewing the Growth page, **When** they enter $50 in the "One-time extra deposit" input, **Then** the graph updates to show growth starting from a higher initial balance (current balance + $50).
3. **Given** a child viewing the Growth page, **When** they enter $20 in the "One-time withdrawal" input, **Then** the graph updates to show growth starting from a lower initial balance (current balance - $20).
4. **Given** a child with a $30 balance, **When** they enter $50 in the "One-time withdrawal" input, **Then** the system caps the withdrawal at the current balance and displays a message that they cannot withdraw more than they have.
5. **Given** a child adjusting multiple scenario inputs simultaneously (e.g., $3/week spending and $25 extra deposit), **When** the inputs are changed, **Then** the graph reflects the combined effect of all active scenarios.

---

### User Story 3 - Read Plain-English Growth Explanation (Priority: P3)

Below the graph and scenario controls, the child sees a plain-English summary that explains the projected outcome in simple, concrete terms. This summary breaks down the total projected balance into its components: starting principal, interest earned, and new savings from allowance — minus any projected spending. The explanation updates live as scenario inputs change.

**Why this priority**: This makes the numbers accessible and educational for children who may not fully understand graphs. It reinforces the visual with concrete language. It depends on P1 (projection) and P2 (scenarios) being functional.

**Independent Test**: Can be tested by verifying the explanation text matches the projected values shown on the graph, and that it updates when scenario inputs change.

**Acceptance Scenarios**:

1. **Given** a child with a $100 balance, $10/week allowance, and 5% annual interest, **When** they view the Growth page with default settings and a 1-year horizon, **Then** they see an explanation like: "If you keep saving your $10 weekly allowance, in 1 year your account will have $634.57. That's $100 from what you have now, $14.57 from interest, and $520 from allowance deposits."
2. **Given** a child has entered $3/week spending in the scenario inputs, **When** the explanation updates, **Then** it includes spending in the breakdown, e.g., "...minus $156 in spending."
3. **Given** a child has no allowance or interest configured, **When** they view the explanation, **Then** it shows a message like: "Your balance will stay at $100 unless you get an allowance or interest set up. Ask your parent!"
4. **Given** the child changes the time horizon, **When** the explanation updates, **Then** all dollar amounts and the time period in the text reflect the new horizon.

---

### User Story 4 - Adjust Projection Time Horizon (Priority: P4)

A child can change the time horizon for the projection to see shorter-term or longer-term outcomes. Available horizons include 3 months, 6 months, 1 year, 2 years, and 5 years. Changing the horizon updates the graph, the explanation, and all projected values.

**Why this priority**: Enhances the educational value by letting children think in different time scales, but the feature is fully functional with a single default horizon.

**Independent Test**: Can be tested by selecting each time horizon option and verifying the graph x-axis, projected values, and explanation all update appropriately.

**Acceptance Scenarios**:

1. **Given** a child viewing the Growth page, **When** they select "5 years" as the time horizon, **Then** the graph shows 5 years of projected data and the explanation references "in 5 years."
2. **Given** a child viewing the Growth page, **When** they select "3 months" as the time horizon, **Then** the graph shows 3 months of projected data and the explanation references "in 3 months."
3. **Given** a child switches from "1 year" to "2 years" while scenario inputs are active, **When** the horizon changes, **Then** scenario inputs remain and the projection recalculates for the new horizon.

---

### Edge Cases

- What happens when the child's balance is $0 with no allowance or interest? The graph shows a flat line at $0 with an encouraging message to ask their parent about setting up an allowance.
- What happens when weekly spending exceeds weekly allowance income? The graph shows the balance declining over time and eventually reaching $0. The explanation warns: "At this rate, your account will run out of money in X weeks."
- What happens when the projected balance would go negative? The projection floors the balance at $0 — the child cannot owe money. The graph flatlines at $0 once it's reached.
- What happens when interest rate is very high (e.g., 100% annual)? The projection calculates correctly using compound interest math; the graph scales appropriately to show large values.
- What happens if the child's allowance or interest schedule is paused? Paused schedules are excluded from the projection. The explanation notes that their allowance/interest is currently paused.

## Requirements

### Functional Requirements

- **FR-001**: System MUST provide a dedicated Growth page accessible only to child accounts.
- **FR-002**: System MUST add a "Growth" navigation item in the child's sidebar, positioned after "Home."
- **FR-003**: System MUST display a graph showing projected account balance over time, starting from the child's current balance.
- **FR-004**: System MUST incorporate the child's active allowance schedule (amount and frequency) into the growth projection.
- **FR-005**: System MUST incorporate the child's active interest schedule (rate and compounding frequency) into the growth projection.
- **FR-006**: System MUST provide a "Weekly spending" input that allows the child to model recurring weekly expenses, defaulting to $0.
- **FR-007**: System MUST provide a "One-time extra deposit" input that allows the child to model adding money to their account, defaulting to $0.
- **FR-008**: System MUST provide a "One-time withdrawal" input that allows the child to model taking money out of their account, defaulting to $0.
- **FR-009**: One-time withdrawal MUST NOT exceed the child's current balance; the system MUST display a validation message if the child attempts to exceed it.
- **FR-010**: System MUST update the graph and explanation immediately as the child changes any scenario input or time horizon.
- **FR-011**: System MUST display a plain-English explanation that breaks down the projected balance into components: starting principal, interest earned, allowance contributions, and spending deducted.
- **FR-012**: System MUST provide selectable time horizons: 3 months, 6 months, 1 year, 2 years, and 5 years, defaulting to 1 year.
- **FR-013**: System MUST floor projected balances at $0 — the projection MUST NOT show negative balances.
- **FR-014**: System MUST handle the case where no growth sources are configured (no allowance, no interest) by showing a flat-line projection and an encouraging message.
- **FR-015**: System MUST exclude paused allowance and interest schedules from the projection and note their paused status in the explanation.
- **FR-016**: When projected spending causes the balance to reach $0 before the time horizon ends, the explanation MUST warn the child approximately when their money will run out.

### Key Entities

- **Growth Projection**: A calculated time series of projected balance values over a given time horizon, derived from current balance, allowance schedule, interest schedule, and scenario adjustments. Not persisted — computed on demand.
- **Scenario Inputs**: User-adjustable parameters (weekly spending, one-time deposit, one-time withdrawal) that modify the base projection. Transient — not saved between sessions.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Children can view their projected account growth within 2 seconds of navigating to the Growth page.
- **SC-002**: Graph and explanation update within 500 milliseconds of any input change.
- **SC-003**: Projected balance values are accurate to within $0.01 when compared against manual compound interest calculations using the same inputs.
- **SC-004**: 100% of children with at least one active schedule (allowance or interest) see a non-trivial growth projection on their first visit.
- **SC-005**: The plain-English explanation correctly itemizes all balance components (principal, interest, allowance, spending) that sum to the projected total.
- **SC-006**: Children can access the Growth page via the navigation within one click from any page in the application.

## Assumptions

- The projection is calculated entirely on the client side using the child's current balance, allowance schedule, and interest configuration fetched from existing endpoints. No new projection-specific backend endpoint is required.
- Interest compounding in the projection follows the same schedule-based model used by the actual system: interest is applied at the frequency specified in the interest schedule (weekly, biweekly, or monthly), not continuously.
- All monetary values use cents internally and display as formatted dollar amounts to the child.
- Scenario inputs are ephemeral — they reset when the child navigates away and returns. No scenario data is persisted.
- The graph uses a line chart format, which is the clearest representation for showing balance growth over time.
- The child cannot modify their actual balance or schedules from this page — it is purely a read-only projection tool.
- Dollar amounts in the explanation are rounded to the nearest cent.

## Scope Boundaries

### In Scope

- New child-facing Growth page with graph, scenario inputs, and plain-English explanation
- Navigation update to add "Growth" item for child accounts
- Client-side projection calculation using existing account data
- Time horizon selector (3 months to 5 years)
- Three scenario input controls: weekly spending, one-time deposit, one-time withdrawal

### Out of Scope

- Saving or sharing projections
- Goal-setting features (e.g., "I want to save up for a $200 bike")
- Parent-facing projection views
- Historical growth tracking or comparison to past projections
- Multiple scenario comparison (side-by-side)
- Backend projection calculation or new API endpoints beyond what already exists
- Notifications or alerts based on projections

## Dependencies

- Existing child balance endpoint for current balance and interest rate
- Existing allowance schedule endpoint for allowance amount and frequency
- Existing interest schedule endpoint for compounding frequency
- Existing navigation/sidebar component for adding the new menu item
