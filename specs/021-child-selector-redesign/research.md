# Research: Child Selector Redesign

## R1: Horizontal Scroll with Overflow Indicators

**Decision**: CSS `overflow-x: auto` with gradient fade masks on edges

**Rationale**: The simplest approach using native browser scrolling. Gradient fades (via CSS `mask-image` or pseudo-elements with `pointer-events: none`) visually indicate overflowing content without adding JavaScript scroll listeners or arrow button complexity. This is well-supported across all modern browsers and works naturally with touch devices.

**Alternatives considered**:
- **Arrow buttons (prev/next)**: More complex, requires scroll position tracking via `IntersectionObserver` or `scrollLeft`, and duplicates native scroll behavior. Overkill for this use case.
- **CSS Scroll Snap**: Adds snap behavior but doesn't solve the indicator problem alone. Could be combined with fades but adds complexity without clear benefit for a chip selector.
- **Wrapping to multiple rows**: Explicitly rejected by spec (FR-005: "single horizontal row").

## R2: Chip Component Design Pattern

**Decision**: Inline `button` elements within a flex container, not a separate `ui/Chip.tsx` base component

**Rationale**: The AvatarPicker already uses a similar pattern of selectable button items in a grid. The child chip has specific requirements (avatar, name, lock icon, selected state) that don't generalize to a reusable primitive. Creating a `ChildSelectorBar` component that renders chips internally is simpler and more cohesive than abstracting a generic `Chip` component that only has one consumer. Follows the Simplicity principle.

**Alternatives considered**:
- **Generic `ui/Chip.tsx`**: Premature abstraction. Only one use case exists. If future features need chips, extract then.
- **Radio button group with custom styling**: Semantically closer to single-selection, but toggle-off (FR-012) breaks the radio pattern. Buttons with `aria-pressed` are the correct semantic choice.

## R3: Balance Display on Chips

**Decision**: Chips show avatar, name, and balance.

**Rationale**: FR-003 says "compact chip showing their avatar (emoji or initial) and first name." The edge case explicitly states "The chip should show the avatar and name, keeping chips compact." Balance is nto explicitly required, but it's important contect in most situations where a child can be selected. The balance is also displayed in the full ManageChild panel after selection.

## R4: Component Reuse vs. ChildList Replacement

**Decision**: Create a new `ChildSelectorBar` component; keep `ChildList` temporarily (may be removed if no longer imported)

**Rationale**: The existing `ChildList` renders vertical list items with balance, chevrons, and a heading. Its layout and information density are fundamentally different from the new horizontal chip bar. Modifying `ChildList` to support both layouts would be more complex than creating a focused new component. The data-fetching pattern (calling `get<ChildListResponse>("/children")`) is trivial to duplicate.

**Alternatives considered**:
- **Refactor ChildList with a `variant` prop**: Adds conditional rendering complexity without benefit. The two layouts share almost no markup.
- **Extract shared hook for child fetching**: Only two consumers, each with slightly different loading needs. Not worth the abstraction yet.

## R5: Selected State Management

**Decision**: `selectedChild` state stays in each consuming page component (ParentDashboard, ChildrenSettings), passed to `ChildSelectorBar` via props

**Rationale**: This matches the current pattern. Both pages already manage `selectedChild` as local state and pass it to child components. No need for context or global state since selection is page-scoped.

**Alternatives considered**:
- **React Context for selected child**: Overkill for two pages. Selection doesn't need to survive navigation.
- **URL params for selected child**: Would enable deep linking but adds complexity not requested by the spec.

## R6: Frontend Testing Strategy

**Decision**: No component tests for this feature; rely on manual acceptance testing per spec scenarios

**Rationale**: The project has no frontend testing infrastructure (no test runner, no component test files). Setting up Vitest + React Testing Library + JSDOM for a layout refactor would be scope creep and violates the Simplicity principle. The acceptance scenarios are straightforward visual/interaction checks. The constitution's Test-First principle primarily targets backend code handling financial data — this feature changes no financial logic.

**Alternatives considered**:
- **Set up Vitest + RTL**: Significant infrastructure investment for a pure layout change. Should be its own feature if desired.
- **E2E tests with Playwright**: Even more infrastructure. Not justified for this scope.
