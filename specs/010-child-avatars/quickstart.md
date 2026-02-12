# Quickstart: Child Avatars

**Feature**: 010-child-avatars
**Branch**: `010-child-avatars`

## Overview

Add optional emoji avatars to child accounts. Parents pick an avatar from a grid of 16 emojis when creating or updating a child. The avatar replaces the first-letter initial in the child selection list.

## Files to Change

### Backend

| File | Change |
|------|--------|
| `backend/internal/store/sqlite.go` | Add `avatar TEXT` column migration |
| `backend/internal/store/child.go` | Add `Avatar *string` to struct; update Create, GetByID, GetByFamilyAndName, ListByFamily queries; add/extend UpdateName to handle avatar |
| `backend/internal/store/child_test.go` | Tests for avatar in Create, Get, List, Update |
| `backend/internal/family/handlers.go` | Parse avatar in HandleCreateChild and HandleUpdateName; return avatar in HandleListChildren response |

### Frontend

| File | Change |
|------|--------|
| `frontend/src/types.ts` | Add `avatar?: string` to `Child` and `ChildCreateResponse` interfaces |
| `frontend/src/components/AvatarPicker.tsx` | **New** — shared emoji grid component |
| `frontend/src/components/AddChildForm.tsx` | Add AvatarPicker, send avatar in POST |
| `frontend/src/components/ManageChild.tsx` | Rename "Update Name" to "Update Name and Avatar", add AvatarPicker, send avatar in PUT |
| `frontend/src/components/ChildList.tsx` | Conditional rendering: emoji avatar or first-letter fallback |

## Implementation Order

1. Database migration (avatar column)
2. Go struct + store methods
3. Handler changes (create, list, update)
4. Backend tests
5. Frontend types
6. AvatarPicker component
7. AddChildForm integration
8. ManageChild integration
9. ChildList display

## Key Decisions

- Avatar is optional (NULL = no avatar, show first letter)
- No server-side emoji validation
- Single `AvatarPicker` component reused in both forms
- Tap-to-select, tap-again-to-deselect
- Existing `PUT /api/children/{id}/name` endpoint extended (not a new endpoint)
- Login page NOT modified (no child picker on that screen)
