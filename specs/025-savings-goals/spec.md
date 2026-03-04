# Feature Specification: Savings Goals

**Feature Branch**: `025-savings-goals`
**Created**: 2026-03-02
**Status**: Draft
**Input**: User description: "As a child, I want to set up and manage savings goals and be able to progress to achieving them. This feature should be very visually appealing to make saving feel fun and rewarding."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create a Savings Goal (Priority: P1)

A child wants to save up for something specific (e.g., a new toy, a video game, a bike). They open their dashboard and tap a button to create a new savings goal. They enter a name for the goal (e.g., "New Skateboard"), set the target amount (e.g., $45.00), optionally choose an emoji to represent the goal, and optionally set a target date. The goal appears on their dashboard with a visual progress indicator showing how close they are to reaching it.

**Why this priority**: This is the foundational feature — without the ability to create goals, no other savings goal functionality can exist. It delivers immediate value by letting children visualize what they're saving for.

**Independent Test**: Can be fully tested by having a child log in, create a goal with a name and target amount, and verifying it appears on their dashboard with a progress indicator showing 0% progress.

**Acceptance Scenarios**:

1. **Given** a child is logged in and on their dashboard, **When** they tap "Add Goal" and enter a goal name and target amount, **Then** the goal is created and displayed on their dashboard with a progress indicator at 0%.
2. **Given** a child is creating a goal, **When** they enter a goal name, target amount, choose an emoji, and set a target date, **Then** all details are saved and displayed on the goal card.
3. **Given** a child is creating a goal, **When** they submit without entering a goal name or target amount, **Then** the system shows a validation message indicating these fields are required.
4. **Given** a child already has 5 active goals, **When** they try to create another goal, **Then** the system informs them they've reached the maximum number of active goals.

---

### User Story 2 - Allocate Money Toward a Goal (Priority: P1)

A child views their savings goals and decides to put money toward one of them. They tap on a goal and choose to allocate a specific amount from their available balance. The allocated amount is reserved — it reduces the child's "available balance" and increases the goal's saved amount. The child's dashboard shows their total balance (available + saved) as the primary, most visually prominent number, with a secondary breakdown showing available balance and total saved toward goals. The progress indicator updates immediately with a satisfying visual animation, showing them getting closer to their target.

**Why this priority**: Allocation is the core interaction that makes savings goals meaningful. Without it, goals are just static labels. This enables the progress tracking that makes saving feel rewarding.

**Independent Test**: Can be fully tested by creating a goal, allocating money toward it, and verifying the progress indicator updates accordingly.

**Acceptance Scenarios**:

1. **Given** a child has an active goal and a positive available balance, **When** they allocate $10.00 toward their goal, **Then** the goal's progress updates to reflect the new amount saved and a progress animation plays.
2. **Given** a child has $20.00 available and tries to allocate $25.00 to a goal, **When** they confirm the allocation, **Then** the system shows an error indicating insufficient available funds.
3. **Given** a child allocates money to a goal, **When** the allocation succeeds, **Then** a record of the allocation is visible in the goal's detail view.

---

### User Story 3 - View Goal Progress with Visual Feedback (Priority: P1)

A child opens their dashboard and sees all their active savings goals with rich, visually engaging progress indicators. Each goal shows a progress bar or ring that fills up as they get closer to their target. The current amount saved and remaining amount are clearly displayed. Goals that are close to completion (e.g., over 75%) show encouraging visual cues. Goals use the child's selected theme colors to feel personalized.

**Why this priority**: Visual engagement is explicitly called out in the feature request. The appeal and fun of the progress visualization is what makes children want to save. This is critical for user engagement and educational value.

**Independent Test**: Can be fully tested by creating goals at various progress levels and verifying the visual indicators display correctly at 0%, 25%, 50%, 75%, and 100%.

**Acceptance Scenarios**:

1. **Given** a child has multiple active goals at different progress levels, **When** they view their dashboard, **Then** each goal displays a visually distinct progress indicator reflecting its completion percentage.
2. **Given** a goal is over 75% complete, **When** the child views it, **Then** the goal card shows an encouraging visual cue (e.g., sparkle effect, celebratory color).
3. **Given** a child has the "Arctic" theme selected, **When** they view their goals, **Then** the goal progress indicators and cards use the Arctic theme's color palette.

---

### User Story 4 - Achieve a Goal (Priority: P2)

When a child's savings toward a goal reach or exceed the target amount, the goal is marked as achieved. A celebratory visual experience plays (e.g., confetti animation, congratulations message) to make the accomplishment feel exciting and rewarding. The achieved goal moves to a "completed" section so the child can look back on their accomplishments.

**Why this priority**: Goal achievement is the payoff for the saving effort. The celebration moment is what creates positive reinforcement and makes saving feel worth it. It depends on P1 stories being complete.

**Independent Test**: Can be fully tested by creating a goal with a small target amount, allocating enough funds to meet it, and verifying the celebration animation plays and the goal moves to a completed section.

**Acceptance Scenarios**:

1. **Given** a goal needs $5.00 more to reach its target, **When** the child allocates $5.00 or more, **Then** a celebration animation plays and the goal is marked as achieved.
2. **Given** a goal has been achieved, **When** the child views their goals, **Then** the achieved goal appears in a "Completed Goals" section with a visual badge indicating achievement.
3. **Given** a goal has been achieved, **When** the child views the completed goal, **Then** they can see the total amount saved and the date it was achieved.

---

### User Story 5 - Manage Goals (Priority: P2)

A child can edit or delete their savings goals. They can update the goal name, target amount, emoji, or target date. They can also delete a goal they no longer want to pursue. When a goal is deleted, any funds allocated to it become available again.

**Why this priority**: Goal management is important for flexibility — children's interests change and they need the ability to adjust their goals. This depends on goals existing (P1).

**Independent Test**: Can be fully tested by creating a goal, editing its details, verifying the changes persist, then deleting it and verifying it is removed and funds are released.

**Acceptance Scenarios**:

1. **Given** a child has an active goal, **When** they edit the goal name and target amount, **Then** the updated details are saved and reflected on the dashboard.
2. **Given** a child has an active goal with $15.00 allocated, **When** they delete the goal, **Then** the goal is removed and the $15.00 becomes available in their balance again.
3. **Given** a child has an active goal, **When** they increase the target amount, **Then** the progress indicator recalculates to show the new percentage.
4. **Given** a child has a goal with $30.00 saved, **When** they withdraw $10.00 from the goal, **Then** $10.00 returns to their available balance and the goal's progress updates to reflect $20.00 saved.

---

### User Story 6 - Parent Views Child's Goals (Priority: P3)

A parent can view their child's savings goals from the parent dashboard. They can see what each child is saving for, how much progress they've made, and whether any goals have been achieved. This gives parents visibility into their child's financial behavior and goal-setting habits.

**Why this priority**: Parents need oversight but don't need to manage goals directly. This is a read-only view that adds value without adding complexity to the child's workflow.

**Independent Test**: Can be fully tested by having a child create goals, then logging in as the parent and verifying the child's goals are visible from the parent dashboard.

**Acceptance Scenarios**:

1. **Given** a parent is viewing a child's account details, **When** the child has active savings goals, **Then** the parent sees a list of the child's goals with names, target amounts, and progress.
2. **Given** a parent is viewing a child's account, **When** the child has completed goals, **Then** the parent can see the child's achievement history.

---

### Edge Cases

- What happens when a parent makes a withdrawal that brings the child's balance below their total goal allocations? The system warns the parent before proceeding. If the parent confirms, goal allocations are reduced proportionally to fit the new balance.
- What happens when a child tries to allocate more than their available balance across all goals combined? The system should prevent over-allocation and show the available amount.
- What happens when a goal's target date passes without the goal being achieved? The goal should remain active but display a visual indicator that the target date has passed (no automatic deletion).
- What happens if a child edits a goal's target amount to be less than the amount already allocated? The goal should be marked as achieved and trigger the celebration experience.
- What happens when interest or allowance is deposited? It increases the total balance but does not automatically allocate to any goal — the child decides where to put their money.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Children MUST be able to create savings goals with a required name (max 50 characters) and required target amount (minimum $0.01).
- **FR-002**: Children MUST be able to optionally assign an emoji icon and a target date to each goal.
- **FR-003**: Children MUST be limited to a maximum of 5 active (non-completed) savings goals at any time.
- **FR-004**: Children MUST be able to allocate money from their available balance toward a specific goal. Allocated funds are reserved — they reduce the available balance and are committed to the goal.
- **FR-005**: The system MUST prevent a child from allocating more than their available balance (total balance minus all existing goal allocations) to goals.
- **FR-005a**: The child's dashboard MUST display the total balance (available + saved) as the most visually prominent number, with a secondary breakdown of available balance and total saved toward goals.
- **FR-005b**: Interest MUST be calculated on the child's total balance (available + saved toward goals), not just the available balance. Saving toward goals does not reduce interest earnings.
- **FR-006**: The system MUST display a visual progress indicator for each goal showing the percentage of the target amount that has been saved.
- **FR-007**: Goal progress indicators MUST use the child's selected visual theme colors.
- **FR-008**: The system MUST display a celebration animation when a goal's allocated amount reaches or exceeds the target amount.
- **FR-009**: Completed goals MUST be moved to a separate "Completed Goals" section and display the achievement date. The section displays the 5 most recently completed goals by default, with a "View all" option to see the full history.
- **FR-010**: Children MUST be able to edit the name, target amount, emoji, and target date of their active goals.
- **FR-011**: Children MUST be able to delete active goals, with any allocated funds returned to their available balance.
- **FR-011a**: Children MUST be able to withdraw any amount from a goal's saved funds back to their available balance without deleting the goal. The goal's progress adjusts accordingly.
- **FR-012**: Parents MUST be able to view their children's savings goals and progress from the parent dashboard (read-only).
- **FR-013**: The system MUST record each goal allocation as a trackable event with the amount and date.
- **FR-014**: Goal cards MUST show the goal name, emoji (if set), amount saved, target amount, and progress percentage.
- **FR-015**: Goals nearing completion (75%+ progress) MUST display an encouraging visual cue to motivate the child.
- **FR-016**: When a parent withdrawal would bring the child's total balance below their total goal allocations, the system MUST warn the parent that the withdrawal will affect the child's savings goals. The parent can proceed, and goal allocations are reduced proportionally to fit the new available balance.

### Key Entities

- **Savings Goal**: Represents a child's savings target. Belongs to a child. Attributes: name, target amount, current allocated amount, emoji icon (optional), target date (optional), status (active/completed), creation date, completion date (when achieved).
- **Goal Allocation**: Represents a transfer of funds from the child's available balance to a specific goal. Attributes: goal reference, amount, date of allocation.

## Assumptions

- Children can create and manage goals independently without parent approval. The feature prioritizes child autonomy and learning.
- The maximum of 5 active goals is chosen to keep the experience focused and age-appropriate — children shouldn't feel overwhelmed by too many goals.
- The minimum goal amount of $0.01 allows for any size goal; there is no maximum.
- Completed goals are retained for historical viewing but do not count toward the 5-goal limit.
- Interest and allowance deposits increase the total balance but do not auto-allocate to any goal. Children choose where to direct their savings.
- Interest is calculated on the full total balance (available + saved toward goals). Reserving funds for goals does not penalize interest earnings.
- The celebration animation on goal completion is a brief, non-blocking visual effect (confetti or similar) that plays once.
- Goal progress visualization integrates with the existing child theme system, using theme colors for consistency.
- Parents have read-only access to goals — they cannot create, edit, or delete goals on behalf of children.

## Clarifications

### Session 2026-03-02

- Q: Should allocated funds be reserved from the available balance or should goals be purely visual trackers? → A: Reserved funds. Allocating to a goal reduces "available balance." Interest is still calculated on total balance (available + saved). Total balance remains the most visually prominent number on the balance card.
- Q: When a parent withdrawal would drop the balance below total goal allocations, what should happen? → A: Warn the parent and allow them to proceed. Goal allocations reduce proportionally.
- Q: Can a child remove money from a goal without deleting it? → A: Yes, allow partial de-allocation. Child can withdraw any amount from a goal back to available balance.
- Q: How many completed goals are shown in the completed section? → A: Show the 5 most recent, with a "View all" option to see full history.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Children can create a new savings goal in under 30 seconds.
- **SC-002**: Children can allocate funds to a goal in under 10 seconds (2 taps/clicks maximum).
- **SC-003**: Goal progress updates are reflected visually within 1 second of allocation.
- **SC-004**: 100% of goal completions trigger the celebration animation experience.
- **SC-005**: Children can view all active goals and their progress at a glance from the main dashboard.
- **SC-006**: Parents can view any child's goals within 2 taps/clicks from the parent dashboard.
- **SC-007**: Goal management (edit/delete) is completable in under 20 seconds.
- **SC-008**: Visual progress indicators are legible and appealing across all 12 available child themes.
