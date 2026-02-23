# Data Model: Child Selector Redesign

**No new data entities**. This feature is a pure frontend layout refactor operating on existing data.

## Existing Entities Used

### Child (from `types.ts`)

```typescript
interface Child {
  id: number;
  first_name: string;
  is_locked: boolean;
  balance_cents: number;
  created_at: string;
  avatar?: string | null;
}
```

**Used by the selector**: `id` (key + selection tracking), `first_name` (chip label), `avatar` (chip icon), `is_locked` (lock indicator on chip).

**Not displayed on chip**: `balance_cents`, `created_at` — these are only relevant in the detail panel after selection.

### ChildListResponse (from `types.ts`)

```typescript
interface ChildListResponse {
  children: Child[];
}
```

**API endpoint**: `GET /children` — returns all children for the authenticated parent's family. No changes needed.

## New Component Interfaces

### ChildSelectorBar Props

```typescript
interface ChildSelectorBarProps {
  children: Child[];
  selectedChildId: number | null;
  onSelectChild: (child: Child | null) => void;
  loading?: boolean;
}
```

- `children`: The list of children to display as chips. Fetched by the parent page component.
- `selectedChildId`: Currently selected child ID, or `null` for no selection.
- `onSelectChild`: Callback when a chip is clicked. Passes the `Child` object on select, or `null` on deselect (toggle-off).
- `loading`: Optional loading state to show a skeleton/spinner.

**Design note**: Data fetching responsibility stays in the page component (ParentDashboard, ChildrenSettings) rather than inside the selector. This keeps the component pure and testable, and allows the page to control refresh logic (e.g., after adding/deleting a child).

## State Transitions

```
No selection → Click chip → Child selected (detail panel shown)
Child selected → Click same chip → No selection (toggle-off, empty state shown)
Child selected → Click different chip → Different child selected
Child selected → Child deleted → No selection (handled by page component)
Any state → Child added → Selector re-renders with new child (via refreshKey pattern)
```
