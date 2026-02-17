# Research: Parent Settings Page

**Feature Branch**: `013-parent-settings`
**Date**: 2026-02-16

## R1: Timezone Storage Strategy

**Decision**: Add a `timezone` TEXT column to the existing `families` table with default `'America/New_York'`.

**Rationale**: The spec states timezone applies at the family level. Adding a column to `families` is the simplest approach — no new tables, no new joins. The IANA timezone identifier is a short string (max ~30 chars), making TEXT appropriate.

**Alternatives considered**:
- Separate `family_settings` table with key-value rows: Over-engineered for a single setting. Violates Simplicity principle. Can always migrate later if dozens of settings emerge.
- JSONB column on `families`: Loses type safety and makes individual settings harder to query/index. Unnecessary complexity.
- Separate `family_settings` table with typed columns: Extra table and join for no immediate benefit. One column on `families` is simpler.

## R2: Timezone Validation

**Decision**: Validate timezone values server-side using Go's `time.LoadLocation()`. This ensures only valid IANA timezone identifiers are accepted.

**Rationale**: `time.LoadLocation()` is a standard library function that loads from the system's IANA timezone database. No external dependency needed. Invalid identifiers return an error, providing a clean validation path.

**Alternatives considered**:
- Hardcoded allowlist: Requires manual updates when new timezones are added. Unnecessary maintenance burden.
- No validation: Could lead to corrupt data that breaks time calculations. Unacceptable for financial scheduling.

## R3: Frontend Timezone List

**Decision**: Use a curated list of ~40 commonly used timezones embedded in the frontend, with IANA identifiers as values and human-friendly labels as display text. Implement a searchable/filterable dropdown.

**Rationale**: The full IANA database has ~400+ entries, many of which are legacy aliases or obscure. A curated list covers >99% of users while keeping the UI clean. The search/filter allows finding any included timezone quickly. The existing `Select` UI component can be extended or a custom searchable select can be built.

**Alternatives considered**:
- Full IANA list from server API: Adds an API endpoint for static data. Over-engineered.
- Browser `Intl.supportedValuesOf('timeZone')`: Not supported in all browsers. Inconsistent across environments.
- Third-party timezone library (e.g., moment-timezone): Adds a heavy dependency for a simple list. Violates Simplicity principle.

## R4: Settings Page Architecture

**Decision**: Build the settings page with a category-based sidebar (desktop) / tab (mobile) navigation pattern. Categories are defined as a static configuration array in the frontend. Each category maps to a component. Only categories with content are rendered.

**Rationale**: The spec emphasizes extensible information architecture. A static category registry makes it trivial to add new categories later — just add an entry to the array and create the component. No backend changes needed to add a new category.

**Alternatives considered**:
- Server-driven settings schema: Over-engineered. Settings categories are a UI concern, not a data concern.
- Single flat page with sections: Doesn't scale to future categories like Notifications, My Bank, etc. Poor IA.
- Route-per-category (e.g., `/settings/general`, `/settings/notifications`): Adds routing complexity for categories that are lightweight. A single `/settings` route with client-side category switching is simpler.

## R5: Settings API Design

**Decision**: Use `GET /api/settings` to retrieve all family settings and `PUT /api/settings/timezone` to update the timezone. Parent-only endpoints.

**Rationale**: A single GET endpoint returns all settings at once (currently just timezone), keeping the frontend simple. Individual PUT endpoints per setting allow granular updates without sending the entire settings object. This scales well — adding a new setting means adding a new PUT endpoint.

**Alternatives considered**:
- `PATCH /api/settings` with partial updates: More complex to validate. Individual endpoints are clearer.
- Settings on the family endpoint (`GET /api/families/{id}`): Mixes concerns. Settings are a separate feature area.

## R6: Navigation Entry Point

**Decision**: Add a Settings gear icon to both the desktop nav bar (between display name and logout) and the mobile bottom tab bar. Only visible to parent users.

**Rationale**: The Layout component already has conditional rendering for parent vs. child. Adding a settings button follows the same pattern. A gear icon (lucide-react `Settings` icon) is universally recognized.

**Alternatives considered**:
- Settings in a dropdown menu: Adds UI complexity. The nav is simple enough to accommodate a direct link.
- Settings accessible only via URL: Poor discoverability. Users expect a nav entry point.
