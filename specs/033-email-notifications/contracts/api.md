# API Contracts: Email Notifications

**Feature Branch**: `033-email-notifications`  
**Date**: 2026-04-02

## Notification Preferences

### GET /api/settings/notifications

**Auth**: Parent only  
**Description**: Get current parent's notification preferences.

**Response 200**:
```json
{
  "notify_withdrawal_requests": true,
  "notify_chore_completions": true,
  "notify_decisions": false
}
```

---

### PUT /api/settings/notifications

**Auth**: Parent only  
**Description**: Update current parent's notification preferences. All fields optional — only provided fields are updated.

**Request**:
```json
{
  "notify_withdrawal_requests": true,
  "notify_chore_completions": false,
  "notify_decisions": true
}
```

**Response 200**:
```json
{
  "message": "Notification preferences updated",
  "notify_withdrawal_requests": true,
  "notify_chore_completions": false,
  "notify_decisions": true
}
```

**Response 400** (invalid field value):
```json
{
  "error": "invalid_request",
  "message": "notification preferences must be boolean values"
}
```

---

### GET /api/notifications/unsubscribe?token={token}

**Auth**: None (token-authenticated)  
**Description**: One-click unsubscribe from all email notifications. Token is HMAC-signed and contains parent ID.

**Response 200** (HTML):
```
Plain text or minimal HTML page confirming unsubscribe:
"You have been unsubscribed from all Bank of Dad email notifications.
You can re-enable notifications in your settings at any time."
```

**Response 400** (invalid/expired token):
```
"Invalid or expired unsubscribe link."
```

## Extended Existing Endpoints

### GET /api/settings (modified)

**Change**: Response now includes notification preferences alongside existing timezone and bank_name fields.

**Response 200**:
```json
{
  "timezone": "America/New_York",
  "bank_name": "Dad",
  "notifications": {
    "notify_withdrawal_requests": true,
    "notify_chore_completions": true,
    "notify_decisions": true
  }
}
```

## Email Templates (Plain Text)

### Withdrawal Request Notification

**Subject**: `{child_name} requested a withdrawal from {bank_name}`

**Body**:
```
Hi {parent_name},

{child_name} has requested a withdrawal of {amount} from their {bank_name} account.

Reason: {reason}

Log in to approve or deny this request.

—
{bank_name}
To unsubscribe from these notifications: {unsubscribe_url}
```

### Chore Completion Notification (Single)

**Subject**: `{child_name} completed a chore in {bank_name}`

**Body**:
```
Hi {parent_name},

{child_name} completed "{chore_name}" (reward: {reward_amount}) and is waiting for your approval.

Log in to review.

—
{bank_name}
To unsubscribe from these notifications: {unsubscribe_url}
```

### Chore Completion Notification (Batched)

**Subject**: `{child_name} completed {count} chores in {bank_name}`

**Body**:
```
Hi {parent_name},

{child_name} completed the following chores and is waiting for your approval:

- {chore_name_1} (reward: {amount_1})
- {chore_name_2} (reward: {amount_2})
- {chore_name_3} (reward: {amount_3})

Total pending reward: {total_amount}

Log in to review.

—
{bank_name}
To unsubscribe from these notifications: {unsubscribe_url}
```

### Decision Notification (Approval/Denial)

**Subject**: `{action_parent_name} {action} {child_name}'s {request_type} in {bank_name}`

**Body**:
```
Hi {parent_name},

{action_parent_name} {approved/denied} {child_name}'s {withdrawal request for {amount} / chore "{chore_name}"}.

{If denied: Reason: {denial_reason}}

—
{bank_name}
To unsubscribe from these notifications: {unsubscribe_url}
```
