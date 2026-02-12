# Research: Child Avatars

**Feature**: 010-child-avatars
**Date**: 2026-02-11

## Research Findings

### R1: Avatar Storage Format

**Decision**: Store avatar as a nullable TEXT column on the `children` table.

**Rationale**: Emojis are UTF-8 strings (1-4 bytes per codepoint, some multi-codepoint). SQLite TEXT handles UTF-8 natively. A nullable column means no avatar = NULL, which maps cleanly to Go `*string` and TypeScript `string | undefined`. No separate table needed for a single optional field.

**Alternatives considered**:
- Separate `avatars` table with FK — over-engineered for a single optional field
- Integer enum mapping to emojis — adds indirection, harder to extend
- Storing in a JSON preferences blob — premature abstraction

### R2: Avatar Validation

**Decision**: No server-side validation of the avatar value. The backend stores whatever string is sent (including empty string treated as null). The frontend controls which emojis are selectable.

**Rationale**: The avatar is purely cosmetic. The fixed emoji set is a frontend UI concern. Server-side validation would couple the backend to a specific emoji list, making updates require backend changes. The worst case of an invalid avatar is a display oddity, not a security or data integrity issue.

**Alternatives considered**:
- Server-side allowlist validation — adds coupling for no safety benefit
- Regex validation for emoji Unicode ranges — fragile and complex

### R3: API Design — Create vs Update

**Decision**: Add optional `avatar` field to the existing `POST /api/children` (create) request body and the existing `PUT /api/children/{id}/name` (update) endpoint. Rename the update endpoint conceptually to handle both name and avatar, but keep the same URL path for backwards compatibility.

**Rationale**: The spec says "rename the form to Update Name and Avatar." This means the same form submits both fields. Extending the existing endpoint is simpler than adding a new one. The `PUT` endpoint already accepts a JSON body with `first_name` — adding `avatar` is a non-breaking additive change.

**Alternatives considered**:
- Separate `PUT /api/children/{id}/avatar` endpoint — unnecessary split for two fields on the same form
- New combined endpoint — the existing one already does what we need

### R4: Avatar Display Locations

**Decision**: Display avatars in the child selection list (`ChildList.tsx`) only. The child login page (`FamilyLogin.tsx`) shows a generic icon, not per-child avatars, because the login is name-based (children type their name, there's no child picker).

**Rationale**: The login page doesn't know which child is logging in until after form submission. There's no child selection list on the login screen — just a text input for name and password. Showing per-child avatars would require listing all children in the family on the login page, which is a different feature (child picker login).

**Alternatives considered**:
- Add a child picker to the login page — out of scope, significant UX change
- Show avatar after login on child dashboard — possible future enhancement but not in spec

### R5: Migration Strategy

**Decision**: Use the existing `addColumnIfNotExists` helper to add `avatar TEXT` to the `children` table. This is the same idempotent pattern used for `balance_cents`, `interest_rate_bps`, and `last_interest_at`.

**Rationale**: Proven pattern in the codebase. Nullable TEXT with no default is correct — existing children get NULL (no avatar), which triggers the first-letter fallback.

### R6: Emoji Set Selection

**Decision**: Use 16 nature/animal-themed emojis from the spec: 🌻 🌿 🍂 🌸 🌊 🌙 ⭐ 🦋 🐝 🍄 🐸 🦊 🐻 🐰 🐢 🎨

**Rationale**: Matches the reference image's aesthetic. Kid-friendly, gender-neutral, visually distinct at small sizes. 16 emojis fill a 4x4 or 8x2 grid neatly. The paint palette (🎨) adds a creative option. All are standard Unicode emojis with broad platform support.

### R7: Reusable Avatar Grid Component

**Decision**: Create a shared `AvatarPicker` component used in both `AddChildForm` and `ManageChild` (update name and avatar form). The emoji list is defined as a constant in this component.

**Rationale**: DRY — both forms need identical avatar selection UI. A single component ensures visual consistency and makes future emoji set changes trivial (one file to edit).
