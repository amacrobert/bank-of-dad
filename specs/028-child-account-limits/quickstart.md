# Quickstart: Free Tier Child Account Limits

## What This Feature Does

Limits free-tier families to 2 active child accounts. Children beyond the 2nd are created in a "disabled" state — they can't log in, receive transactions, or collect allowance/interest. Upgrading to Plus enables all children. Downgrading disables children beyond the earliest 2.

## Key Files to Modify

### Backend

| File | What Changes |
|------|-------------|
| `backend/migrations/006_add_child_disabled.up.sql` | Add `is_disabled` column to `children` table |
| `backend/internal/store/child.go` | Add `IsDisabled` to struct, new methods: `EnableAllChildren`, `DisableExcessChildren`, `CountEnabledByFamily`, `ReconcileChildLimits` |
| `backend/internal/store/schedule.go` | Add `AND c.is_disabled = FALSE` to `ListDue` and `ListAllActiveWithTimezone` queries |
| `backend/internal/store/interest_schedule.go` | Add `AND c.is_disabled = FALSE` to `ListDue` query |
| `backend/internal/store/interest.go` | Add `AND c.is_disabled = FALSE` to `ListDueForInterest` query |
| `backend/internal/family/handlers.go` | Change `HandleCreateChild` to allow >2 children, set `is_disabled` for free tier |
| `backend/internal/auth/child.go` | Add `is_disabled` check in `HandleChildLogin` |
| `backend/internal/balance/handler.go` | Add `is_disabled` check in `HandleDeposit` and `HandleWithdraw` |
| `backend/internal/subscription/handlers.go` | Hook child enable/disable into webhook handlers |

### Frontend

| File | What Changes |
|------|-------------|
| `frontend/src/types.ts` | Add `is_disabled` to `Child` interface |
| `frontend/src/components/ChildSelectorBar.tsx` | Grayed-out styling, tooltip, upgrade CTA for disabled children |
| `frontend/src/pages/ParentDashboard.tsx` | Filter disabled children from selection |
| `frontend/src/pages/FamilyLogin.tsx` | Filter disabled children from login selector |

## Development Sequence

1. **Migration** — Add `is_disabled` column
2. **Store layer** — Add field to struct, scanning, new methods
3. **Scheduler filters** — Update SQL queries to skip disabled children
4. **Backend guards** — Login, deposit, withdraw rejection
5. **Child creation logic** — Free tier disabled logic in `HandleCreateChild`
6. **Webhook hooks** — Enable on upgrade, disable on downgrade
7. **Frontend types** — Add `is_disabled` to `Child` interface
8. **ChildSelectorBar** — Visual treatment for disabled children
9. **Dashboard/login filtering** — Skip disabled children in selection

## Testing

```bash
# Backend tests
cd backend && go test -p 1 ./...

# Frontend
cd frontend && npx tsc --noEmit && npm run build && npm run lint
```
