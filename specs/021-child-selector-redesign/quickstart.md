# Quickstart: Child Selector Redesign

## Overview

Replace the vertical two-column child list + detail layout on ParentDashboard and ChildrenSettings with a horizontal chip-based selector bar above full-width content.

## Key Files to Modify

| File | Change |
|------|--------|
| `frontend/src/components/ChildSelectorBar.tsx` | **NEW** — Reusable horizontal chip selector |
| `frontend/src/pages/ParentDashboard.tsx` | Replace two-column grid with selector + full-width content |
| `frontend/src/components/ChildrenSettings.tsx` | Replace two-column grid with selector + full-width content |
| `frontend/src/components/ManageChild.tsx` | Remove close button (selector handles switching) |

## Key Files to Reference (read-only)

| File | Why |
|------|-----|
| `frontend/src/components/ChildList.tsx` | Current data-fetching pattern and child rendering |
| `frontend/src/types.ts` | `Child` interface, `ChildListResponse` |
| `frontend/src/components/AvatarPicker.tsx` | Similar selectable-item pattern to follow |
| `frontend/src/themes.ts` | Color palette reference (forest, cream, sand, etc.) |
| `frontend/src/components/ui/Card.tsx` | Base card component used for chip container |

## Development Flow

1. Create `ChildSelectorBar` component with chip rendering + horizontal scroll
2. Integrate into `ParentDashboard` (replace grid layout)
3. Integrate into `ChildrenSettings` (replace grid layout)
4. Test across family sizes (0, 1, 4, 8, 12) and screen widths
5. Clean up unused `ChildList` references if fully replaced

## Local Development

```bash
cd frontend && npm run dev    # Start Vite dev server
cd frontend && npx tsc --noEmit  # Type check
cd frontend && npm run lint   # Lint
```

No backend changes. No database migrations. No new API endpoints.
