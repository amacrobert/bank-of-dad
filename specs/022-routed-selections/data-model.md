# Data Model: Routed Selections

This feature is frontend-only. No database changes, no new API endpoints, no backend modifications.

## URL Schema

| Page | URL Pattern | Parameters | Fallback |
|------|------------|------------|----------|
| Settings (index) | `/settings` | none | Redirect → `/settings/general` |
| Settings category | `/settings/:category` | `category`: `general` \| `children` \| `account` | Invalid → redirect to `/settings/general` |
| Settings children + child | `/settings/children/:childName` | `childName`: lowercase first name | Invalid → redirect to `/settings/children` |
| Dashboard | `/dashboard` | none | No selection |
| Dashboard + child | `/dashboard/:childName` | `childName`: lowercase first name | Invalid → redirect to `/dashboard` |
| Growth | `/growth` | none | No selection |
| Growth + child | `/growth/:childName` | `childName`: lowercase first name | Invalid → redirect to `/growth` |

## Existing Entities Used (unchanged)

### Child
- `id`: number (primary key)
- `first_name`: string (used for URL matching — lowercase comparison)
- `balance_cents`: number
- `is_locked`: boolean
- `avatar`: string | null

### Settings Category
- Static list: `["general", "children", "account"]`
- Defined in `CATEGORIES` array in SettingsPage.tsx
- `key` field used as URL slug

## Child Name Resolution

URL `childName` parameter → matched against `children` array:
1. Fetch children list from `/children` API (existing endpoint, no changes)
2. Find first child where `child.first_name.toLowerCase() === childName.toLowerCase()`
3. If match found: select that child
4. If no match: redirect to parent route (remove child segment) using `navigate(..., { replace: true })`
