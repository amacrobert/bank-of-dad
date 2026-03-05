# Feature Specification: Child Auto-Setup

**Feature Branch**: `026-child-auto-setup`
**Created**: 2026-03-05
**Status**: Draft
**Input**: User description: "As a parent, I want no additional setup necessary after onboarding. In the Add Child form (in both the onboarding and Children settings), add three fields next to each other for Initial Deposit, Weekly Allowance and Annual Interest. If the parent enters an initial deposit, create a deposit transaction for the child with note 'Initial deposit'. If the parent enters an allowance, automatically set up the child's allowance, scheduled weekly, with the entered amount and the note 'Weekly allowance'. If the parent enters an interest rate, automatically set up the child's interest with a monthly schedule on the first of the month."

## User Scenarios & Testing

### User Story 1 - Add Child with Full Setup During Onboarding (Priority: P1)

As a parent going through onboarding, I want to enter my child's name, password, initial deposit, weekly allowance, and annual interest rate in a single form so that everything is ready to go immediately — no extra configuration steps needed.

**Why this priority**: This is the core value proposition — eliminating post-onboarding setup friction. Most parents will encounter this form during their first experience with the app.

**Independent Test**: Can be fully tested by completing the onboarding wizard with all three optional fields filled in, then verifying the child's account shows the correct balance, an active weekly allowance schedule, and an active monthly interest schedule.

**Acceptance Scenarios**:

1. **Given** a parent is on the Add Child step of onboarding, **When** they enter a child name, password, initial deposit of $50, weekly allowance of $10, and annual interest of 5%, **Then** the child account is created with a $50 balance, an active weekly allowance of $10, and a monthly interest schedule at 5% on the 1st of each month.
2. **Given** a parent is on the Add Child step of onboarding, **When** they enter only a child name and password (leaving all three optional fields empty), **Then** the child account is created with a $0 balance and no allowance or interest schedules — same behavior as today.
3. **Given** a parent is on the Add Child step of onboarding, **When** they enter a child name, password, and only a weekly allowance of $5, **Then** the child account is created with a $0 balance, an active weekly allowance of $5, and no interest schedule.

---

### User Story 2 - Add Child with Full Setup from Children Settings (Priority: P1)

As a parent managing an existing family, I want to add a new child from the Children settings with the same streamlined form so that the new child's account is immediately configured.

**Why this priority**: Equal priority with onboarding — the same form component is used in both contexts, and parents adding children after initial setup should have the same seamless experience.

**Independent Test**: Can be fully tested by opening Children settings, adding a new child with all optional fields filled, and verifying the resulting account state.

**Acceptance Scenarios**:

1. **Given** a parent is in Children settings, **When** they add a child with name, password, $25 initial deposit, $15 weekly allowance, and 3% annual interest, **Then** the child is created with a $25 balance, active weekly allowance of $15, and monthly interest at 3%.
2. **Given** a parent is in Children settings, **When** they add a child with only a name and password, **Then** the child is created with no deposit, no allowance, and no interest — identical to current behavior.

---

### Edge Cases

- What happens when the parent enters $0 for initial deposit? The deposit is skipped (no $0 transaction created).
- What happens when the parent enters 0% for annual interest? The interest schedule is skipped (no 0% schedule created).
- What happens when the parent enters $0 for weekly allowance? The allowance schedule is skipped (no $0 allowance created).
- What happens if child creation succeeds but one of the subsequent operations (deposit, allowance, or interest) fails? The child account is still created. Any operations that succeeded remain in place. The parent sees an error indicating which setup step failed, and they can configure the failed item manually from the child's settings.
- What happens if the parent enters a very large initial deposit (e.g., $999,999.99)? It should be accepted — the same validation limits as the existing deposit form apply.
- What happens if the parent enters a very high interest rate (e.g., 100%)? It should be accepted — the same validation limits as the existing interest form apply (0-100%).

## Requirements

### Functional Requirements

- **FR-001**: The Add Child form MUST display three optional fields — Initial Deposit, Weekly Allowance, and Annual Interest — arranged side by side in a single row.
- **FR-002**: All three new fields MUST be optional. Leaving any field empty or at zero skips the corresponding setup action.
- **FR-003**: When a non-zero Initial Deposit is provided, the system MUST create a deposit transaction for the newly created child with the note "Initial deposit".
- **FR-004**: When a non-zero Weekly Allowance is provided, the system MUST create an active weekly allowance schedule for the child with the entered amount and the note "Weekly allowance". The day of week defaults to the current day of week.
- **FR-005**: When a non-zero Annual Interest rate is provided, the system MUST create an active monthly interest schedule for the child, running on the 1st of each month.
- **FR-006**: The form MUST use the same validation rules as the existing standalone forms (deposit: $0.01–$999,999.99; allowance: $0.01–$999,999.99; interest: 0.01%–100%).
- **FR-007**: The form MUST behave identically in both the onboarding wizard and the Children settings modal.
- **FR-008**: If the child is created successfully but a subsequent setup operation fails, the system MUST still retain the created child and any successfully completed setup operations, and display an error message indicating what failed.
- **FR-009**: The success confirmation after child creation MUST reflect what was set up (e.g., showing the initial balance, confirming allowance and interest are active).

### Key Entities

- **Child**: Existing entity — gains initial balance, allowance, and interest at creation time rather than requiring separate configuration.
- **Transaction**: Existing entity — an "Initial deposit" transaction is created when initial deposit is specified.
- **Allowance Schedule**: Existing entity — a weekly schedule is created when allowance amount is specified.
- **Interest Schedule**: Existing entity — a monthly schedule on the 1st is created when interest rate is specified.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Parents can fully configure a child's account (deposit, allowance, interest) in a single form submission, reducing required steps from 4 separate forms to 1.
- **SC-002**: 100% of existing Add Child functionality continues to work unchanged when the new optional fields are left empty.
- **SC-003**: All three optional setup actions (deposit, allowance, interest) produce identical results to configuring each one individually through the existing dedicated forms.
- **SC-004**: The form layout remains clean and usable on mobile screens (fields stack vertically on narrow viewports).

## Assumptions

- The weekly allowance day defaults to the current day of the week at time of creation, since the form doesn't expose a day picker (keeping it simple).
- Interest is always monthly on the 1st — no frequency or day picker is shown, matching the user's specification.
- The three setup operations (deposit, allowance, interest) are performed as separate API calls after child creation, not as a single atomic backend operation. This keeps the implementation simple and reuses existing endpoints.
- The Initial Deposit, Weekly Allowance, and Annual Interest fields use currency/percentage input formatting consistent with the existing deposit and interest forms.
