# Research: Growth Projector Scenarios

## R-001: Multi-line Recharts Configuration

**Decision**: Use multiple `<Line>` components within the existing `<LineChart>`, one per scenario, each with a unique `dataKey` and `stroke` color.

**Rationale**: Recharts natively supports multiple `<Line>` elements in a single `<LineChart>`. The chart data array needs to be restructured so each data point has a property per scenario (e.g., `balanceCents_0`, `balanceCents_1`). The custom tooltip already receives all payloads and can display multiple scenario values.

**Alternatives considered**:
- Separate charts stacked vertically: Rejected — defeats the purpose of side-by-side comparison on a single graph.
- Canvas-based custom chart: Rejected — unnecessary complexity when Recharts handles this natively.

## R-002: URL Query Parameter Serialization for Scenarios

**Decision**: Serialize the scenario set as a JSON-encoded string in a single `scenarios` query parameter, base64-encoded to avoid URL-encoding issues with special characters.

**Rationale**: A single parameter keeps the URL clean. Base64 encoding avoids issues with brackets, commas, and other JSON characters in URLs. The `horizonMonths` is a separate simple query parameter (`h=12`).

**Alternatives considered**:
- Multiple indexed params (`s0_weekly=500&s0_dir=spending&s1_weekly=0`): Rejected — verbose, hard to parse, URL becomes very long with 5 scenarios.
- Raw JSON in query string: Rejected — requires heavy URL encoding, produces unreadable URLs.
- Compact custom format (`spending:500,deposit:1000|saving:200,withdrawal:0`): Considered — simpler but fragile to parse and extend.

**URL format**: `?scenarios=<base64>&h=<months>`

**Fallback**: If `scenarios` param is missing, malformed, or decodes to invalid data, fall back to default scenarios.

## R-003: Scenario Color Palette

**Decision**: Use a fixed 5-color palette designed for legibility on cream/white backgrounds and distinguishability for color-blind users.

**Rationale**: 5 is the maximum scenario count. Colors should be distinct from each other and from the background. Using a fixed palette (not derived from child themes) ensures consistency.

**Palette**:
1. `#2563eb` — Blue
2. `#dc2626` — Red
3. `#16a34a` — Green
4. `#9333ea` — Purple
5. `#ea580c` — Orange

**Alternatives considered**:
- Child theme color + variations: Rejected — theme colors may not produce 5 distinguishable variants.
- Auto-generated from HSL rotation: Rejected — can produce clashing or low-contrast colors.

## R-004: Weekly Saving Mechanic in Projection Engine

**Decision**: Modify `calculateProjection` to accept `weeklySavingsCents` as a new field on `ScenarioInputs`. When positive, it adds to the balance each week (in addition to any allowance deposits). This is mutually exclusive with `weeklySpendingCents` at the UI level.

**Rationale**: The existing projection engine only has `weeklySpendingCents` (subtraction). Adding `weeklySavingsCents` (addition) keeps the engine logic clean — spending subtracts, savings adds, and the combined UI toggle controls which field is populated.

**Alternatives considered**:
- Negative spending to represent savings: Rejected — confusing semantics, breaks the `balance < 0` floor logic.
- Adding savings to allowance amount: Rejected — conflates two distinct concepts and complicates title generation.

## R-005: Scenario Title Generation Architecture

**Decision**: Implement title generation as a pure function `generateScenarioTitle(config)` in a new `utils/scenarioTitle.ts` module, fully unit-testable. The function takes scenario inputs + allowance context and returns a formatted string.

**Rationale**: The title rules are complex with many branches (allowance vs no allowance, spending vs saving, deposit vs withdrawal). A pure function is easy to test exhaustively and keeps rendering logic separate from generation logic.

**Alternatives considered**:
- Inline in component: Rejected — too complex for inline rendering, hard to test.
- Template string approach: Considered but the conditional logic is too branchy for simple templates.
