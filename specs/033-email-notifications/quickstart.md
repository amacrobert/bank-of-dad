# Quickstart: Email Notifications

**Feature Branch**: `033-email-notifications`  
**Date**: 2026-04-02

## Overview

Add email notifications to Bank of Dad so parents are alerted when children complete chores or request withdrawals, and when co-parents make approval decisions. Notifications are sent via the existing Brevo transactional email service. Parents can manage their notification preferences in settings and one-click unsubscribe from any email.

## Architecture

```
Child action (chore complete / withdrawal request)
  ↓
Handler completes core action (DB write, response sent)
  ↓
Fire-and-forget goroutine
  ↓
Notification service
  ├── Look up parents in family
  ├── Filter by notification preferences
  ├── For chore completions: buffer in batcher (5-min window)
  └── Send via Brevo TransactionalEmailsApi
```

## Key Design Decisions

1. **Fire-and-forget**: Emails sent in goroutines — never block handler responses (FR-007)
2. **Preferences on parents table**: 3 boolean columns, not a separate table (Simplicity principle)
3. **In-memory batcher**: Chore completions batched per-family with 5-min flush timer
4. **HMAC unsubscribe tokens**: Stateless, signed with JWT secret — no database table needed
5. **Plain text emails**: No HTML templates in v1 — future enhancement

## New Components

| Component | Location | Purpose |
|-----------|----------|---------|
| Notification service | `backend/internal/notification/` | Email composition, Brevo sending, preference checking |
| Chore batcher | `backend/internal/notification/batcher.go` | Accumulates chore completions, flushes after 5 min |
| Unsubscribe handler | `backend/internal/notification/unsubscribe.go` | Token validation, preference update |
| Migration 014 | `backend/migrations/014_notification_preferences.{up,down}.sql` | Add preference columns to parents |
| NotificationSettings component | `frontend/src/components/NotificationSettings.tsx` | Toggle UI for notification preferences |
| Settings category | `frontend/src/pages/SettingsPage.tsx` | New "Notifications" tab in settings |

## Modified Components

| Component | Change |
|-----------|--------|
| `backend/models/parent.go` | Add 3 boolean fields |
| `backend/repositories/parent_repo.go` | Add `GetByFamilyID`, `UpdateNotificationPrefs` methods |
| `backend/internal/chore/handler.go` | Call notification service after chore completion/approval/rejection |
| `backend/internal/withdrawal/handler.go` | Call notification service after request/approval/denial |
| `backend/internal/settings/handlers.go` | Add notification preference get/update handlers |
| `backend/main.go` | Wire notification service, register new routes |
| `frontend/src/api.ts` | Add notification preference API functions |
| `frontend/src/types.ts` | Add notification preference types |
| `frontend/src/pages/SettingsPage.tsx` | Add Notifications category |

## Implementation Order

1. **Migration + Model**: Schema change, GORM model update
2. **Parent repo methods**: `GetByFamilyID`, `UpdateNotificationPrefs`
3. **Notification service**: Core send logic, preference filtering
4. **Chore batcher**: In-memory batching with flush timer
5. **Unsubscribe handler**: Token generation/validation
6. **Integration into handlers**: Wire notification calls into chore + withdrawal handlers
7. **Settings API**: GET/PUT notification preferences
8. **Frontend**: Notification settings UI
9. **Main.go wiring**: Route registration, dependency injection

## Environment Variables

No new environment variables required. Uses existing `BREVO_API_KEY` and `JWT_SECRET` (for HMAC signing).
