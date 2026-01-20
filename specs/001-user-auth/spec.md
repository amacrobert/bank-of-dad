# Feature Specification: User Authentication

**Feature Branch**: `001-user-auth`
**Created**: 2026-01-20
**Status**: Draft
**Input**: User authentication for Bank of Dad application

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parent Registration via Google (Priority: P1)

A parent discovers Bank of Dad and wants to create an account using their Google account. During registration, they set up their family's unique bank URL that their children will use to access their accounts.

**Why this priority**: Registration is the entry point for all users. Without the ability to create an account, no other features can be used. This is the foundation of the entire application.

**Independent Test**: Can be fully tested by completing the Google sign-in flow and verifying the parent can access the application with their family bank URL created. Delivers immediate value by enabling family onboarding.

**Acceptance Scenarios**:

1. **Given** a visitor is on the registration page, **When** they click "Sign in with Google" and authorize the application, **Then** their account is created, a family is created, and they are prompted to choose a unique bank URL slug for their family.
2. **Given** a visitor has authorized via Google, **When** they enter a bank URL slug that is already taken, **Then** they see an error message and suggestions for available alternatives.
3. **Given** a visitor has authorized via Google, **When** they enter a valid unique bank URL slug, **Then** their family bank is created at that URL (e.g., `bankofda.de/smith-family`) and they are redirected to their parent dashboard.
4. **Given** a visitor attempts to register, **When** they have previously registered with the same Google account, **Then** they are logged in to their existing account instead of creating a duplicate.

---

### User Story 2 - Parent Login via Google (Priority: P1)

A registered parent returns to Bank of Dad and needs to log in using their Google account.

**Why this priority**: Login is essential for returning users. Combined with registration, these two stories enable the complete access cycle and are required before any other functionality.

**Independent Test**: Can be fully tested by logging in with Google and verifying access to the parent dashboard. Delivers immediate value by enabling returning user access.

**Acceptance Scenarios**:

1. **Given** a registered parent is on the login page, **When** they click "Sign in with Google" and authenticate, **Then** they are redirected to their family dashboard.
2. **Given** a visitor on the login page, **When** they sign in with a Google account that has no Bank of Dad account, **Then** they are redirected to complete registration (choose bank URL slug).
3. **Given** a logged-in parent, **When** they click "Log out", **Then** their session is ended and they are redirected to the home page.

---

### User Story 3 - Parent Creates Child Account (Priority: P1)

A parent wants to add their child to the family bank so the child can log in and see their account balance and learn about saving.

**Why this priority**: Creating child accounts is the core value proposition of the app. Without children, there's no one to teach about compound interest. This is essential to deliver the educational purpose.

**Independent Test**: Can be fully tested by a parent creating a child account and verifying the child appears in the family. Delivers value by enabling the primary educational use case.

**Acceptance Scenarios**:

1. **Given** a logged-in parent on their dashboard, **When** they click "Add Child" and enter a first name and password, **Then** a child account is created in their family.
2. **Given** a parent creating a child account, **When** they enter a password that does not meet requirements, **Then** they see specific feedback about what the password needs.
3. **Given** a parent has created a child account, **When** creation is complete, **Then** they see the child's login URL and credentials so they can share them with the child.

---

### User Story 4 - Child Login via Family Bank URL (Priority: P1)

A child wants to view their bank account. They navigate to their family's unique bank URL and log in with their first name and password.

**Why this priority**: Child login is essential for the educational purpose. If children can't access their accounts independently, they can't learn by engaging with their balance and interest.

**Independent Test**: Can be fully tested by a child navigating to their family URL, logging in, and viewing their account. Delivers the core educational value of the application.

**Acceptance Scenarios**:

1. **Given** a child navigates to their family's bank URL (e.g., `bankofda.de/smith-family`), **When** they enter their first name and correct password, **Then** they are logged in and see their child dashboard with their account balance.
2. **Given** a child on their family's bank URL login page, **When** they enter incorrect credentials, **Then** they see a friendly error message encouraging them to try again or ask their parent for help.
3. **Given** a child attempts to log in, **When** they fail 5 times consecutively, **Then** their account is temporarily locked and their parent is notified.
4. **Given** a child is logged in, **When** they click "Log out", **Then** their session ends and they return to the family bank login page.

---

### User Story 5 - Parent Manages Child Credentials (Priority: P2)

A parent needs to reset their child's password because the child forgot it, or update the child's display name.

**Why this priority**: Credential management improves the experience but children can work around forgotten passwords by asking parents to check the original credentials or reset them. Not blocking for initial launch.

**Independent Test**: Can be fully tested by a parent resetting a child's password and the child logging in with the new password. Delivers value by enabling self-service account management.

**Acceptance Scenarios**:

1. **Given** a logged-in parent viewing their family, **When** they select a child and click "Reset Password", **Then** they can enter a new password for that child.
2. **Given** a parent has reset their child's password, **When** the child next attempts to log in with the old password, **Then** login fails and they must use the new password.
3. **Given** a logged-in parent, **When** they update a child's display name, **Then** the new name is shown throughout the application.

---

### User Story 6 - Session Persistence (Priority: P2)

A logged-in user (parent or child) closes their browser and returns later, expecting to still be logged in.

**Why this priority**: Session persistence improves user experience significantly but is not strictly required. The app works without it (users just log in each time).

**Independent Test**: Can be fully tested by logging in, closing the browser, and verifying the user remains authenticated on return. Delivers value by reducing login friction.

**Acceptance Scenarios**:

1. **Given** a logged-in parent closes their browser, **When** they return within 7 days, **Then** they remain logged in without needing to re-authenticate.
2. **Given** a logged-in child closes their browser, **When** they return within 24 hours, **Then** they remain logged in.
3. **Given** a logged-in user whose session has expired, **When** they attempt to access protected content, **Then** they are redirected to the appropriate login page.

---

### Edge Cases

- What happens when Google's authentication service is temporarily unavailable? System shows a friendly error message asking users to try again shortly.
- What happens if a child's first name matches another child in the same family? System requires unique first names within a family; parent is prompted to add a distinguishing identifier (e.g., "Tommy J" vs "Tommy M").
- What happens if someone navigates to a non-existent family bank URL? System shows a 404 page with a message that this bank doesn't exist and a link to create your own.
- What happens if a parent deletes their Google account after registering? They cannot log in; they would need to contact support for account recovery options.
- How does the system handle a child trying to access a different family's URL? The child's credentials won't work on another family's login page; they only work for their own family.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow parents to register and log in exclusively via Google OAuth
- **FR-002**: System MUST require new parents to choose a unique URL slug for their family bank during registration
- **FR-003**: System MUST validate that family bank URL slugs contain only lowercase letters, numbers, and hyphens, and are between 3-30 characters
- **FR-004**: System MUST prevent duplicate family bank URL slugs
- **FR-005**: System MUST allow parents to create child accounts with a first name and password
- **FR-006**: System MUST enforce child password requirements: minimum 6 characters (simple enough for children to remember)
- **FR-007**: System MUST require unique first names for children within the same family
- **FR-008**: System MUST securely hash child passwords before storage
- **FR-009**: System MUST authenticate children via their family's unique URL using first name and password
- **FR-010**: System MUST implement account lockout for children after 5 failed login attempts
- **FR-011**: System MUST notify parents when their child's account is locked due to failed login attempts
- **FR-012**: System MUST maintain parent sessions for up to 7 days of inactivity
- **FR-013**: System MUST maintain child sessions for up to 24 hours of inactivity
- **FR-014**: System MUST allow users to explicitly log out, invalidating their session
- **FR-015**: System MUST allow parents to reset their children's passwords
- **FR-016**: System MUST allow parents to update their children's display names
- **FR-017**: System MUST restrict child accounts to viewing only their own account data
- **FR-018**: System MUST log all authentication events (login success, login failure, logout, account creation) without exposing sensitive data

### Key Entities

- **Parent**: A user who registers via Google OAuth and manages a family. Has Google ID, email, display name (from Google), and owns exactly one Family.
- **Child**: A user created by a parent within a family. Has first name, hashed password, and belongs to exactly one Family. Can only access their own account data.
- **Family**: An organizational unit created when a parent registers. Has a unique URL slug (e.g., "smith-family") and contains one parent and zero or more children.
- **Session**: Represents an active authentication state for a user. Has expiration time (7 days for parents, 24 hours for children) and can be invalidated.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Parents can complete registration (Google auth + URL slug selection) in under 2 minutes
- **SC-002**: Parents can create a child account in under 1 minute
- **SC-003**: Children can log in within 30 seconds of entering credentials
- **SC-004**: System supports at least 100 concurrent authenticated users without degradation
- **SC-005**: Zero plaintext passwords stored in the system (verified by security audit)
- **SC-006**: Children aged 6+ can successfully log in on first attempt 90% of the time
- **SC-007**: Failed login attempts are blocked within 1 second of the 5th failure
- **SC-008**: 100% of authentication events are logged for security auditing

## Assumptions

- Google OAuth is the only authentication method for parents (no email/password option for parents)
- Each parent has exactly one family (no support for parents belonging to multiple families initially)
- Children do not have email addresses in the system
- The family bank URL format will be `[domain]/[slug]` where slug is chosen by the parent
- Mobile apps may come later; initial implementation targets web browsers
- Parents are responsible for sharing login credentials with their children (no automated notification to children)
