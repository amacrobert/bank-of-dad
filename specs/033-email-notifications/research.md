# Research: Email Notifications

**Feature Branch**: `033-email-notifications`  
**Date**: 2026-04-02

## R1: Email Sending Pattern

**Decision**: Use existing Brevo `TransactionalEmailsApi.SendTransacEmail` pattern from the contact handler.

**Rationale**: The contact handler already demonstrates the exact API call pattern — `SendSmtpEmail` struct with Sender, To, Subject, TextContent fields. Reusing this pattern requires no new dependencies and keeps the codebase consistent.

**Alternatives considered**:
- SMTP direct send: Would require new dependency and connection management. Rejected — Brevo SDK is already in use.
- Brevo templates: Would require managing templates in Brevo dashboard. Rejected — plain text emails are simpler and spec says HTML templates are a future enhancement.

## R2: Notification Preferences Storage

**Decision**: Add three boolean columns to the `parents` table: `notify_withdrawal_requests`, `notify_chore_completions`, `notify_decisions`. All default to `true`.

**Rationale**: Only 3 notification types exist. Adding columns to the existing table is simpler than creating a new table with a many-to-many relationship. The constitution mandates simplicity and YAGNI — a separate preferences table would be over-engineering for 3 booleans.

**Alternatives considered**:
- Separate `notification_preferences` table with notification_type/enabled pairs: More extensible but unnecessary complexity for 3 fixed types. Rejected per Simplicity principle.
- JSONB column on parents: Flexible but harder to query and validate. Rejected.

## R3: Async Email Delivery (Fire-and-Forget)

**Decision**: Send notification emails in goroutines so they never block the handler response. Log errors but do not retry.

**Rationale**: FR-007 requires that email failures never block core actions. The simplest implementation is `go notifier.Send(...)` from within the handler. No queue infrastructure needed at current scale.

**Alternatives considered**:
- Background job queue (e.g., database-backed queue): Provides retry capability but adds significant complexity. Rejected — spec says delivery status tracking is a future enhancement.
- Channel-based worker pool: Adds concurrency control but unnecessary at current scale (single-digit families). Rejected per YAGNI.

## R4: Chore Completion Batching

**Decision**: Use an in-memory batcher with a 5-minute flush timer per family. When the first chore completion arrives, start a timer. Accumulate completions. On timer expiry, send a single summary email and reset.

**Rationale**: The spec requires batching chore completions within a 5-minute window. An in-memory approach is simple and sufficient — if the server restarts during the window, the worst case is a missed batch (acceptable for non-critical notifications).

**Alternatives considered**:
- Database-backed queue with periodic flush: Survives restarts but adds a table and scheduler. Rejected — the cost of a missed batch during restart is near zero.
- Per-event debounce: Resets timer on each event, could delay indefinitely. Rejected — max 5 minutes from first event is more predictable.

## R5: One-Click Unsubscribe

**Decision**: Generate HMAC-signed tokens containing the parent ID. Include a link in every email: `GET /api/notifications/unsubscribe?token=xxx`. Clicking it sets all three notification preferences to `false` and shows a confirmation page.

**Rationale**: HMAC signatures use the existing JWT secret key and require no database lookups to validate. This is the simplest approach that is also secure (tokens can't be forged or guessed).

**Alternatives considered**:
- Database-stored unsubscribe tokens: Provides revocability but adds a table. Rejected — HMAC tokens are stateless and sufficient.
- Login-required unsubscribe: Poor UX, requires the user to authenticate just to unsubscribe. Rejected.

## R6: Parent Lookup for Notifications

**Decision**: Add a `GetByFamilyID(familyID int64) ([]models.Parent, error)` method to `ParentRepo` to fetch all parents in a family.

**Rationale**: The existing repo only has `GetByID` and `GetByGoogleID`. Notifications need to email all parents in a family, so a family-scoped query is needed. GORM makes this trivial: `db.Where("family_id = ?", familyID).Find(&parents)`.

**Alternatives considered**:
- Load via Family model with GORM preload: Works but couples notification logic to the family repo. Rejected — a direct parent query is clearer.

## R7: Frontend Notification Settings

**Decision**: Add a "Notifications" category to the existing SettingsPage categories array. Display toggle switches for each notification type. Save via new API endpoints.

**Rationale**: The settings page already uses a category-based layout with sidebar navigation. Adding a new category follows the established pattern exactly.

**Alternatives considered**:
- Separate notifications settings page: Would break the established settings UX pattern. Rejected.
- Inline in General settings: Would clutter the general section. Rejected — a dedicated category is cleaner.
