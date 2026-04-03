# Implementation Plan: Email Notifications

**Branch**: `033-email-notifications` | **Date**: 2026-04-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/033-email-notifications/spec.md`

## Summary

Add email notifications so parents are automatically alerted when children complete chores, request withdrawals, or when co-parents make approval decisions. Uses the existing Brevo transactional email service with fire-and-forget delivery, in-memory batching for chore completions, and per-parent notification preferences stored as boolean columns on the parents table.

## Technical Context

**Language/Version**: Go 1.26 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)  
**Primary Dependencies**: Brevo Go SDK (`github.com/getbrevo/brevo-go`), GORM (`gorm.io/gorm`), pgx/v5 (driver), Vite + Tailwind CSS 4 (frontend)  
**Storage**: PostgreSQL 17 — 3 new columns on existing `parents` table  
**Testing**: `go test -p 1 ./...` (backend), `npx tsc --noEmit && npm run build` (frontend)  
**Target Platform**: Linux server (Docker), web browser  
**Project Type**: Web service (Go backend + React SPA)  
**Performance Goals**: Emails delivered within 2 minutes of trigger; email failures add < 500ms to handler response  
**Constraints**: Fire-and-forget delivery — email failures must never block core actions  
**Scale/Scope**: Single-digit concurrent families; 6 notification trigger points across 2 handlers

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development ✅

- Contract tests for new API endpoints (GET/PUT /api/settings/notifications, GET /api/notifications/unsubscribe)
- Integration tests for notification dispatch (chore complete → email sent, withdrawal request → email sent)
- Unit tests for: batcher flush logic, HMAC token generation/validation, preference filtering, email template rendering
- All tests will be written before implementation per TDD cycle

### II. Security-First Design ✅

- Unsubscribe endpoint uses HMAC-signed tokens (not guessable parent IDs)
- Notification preference endpoints require parent authentication (`requireParent` middleware)
- Unsubscribe tokens are stateless and tamper-proof (signed with JWT secret)
- No sensitive data in email bodies (no passwords, no balance amounts in subject lines)
- Email addresses come from Google OAuth (already validated)

### III. Simplicity ✅

- No new tables — 3 boolean columns on existing `parents` table
- No job queue — fire-and-forget goroutines at current scale
- No HTML templates — plain text emails
- No delivery tracking — future enhancement
- Reuses existing Brevo client, existing JWT secret for HMAC
- In-memory batcher with simple timer (no external dependencies)

**Post-Phase 1 Re-check**: All gates still pass. No new complexity introduced during design.

## Project Structure

### Documentation (this feature)

```text
specs/033-email-notifications/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── api.md           # API contract definitions
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── main.go                          # Wire notification service, register routes
├── migrations/
│   ├── 014_notification_preferences.up.sql
│   └── 014_notification_preferences.down.sql
├── models/
│   └── parent.go                    # Add 3 boolean fields
├── repositories/
│   └── parent_repo.go               # Add GetByFamilyID, UpdateNotificationPrefs
├── internal/
│   ├── notification/                # NEW package
│   │   ├── service.go               # Core notification dispatch + email composition
│   │   ├── batcher.go               # In-memory chore completion batcher
│   │   ├── unsubscribe.go           # HMAC token generation/validation + handler
│   │   └── service_test.go          # Tests
│   ├── chore/
│   │   └── handler.go               # Add notification calls
│   ├── withdrawal/
│   │   └── handler.go               # Add notification calls
│   └── settings/
│       └── handlers.go              # Add notification preference endpoints

frontend/
├── src/
│   ├── api.ts                       # Add notification preference API functions
│   ├── types.ts                     # Add notification preference types
│   ├── components/
│   │   └── NotificationSettings.tsx # NEW toggle UI component
│   └── pages/
│       └── SettingsPage.tsx         # Add Notifications category
```

**Structure Decision**: Follows existing project layout. New `notification` package under `backend/internal/` consistent with other domain packages (chore, withdrawal, settings). Frontend extends existing settings page pattern.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
