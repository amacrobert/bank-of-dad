# Research: Routed Selections

## R1: React Router v7 Parameterized Routes

**Decision**: Use React Router v7 `useParams` with path parameters for all selection state.

**Rationale**: The project already uses react-router-dom v7.13.0 which fully supports:
- `useParams()` for reading URL path segments
- `useNavigate()` for programmatic URL updates (already used throughout)
- `Navigate` component for declarative redirects
- Nested `<Route path=":param">` for dynamic segments

No additional dependencies needed.

**Alternatives considered**:
- Query parameters (`?tab=`, `?child=`): Already partially used for settings tabs, but path segments are cleaner for hierarchical selections and match user's explicit requirement.
- State management library (Redux, Zustand): Overkill for URL-driven state; React Router already provides this via URL params.

## R2: Child Name in URL vs Child ID

**Decision**: Use lowercase `first_name` in URL path, matched case-insensitively against loaded children.

**Rationale**: The user explicitly requested `/dashboard/bruce` format with first names. Using IDs would produce unfriendly URLs like `/dashboard/42`. The `Child` type already has `first_name: string` available from the API response.

**Alternatives considered**:
- Child ID in URL: More robust for uniqueness but user explicitly wants name-based URLs.
- URL-encoded full name: Unnecessary since only first name is used, and first names within a family are extremely unlikely to collide.

## R3: Redirect Strategy for Invalid Parameters

**Decision**: Use `useEffect` + `navigate(..., { replace: true })` for invalid parameter redirects.

**Rationale**: The `replace: true` option prevents the invalid URL from polluting browser history. This is the idiomatic React Router v7 pattern. The redirect happens after the children list loads, allowing validation against actual data.

**Alternatives considered**:
- Route-level `loader` functions: React Router v7 supports data loaders, but the project doesn't use them anywhere — adding them would be inconsistent.
- `Navigate` component in render: Works for static redirects (settings category validation) but not for async validation (child name requires fetching children list first).

## R4: Settings Route Structure

**Decision**: Use nested routes under `/settings/:category` with a redirect from `/settings` to `/settings/general`. For children settings, use `/settings/children/:childName?` (optional param).

**Rationale**: This maps directly to the existing CATEGORIES array (`general`, `children`, `account`). The redirect from `/settings` → `/settings/general` is handled by a `Navigate` element at the `/settings` index route. Invalid categories redirect via the component's `useEffect`.

**Alternatives considered**:
- Separate route components per category: Would fragment the settings page into multiple components and lose the sidebar navigation. The current single-component approach with conditional rendering based on `activeCategory` is simpler.
- Layout route with nested outlets: More complex than needed since the settings page is already a single component with conditional content panels.

## R5: ChildrenSettings Component — Receiving childName from Parent

**Decision**: Pass `selectedChildName` prop from SettingsPage to ChildrenSettings, and add `onChildSelect` callback that SettingsPage uses to update the URL.

**Rationale**: ChildrenSettings currently manages its own child selection state internally. To make this URL-driven, the selection must be controlled by the parent SettingsPage (which has access to the URL param). This is a standard "lifting state up" pattern — ChildrenSettings becomes a controlled component for selection.

**Alternatives considered**:
- Have ChildrenSettings read `useParams` directly: Creates a hidden coupling between a sub-component and the URL structure. Passing props keeps the data flow explicit.
- Using React context for child selection: Overkill for a single prop being passed one level down.
