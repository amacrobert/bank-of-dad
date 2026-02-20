# Data Model: 018-child-management-settings

## Overview

**No database schema changes required.** This feature is a frontend-only reorganization of existing child management UI. All entities and relationships remain unchanged.

## Existing Entities (unchanged)

### Child
| Field | Type | Notes |
|-------|------|-------|
| id | SERIAL PRIMARY KEY | Auto-increment |
| family_id | INTEGER FK | References families(id) |
| first_name | TEXT NOT NULL | Child's display name |
| password_hash | TEXT NOT NULL | bcrypt-hashed password |
| avatar | TEXT | Optional emoji avatar |
| is_locked | BOOLEAN DEFAULT FALSE | Account lock status |
| created_at | TIMESTAMPTZ | Auto-set on creation |

### Family
| Field | Type | Notes |
|-------|------|-------|
| id | SERIAL PRIMARY KEY | Auto-increment |
| slug | TEXT UNIQUE NOT NULL | URL-friendly family identifier |
| timezone | TEXT DEFAULT 'America/New_York' | Family timezone |

## Existing API Contracts (unchanged)

All endpoints remain exactly as-is. No new endpoints needed.

| Method | Path | Purpose |
|--------|------|---------|
| GET | /api/children | List all children in family |
| POST | /api/children | Create a child account |
| PUT | /api/children/{id}/name | Update child name and avatar |
| PUT | /api/children/{id}/password | Reset child password |
| DELETE | /api/children/{id} | Delete child account |

## Frontend Component Model Changes

### New Components
- `ChildAccountSettings` — Extracted from `ManageChild.tsx`: contains password reset, name/avatar edit, delete account forms for a single child.
- `ChildrenSettings` — Settings sub-page that composes `ChildList`, `AddChildForm`, and `ChildAccountSettings` into a child management experience.

### Modified Components
- `ManageChild.tsx` — Remove "Account Settings" collapsible section (password reset, name/avatar edit, delete). Retain financial management only.
- `ParentDashboard.tsx` — Remove `AddChildForm` and "Add a Child" button. Add empty state when no children, linking to Settings → Children.
- `SettingsPage.tsx` — Add "Children" category to `CATEGORIES` array. Render `ChildrenSettings` when category is active.
- `SetupPage.tsx` — Extend from 2-step to 3-step flow. Step 2 becomes "Add Children" using `AddChildForm`. Step 3 is the existing confirmation.
- `ChildList.tsx` — Update empty state message to reference Settings → Children instead of "Add your first child below!".
