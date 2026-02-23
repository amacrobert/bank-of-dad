# Quickstart: Routed Selections

## Overview

Convert in-memory selection state (settings category tabs, child selection) into URL-driven routing so that page refreshes, browser back/forward, and deep links all work correctly.

## Scope

**Frontend only** — no backend changes needed.

### Files to modify:
1. `frontend/src/App.tsx` — Add parameterized routes
2. `frontend/src/pages/SettingsPage.tsx` — Read category from URL params instead of query params
3. `frontend/src/pages/ParentDashboard.tsx` — Read child name from URL params
4. `frontend/src/pages/GrowthPage.tsx` — Read child name from URL params
5. `frontend/src/components/ChildrenSettings.tsx` — Accept controlled child selection via props

### Files with minor link updates:
- `frontend/src/pages/ParentDashboard.tsx` — `navigate("/settings?tab=children")` → `navigate("/settings/children")`
- `frontend/src/pages/GrowthPage.tsx` — `navigate("/settings?tab=children")` → `navigate("/settings/children")`

## Key Patterns

### URL param reading
```tsx
import { useParams, useNavigate } from "react-router-dom";

const { childName } = useParams<{ childName: string }>();
const navigate = useNavigate();
```

### Child selection sync with URL
```tsx
// On child select: update URL
const handleSelectChild = (child: Child | null) => {
  if (child) {
    navigate(`/dashboard/${child.first_name.toLowerCase()}`);
  } else {
    navigate("/dashboard");
  }
};

// On mount/param change: resolve child from URL
useEffect(() => {
  if (childName && children.length > 0) {
    const match = children.find(
      c => c.first_name.toLowerCase() === childName.toLowerCase()
    );
    if (match) {
      setSelectedChild(match);
    } else {
      navigate("/dashboard", { replace: true });
    }
  } else if (!childName) {
    setSelectedChild(null);
  }
}, [childName, children]);
```

### Settings category from URL
```tsx
const { category } = useParams<{ category: string }>();
const validCategories = CATEGORIES.map(c => c.key);

// Redirect invalid categories
useEffect(() => {
  if (category && !validCategories.includes(category)) {
    navigate("/settings/general", { replace: true });
  }
}, [category]);
```

## Testing

```bash
cd frontend && npx tsc --noEmit && npm run build
```

Manual verification:
1. Navigate to `/settings` → should redirect to `/settings/general`
2. Click categories → URL should update
3. Refresh → should stay on same category
4. Select child on dashboard → URL should show `/dashboard/{name}`
5. Browser back → should restore previous selection
6. Navigate to `/dashboard/nonexistent` → should redirect to `/dashboard`
