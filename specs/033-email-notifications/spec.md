# Feature Specification: Email Notifications

**Feature Branch**: `033-email-notifications`  
**Created**: 2026-04-02  
**Status**: Draft  
**Input**: User description: "Email Notifications. Parents currently have no way to know when a child completes a chore or requests a withdrawal without opening the app. The approval workflows (chores + withdrawals) create a clear need for push communication. Brevo email integration already exists for the contact form — the infrastructure is partially in place."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parent Receives Email When Child Requests a Withdrawal (Priority: P1)

A parent receives an email notification when one of their children submits a withdrawal request. The email includes the child's name, the requested amount, and the reason provided. The parent can see at a glance what needs their attention without opening the app.

**Why this priority**: Withdrawal requests involve real money leaving a child's account and require explicit parent approval. This is the highest-stakes approval workflow — delays in responding can frustrate children and undermine the app's value as a financial teaching tool.

**Independent Test**: Can be fully tested by having a child submit a withdrawal request and verifying that the parent receives an email with the correct details within a reasonable timeframe.

**Acceptance Scenarios**:

1. **Given** a child submits a withdrawal request for $15.00 with reason "Birthday present for Mom", **When** the request is created, **Then** all parents in the family receive an email with the child's name, amount ($15.00), and reason.
2. **Given** a family has two parents registered, **When** a child submits a withdrawal request, **Then** both parents receive the notification email.
3. **Given** a parent has disabled email notifications for withdrawal requests, **When** a child submits a withdrawal request, **Then** that parent does not receive an email (but other opted-in parents still do).

---

### User Story 2 - Parent Receives Email When Child Completes a Chore (Priority: P1)

A parent receives an email notification when one of their children marks a chore as complete and it moves to pending approval. The email includes the child's name, the chore title, and the reward amount.

**Why this priority**: Chore completion is the other core approval workflow. Without notifications, completed chores sit in a queue unseen, which discourages children from doing chores when they don't get timely recognition.

**Independent Test**: Can be fully tested by having a child mark a chore as complete and verifying the parent receives an email with the chore details.

**Acceptance Scenarios**:

1. **Given** a child completes a chore titled "Clean Room" worth $5.00, **When** the chore status changes to pending approval, **Then** all opted-in parents in the family receive an email with the child's name, chore title, and reward amount.
2. **Given** a child completes multiple chores in quick succession (within 5 minutes), **When** the notifications are processed, **Then** the parent receives a single summary email listing all completed chores rather than individual emails for each.

---

### User Story 3 - Parent Manages Notification Preferences (Priority: P2)

A parent can control which types of email notifications they receive. They can enable or disable notifications for withdrawal requests and chore completions independently. Notifications are enabled by default for new accounts.

**Why this priority**: Not all parents want to receive every notification. Giving control prevents email fatigue and ensures parents stay opted in for the notifications that matter to them.

**Independent Test**: Can be fully tested by toggling notification preferences in settings and verifying that only the enabled notification types result in emails.

**Acceptance Scenarios**:

1. **Given** a newly registered parent, **When** their account is created, **Then** all notification types are enabled by default.
2. **Given** a parent disables chore completion notifications, **When** a child completes a chore, **Then** that parent receives no email, but still receives withdrawal request emails.
3. **Given** a parent disables all notifications, **When** any notifiable event occurs, **Then** that parent receives no emails.
4. **Given** a parent re-enables a previously disabled notification type, **When** the next relevant event occurs, **Then** they receive the email.

---

### User Story 4 - Parent Receives Email When Another Parent Makes a Decision (Priority: P3)

When a parent approves or denies a chore or withdrawal request, the other parents in the family receive a notification that the action was taken. This keeps co-parents informed without requiring them to check the app.

**Why this priority**: In multi-parent families, one parent may approve or deny a request. The other parent benefits from knowing the outcome, but this is informational rather than action-required, making it lower priority.

**Independent Test**: Can be fully tested in a two-parent family by having one parent approve a request and verifying the other parent receives an email.

**Acceptance Scenarios**:

1. **Given** a family with two parents and Parent A approves a withdrawal request, **When** the approval is recorded, **Then** Parent B receives an email indicating the request was approved, including the child's name and amount.
2. **Given** a family with only one parent, **When** that parent approves or denies a request, **Then** no notification email is sent (no other parent to notify).
3. **Given** a parent has disabled decision notifications, **When** the other parent makes a decision, **Then** the opted-out parent does not receive an email.

---

### Edge Cases

- What happens when the email delivery service is unavailable? The approval action (chore approval, withdrawal approval) must still succeed — email delivery failures must never block core functionality.
- What happens when a parent's email address becomes invalid (e.g., Google account deactivated)? The system should gracefully handle bounced emails without affecting the parent's ability to use the app.
- What happens when a child completes many chores at once? Notifications should be batched within a short window (5 minutes) to avoid email spam.
- What happens when a parent opts out of all notifications and later opts back in? They should only receive notifications for events occurring after re-enabling, not a backlog.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST send an email to all opted-in parents in a family when a child submits a withdrawal request, including the child's name, requested amount, and reason.
- **FR-002**: System MUST send an email to all opted-in parents in a family when a child marks a chore as complete, including the child's name, chore title, and reward amount.
- **FR-003**: System MUST batch chore completion notifications that occur within a 5-minute window into a single summary email per parent.
- **FR-004**: System MUST allow each parent to independently enable or disable notifications for each notification type (withdrawal requests, chore completions, approval decisions).
- **FR-005**: System MUST enable all notification types by default for new parent accounts.
- **FR-006**: System MUST send decision notification emails to other parents in the family (not the parent who took the action) when a chore or withdrawal request is approved or denied.
- **FR-007**: System MUST NOT block or delay any core action (chore completion, approval, withdrawal request) if email delivery fails.
- **FR-008**: System MUST include the family's custom bank name in notification emails when one is configured.
- **FR-009**: System MUST NOT send notifications for events that occurred while a parent had that notification type disabled, even if they later re-enable it.
- **FR-010**: System MUST provide a one-click unsubscribe mechanism in every notification email that disables all notifications for that parent.

### Key Entities

- **Notification Preference**: Represents a parent's opt-in/opt-out choice for each notification type. Associated with a parent. Includes a preference per notification category (withdrawal requests, chore completions, approval decisions).
- **Notification Event**: A triggering action (chore completed, withdrawal requested, decision made) that may result in one or more emails being sent. Includes the event type, relevant child, relevant details (amount, chore name, reason), and timestamp.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Parents receive notification emails within 2 minutes of a triggering event (or within 7 minutes for batched chore completions).
- **SC-002**: Email delivery failures do not increase the response time of any approval or request action by more than 500 milliseconds.
- **SC-003**: Parents can update their notification preferences in under 30 seconds from the settings page.
- **SC-004**: 100% of notification emails include a working one-click unsubscribe option.

## Assumptions

- Parents authenticate via Google OAuth, so their email address is always available and valid at the time of account creation.
- Children do not have email addresses in the system. Child-facing notifications (e.g., "your chore was approved") are out of scope for this feature and would be addressed by a future in-app notification system.
- The existing email delivery service (Brevo) is suitable for transactional notification emails and does not require migration to a different provider.
- Email content will be plain text initially. Branded HTML email templates are a future enhancement, not part of this feature.
- The system does not need to track email delivery status (opened, bounced, etc.) — this is a future analytics enhancement.
