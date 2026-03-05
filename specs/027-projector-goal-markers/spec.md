# Feature Specification: Growth Projector Goal Markers

**Feature Branch**: `027-projector-goal-markers`
**Created**: 2026-03-05
**Status**: Draft
**Input**: User description: "As a child, I want the growth projector to show me when I could reach my various savings goals in each scenario. Add a toggle to the Growth Projector chart. When on, label points on each scenario line for when each of the child's savings goals are possible. Expanding (perhaps via hover or tap) a dot shows the goal and how soon it's possible in the scenario."

## User Scenarios & Testing

### User Story 1 - View Goal Milestones on Projection Chart (Priority: P1)

A child opens the Growth Projector and toggles on goal markers. For each active savings goal, a marker dot appears on each scenario line at the point where the projected balance first reaches the goal's remaining amount (target minus already saved). The child can see at a glance when each goal becomes affordable under different scenarios.

**Why this priority**: This is the core value of the feature — it connects the abstract projection lines to concrete, meaningful milestones the child cares about.

**Independent Test**: Can be fully tested by creating a child with at least one active savings goal, opening the Growth Projector, toggling on goal markers, and verifying dots appear at the correct positions on each scenario line.

**Acceptance Scenarios**:

1. **Given** a child has 2 active savings goals ($50 and $100 remaining) and 3 scenarios configured, **When** the child toggles on goal markers, **Then** up to 6 marker dots appear on the chart — one per goal per scenario line — at the week where the projected balance first reaches each goal's remaining amount.
2. **Given** a scenario where the projected balance never reaches a goal amount (e.g., balance depletes before reaching the goal), **When** goal markers are toggled on, **Then** no marker appears for that goal on that scenario line.
3. **Given** goal markers are toggled on, **When** the child changes the time horizon or modifies a scenario, **Then** the markers recalculate and reposition accordingly.

---

### User Story 2 - Inspect Goal Marker Details (Priority: P1)

When a child hovers over (desktop) or taps (mobile) a goal marker dot, a tooltip or popover appears showing the goal name, the goal's emoji (if set), and how soon the goal is reachable in that scenario (e.g., "3 months from now" or "in 8 weeks").

**Why this priority**: Without the detail view, the dots are meaningless — users need context to understand what each marker represents.

**Independent Test**: Can be tested by hovering/tapping any visible goal marker and verifying the tooltip displays the correct goal name, emoji, and time-to-reach.

**Acceptance Scenarios**:

1. **Given** goal markers are visible on the chart, **When** the child hovers over a marker dot on desktop, **Then** a tooltip appears showing the goal name, emoji, and approximate time to reach (e.g., "Game Console in ~3 months").
2. **Given** goal markers are visible on the chart, **When** the child taps a marker dot on mobile, **Then** the same detail information is displayed.
3. **Given** a marker for a goal with no emoji, **When** the child inspects it, **Then** the tooltip shows just the goal name and time without an empty emoji placeholder.

---

### User Story 3 - Toggle Goal Markers On/Off (Priority: P2)

A toggle control near the chart allows the child to show or hide goal markers. The toggle defaults to off so the chart is not cluttered for children without goals or those who prefer a clean view. The toggle state persists for the duration of the session.

**Why this priority**: The toggle prevents visual clutter and gives children control over the chart complexity. It's lower priority because the markers themselves (P1) must work before the toggle matters.

**Independent Test**: Can be tested by clicking the toggle and verifying markers appear/disappear, then navigating away and back to confirm session persistence.

**Acceptance Scenarios**:

1. **Given** the Growth Projector is loaded, **When** the child views the chart initially, **Then** goal markers are hidden and the toggle is in the "off" position.
2. **Given** goal markers are toggled on, **When** the child toggles them off, **Then** all marker dots are immediately removed from the chart.
3. **Given** a child has no active savings goals, **When** the child views the Growth Projector, **Then** the toggle is still visible but toggling it on shows no markers (no error or empty state message needed).

---

### Edge Cases

- What happens when a goal's remaining amount is $0 (already fully funded)? The marker should not appear since the goal is already reached.
- What happens when two goals have the same remaining amount? Both markers should appear at the same chart position, with the tooltip listing both goals.
- What happens when a goal is reached at week 0 (current balance already exceeds goal)? The marker should appear at the start of the line.
- What happens when the projection horizon is too short to reach a goal? No marker appears for that goal on that scenario.
- What happens when a child has many goals (e.g., 5+)? All reachable goals display markers; the chart may be busy but remains functional.

## Requirements

### Functional Requirements

- **FR-001**: System MUST display a toggle control on the Growth Projector page that shows/hides savings goal markers on the chart.
- **FR-002**: When goal markers are enabled, the system MUST calculate where each active savings goal's remaining amount (target minus saved) is first reached on each scenario projection line.
- **FR-003**: The system MUST render a visual marker (dot) on the chart at each calculated goal-reach point.
- **FR-004**: Each marker dot MUST be visually distinguishable and not obscure the projection line it sits on.
- **FR-005**: Users MUST be able to inspect a marker (hover on desktop, tap on mobile) to see the goal name, emoji (if present), and time-to-reach in that scenario.
- **FR-006**: The time-to-reach MUST be displayed in a human-friendly format (e.g., "in ~3 months", "in ~8 weeks").
- **FR-007**: When a scenario is modified, added, or removed, goal markers MUST recalculate automatically.
- **FR-008**: When the time horizon changes, goal markers MUST recalculate to reflect the new projection range.
- **FR-009**: Goal markers MUST NOT appear for goals that are already fully funded (remaining amount is zero or negative).
- **FR-010**: Goal markers MUST NOT appear on a scenario line if the projected balance never reaches the goal's remaining amount within the selected horizon.
- **FR-011**: The system MUST fetch the child's active savings goals data to calculate marker positions.
- **FR-012**: The toggle MUST default to off when the Growth Projector page is first loaded.

### Key Entities

- **Savings Goal**: Existing entity with target amount, saved amount, status, name, and emoji. The remaining amount (target minus saved) determines where the marker appears on projection lines.
- **Goal Marker**: A derived visual element (not persisted) representing the intersection of a savings goal's remaining amount with a scenario's projection line. Attributes include: goal reference, scenario reference, week index, projected date, and time-to-reach.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Children can identify the projected date for reaching any active savings goal within 5 seconds of enabling goal markers.
- **SC-002**: Goal marker positions update within 1 second of any scenario or horizon change.
- **SC-003**: 100% of displayed goal markers show accurate time-to-reach when inspected (tooltip matches the marker's position on the chart).
- **SC-004**: The feature works on both desktop (hover) and mobile (tap) without layout or usability issues.

## Assumptions

- Goal markers are a frontend-only feature; no new backend endpoints are needed. The existing savings goals endpoint provides all required data.
- The "remaining amount" for a goal is calculated as target minus already saved. Only active (non-completed) goals are considered.
- The toggle state does not persist across page reloads or sessions — it resets to "off" each time the page loads.
- Parents viewing a child's Growth Projector also see the toggle and markers (since they can already see the child's goals).
- The marker color matches the scenario line color it sits on, with a distinct shape or border to make it stand out from the line.

## Scope Boundaries

### In Scope
- Toggle control for showing/hiding goal markers
- Marker dots on projection lines at goal-reach points
- Tooltip/popover with goal name, emoji, and time-to-reach
- Automatic recalculation when scenarios or horizon change

### Out of Scope
- Persisting toggle state across sessions or in user preferences
- Allowing children to create/edit goals from the Growth Projector page
- Animating markers or adding transitions
- Showing completed goals on the chart
- Any backend changes or new API endpoints
