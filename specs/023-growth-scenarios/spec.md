# Feature Specification: Growth Projector Scenarios

**Feature Branch**: `023-growth-scenarios`
**Created**: 2026-02-23
**Status**: Draft
**Input**: User description: "growth-scenarios.md"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Compare Two Growth Scenarios Side-by-Side (Priority: P1)

A child (or parent viewing on behalf of a child) opens the growth projector and sees two scenario lines on the graph by default. Each line represents a different "what if" scenario, allowing instant visual comparison of outcomes. The child can see how saving all their allowance compares to spending all of it, or how making a deposit now affects their future balance versus doing nothing.

**Why this priority**: Side-by-side comparison is the core value proposition of this feature. Without multiple lines on one graph, the user cannot compare scenarios, which is the entire purpose of the enhancement.

**Independent Test**: Can be fully tested by opening the growth projector for a child with an allowance and verifying two distinct colored lines appear on the graph, each with its own scenario description.

**Acceptance Scenarios**:

1. **Given** a child with an active allowance schedule, **When** the growth projector loads with no URL query parameters, **Then** two scenario lines appear on the graph: one for "save all allowance" (weekly spending = 0) and one for "spend all allowance" (weekly spending = 100% of calculated weekly allowance amount).
2. **Given** a child with no allowance schedule, **When** the growth projector loads with no URL query parameters, **Then** two scenario lines appear: one for "no spending" (weekly spending = 0) and one for "spend $5/week" (weekly spending = $5).
3. **Given** two scenarios are displayed, **When** the user hovers over (or taps) a data point on the graph, **Then** a tooltip shows the scenario title, date, and projected balance for that scenario.
4. **Given** two scenarios are displayed, **Then** each scenario line has a distinct color that matches the color shown next to its title in the scenario card.

---

### User Story 2 - Edit Scenario Parameters with Combined Input Fields (Priority: P2)

A user edits a scenario's parameters using simplified, combined input fields. The "one-time deposit" and "one-time withdrawal" fields are merged into a single "One time" field with a deposit/withdrawal toggle. Similarly, the "weekly spending" field becomes a "Weekly" field with a spending/saving toggle. This reduces visual clutter and eliminates mutually exclusive field confusion.

**Why this priority**: The combined fields simplify the interface and reduce the number of distinct scenario permutations. This directly supports the multi-scenario display by making each scenario row more compact and understandable.

**Independent Test**: Can be tested by opening a scenario, toggling the "One time" field between deposit and withdrawal, entering a value, and verifying the graph updates accordingly.

**Acceptance Scenarios**:

1. **Given** a scenario is displayed, **When** the user views the "One time" field, **Then** it shows a single amount input with a toggle to select either "Deposit" or "Withdrawal."
2. **Given** a scenario is displayed, **When** the user views the "Weekly" field, **Then** it shows a single amount input with a toggle to select either "Spending" or "Saving."
3. **Given** a user toggles "One time" from deposit to withdrawal (or vice versa), **When** the toggle changes, **Then** the graph and scenario title update to reflect the new direction.
4. **Given** a user enters a value in the "Weekly" field with "Saving" selected, **Then** the scenario projects additional weekly savings on top of any allowance.

---

### User Story 3 - Dynamic Scenario Titles (Priority: P3)

Each scenario row in the "WHAT IF..." card displays a plain-English title that dynamically describes the scenario based on its inputs. Titles change as the user adjusts parameters, giving immediate human-readable context for each line on the graph.

**Why this priority**: Titles make the graph lines meaningful by translating numbers into relatable sentences. Without them, users must mentally map colors to parameter sets, which is cognitively expensive.

**Independent Test**: Can be tested by changing scenario inputs and verifying the title updates to match the expected phrasing from the title generation rules.

**Acceptance Scenarios**:

1. **Given** a child with allowance $20/week and a scenario with weekly spending = 0, **Then** the title reads: "If I save **all** of my $20 allowance."
2. **Given** a child with allowance $20/week and a scenario with weekly spending = $20, **Then** the title reads: "If I save **none** of my $20 allowance."
3. **Given** a child with allowance $20/week and a scenario with weekly spending = $5, **Then** the title reads: "If I save **$15** per week from my $20 allowance."
4. **Given** a child with allowance $20/week, weekly spending = 0, and a one-time deposit of $50, **Then** the title reads: "If I save **all** of my $20 allowance, and deposit $50 now."
5. **Given** a child with no allowance and weekly spending = 0, **Then** the title reads: "If I don't do anything."
6. **Given** a child with no allowance, weekly savings = $10, **Then** the title reads: "If I save **$10** per week."
7. **Given** a child with allowance $20/week, weekly spending = 0, and weekly savings = $5, **Then** the title reads: "If I save **all** of my $20 allowance plus an additional **$5** per week."

---

### User Story 4 - Add and Remove Scenarios (Priority: P4)

A user can add new scenarios beyond the default two, or remove existing ones (as long as at least one remains). This allows power users to compare many "what if" situations simultaneously.

**Why this priority**: Adding/removing scenarios extends the core comparison feature but is not essential for the default two-scenario experience. Most users will get value from the defaults alone.

**Independent Test**: Can be tested by clicking "Add scenario," configuring it, verifying a third line appears on the graph, then deleting a scenario and verifying it disappears.

**Acceptance Scenarios**:

1. **Given** two scenarios exist, **When** the user adds a new scenario, **Then** a third scenario row appears with default values (weekly spending = 0, no one-time amount) and a third line appears on the graph with a new distinct color.
2. **Given** three scenarios exist, **When** the user deletes one, **Then** two scenarios remain and the graph updates to show only two lines.
3. **Given** only one scenario remains, **Then** the delete option for that scenario is disabled or hidden.

---

### User Story 5 - Bookmarkable Scenario URLs (Priority: P5)

Scenario configurations are persisted in the URL query parameters so that a user can bookmark or share a specific set of scenarios. When the page loads with scenario query parameters, it restores the exact scenario set instead of showing defaults.

**Why this priority**: URL persistence is a convenience feature that builds on the core scenario functionality. It is valuable but not required for the primary comparison experience.

**Independent Test**: Can be tested by configuring scenarios, copying the URL, navigating away, pasting the URL, and verifying the same scenarios and graph appear.

**Acceptance Scenarios**:

1. **Given** a user modifies scenario parameters, **When** a parameter changes, **Then** the URL query parameters update to reflect the current scenario set.
2. **Given** a URL with scenario query parameters, **When** the page loads, **Then** the scenarios are restored exactly as encoded in the URL.
3. **Given** a URL with no scenario query parameters, **When** the page loads, **Then** the default two scenarios appear based on the child's allowance status.

---

### Edge Cases

- What happens when a child's allowance amount changes after scenarios are loaded? Scenario titles referencing the allowance amount reflect the current allowance data fetched on page load; they do not update in real-time during a session.
- What happens when all scenario inputs are zero? The scenario projects a flat line at current balance plus any interest accrual.
- What happens when a scenario's weekly spending exceeds the child's balance and allowance combined? The projection floors the balance at $0 and shows a depletion point, consistent with existing behavior.
- What happens when many scenarios are added (e.g., 5+)? A maximum of 5 scenarios is enforced to keep the graph readable.
- What happens when a scenario's one-time withdrawal exceeds the current balance? Validation prevents the withdrawal amount from exceeding the current balance, consistent with existing behavior.
- What happens if the URL query parameters contain invalid scenario data? The system falls back to the default two scenarios.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST display multiple scenario projection lines on a single graph, each with a distinct color.
- **FR-002**: The system MUST show a default of two scenarios when no URL query parameters are present.
- **FR-003**: For children with an allowance, the default first scenario MUST be weekly spending = 0 and the default second scenario MUST be weekly spending = 100% of the calculated weekly allowance amount.
- **FR-004**: For children without an allowance, the default first scenario MUST be weekly spending = 0 and the default second scenario MUST be weekly spending = $5.
- **FR-005**: The system MUST combine the "one-time deposit" and "one-time withdrawal" fields into a single "One time" field with a deposit/withdrawal toggle.
- **FR-006**: The system MUST combine the "weekly spending" field into a "Weekly" field with a spending/saving toggle.
- **FR-007**: Each scenario MUST display a dynamically generated plain-English title that updates as the user changes inputs.
- **FR-008**: Scenario titles MUST follow the title generation rules specified in the feature description (varying by allowance status, weekly spending/saving amount, and one-time deposit/withdrawal).
- **FR-009**: Scenario titles MUST appear in the graph tooltip above the date and balance.
- **FR-010**: Users MUST be able to add new scenarios beyond the default two.
- **FR-011**: Users MUST be able to delete scenarios, as long as at least one scenario remains.
- **FR-012**: The system MUST enforce a maximum of 5 scenarios.
- **FR-013**: Scenario data MUST be stored in URL query parameters so the page state is bookmarkable.
- **FR-014**: When the page loads with scenario query parameters, the system MUST restore the encoded scenarios instead of showing defaults.
- **FR-015**: When query parameters are absent or invalid, the system MUST fall back to the default two scenarios.
- **FR-016**: The existing standalone English scenario description (GrowthExplanation component) MUST be removed.
- **FR-017**: Each scenario row MUST display its associated color indicator matching the graph line color.
- **FR-018**: The time horizon selector MUST apply to all scenarios uniformly.

### Key Entities

- **Scenario**: Represents a single "what if" configuration. Attributes: weekly amount (spending or saving direction), one-time amount (deposit or withdrawal direction), associated color, generated title.
- **Scenario Set**: The collection of all active scenarios for a given projection session. Serialized to/from URL query parameters.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can visually compare at least 2 growth scenarios on a single graph within 2 seconds of page load.
- **SC-002**: Scenario titles accurately reflect input parameters for all documented title generation rules (100% coverage of rules from feature description).
- **SC-003**: Changing any scenario input updates the graph and title within 500 milliseconds.
- **SC-004**: A bookmarked URL with scenario parameters restores the exact scenario set on reload with no user intervention.
- **SC-005**: The graph remains visually readable with up to 5 simultaneous scenarios (distinct colors, clear line separation).

## Assumptions

- The time horizon control remains shared across all scenarios (not per-scenario).
- Scenario colors are drawn from a predefined palette sufficient for up to 5 scenarios.
- The allowance amount used for title generation and default scenarios is the current allowance fetched from the existing API on page load.
- Scenarios are ephemeral client-side state that persists only via URL parameters (no backend storage).
- The weekly allowance "calculated weekly amount" means the allowance amount converted to a per-week basis regardless of the underlying frequency (weekly, biweekly, monthly).
- The "Weekly" field with "Saving" selected means the user is saving additional money per week beyond their allowance (the graph projects additional deposits), distinct from the spending toggle which subtracts from the balance.

## Scope Boundaries

### In Scope
- Multi-scenario graph visualization
- Combined input fields with toggles
- Dynamic scenario titles per the specified rules
- Add/remove scenarios
- URL query parameter persistence
- Removal of GrowthExplanation component

### Out of Scope
- Backend changes or new API endpoints
- Persisting scenarios to a database
- Sharing scenarios between users
- Per-scenario time horizons
- Exporting or printing scenario comparisons
