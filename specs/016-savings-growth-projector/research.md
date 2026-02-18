# Research: Savings Growth Projector

**Feature**: 016-savings-growth-projector
**Date**: 2026-02-17

## R-001: Charting Library Selection

**Decision**: Recharts

**Rationale**: Recharts is a React-native charting library built on D3. It uses declarative JSX components (`<LineChart>`, `<Line>`, `<XAxis>`, etc.) which fits naturally into the existing React + TypeScript codebase. It's lightweight (~45kB gzipped), has strong TypeScript support, and is the most popular React charting library. A simple line chart for balance projection is its bread and butter.

**Alternatives considered**:
- **Chart.js + react-chartjs-2**: Good library, but requires an imperative bridge layer. Less natural in React.
- **visx**: Very flexible but low-level — requires significantly more code for a simple line chart. Better suited for custom/complex visualizations.
- **Nivo**: Built on D3, higher-level than visx, but heavier and more opinionated than Recharts for this use case.
- **No library (canvas/SVG by hand)**: Would violate the Simplicity principle. A line chart is a solved problem.

## R-002: Projection Calculation Approach

**Decision**: Client-side pure function, no new backend endpoints

**Rationale**: The projection uses only data already available from existing endpoints (balance, allowance schedule, interest schedule). The calculation is a simple iterative simulation: step week-by-week through the time horizon, applying allowance deposits, interest compounding, and scenario adjustments at each step. This runs in <1ms for a 5-year projection and requires no server round-trips after initial data load.

**Alternatives considered**:
- **Backend projection endpoint**: Adds unnecessary complexity. The calculation is deterministic given the inputs, has no data dependencies beyond what's already fetched, and doesn't need to be persisted. Violates YAGNI.
- **Web Worker**: Unnecessary for the computational load. A 260-week (5-year) iterative calculation with simple arithmetic completes instantly on any device.

## R-003: Compounding Model for Projections

**Decision**: Schedule-based discrete compounding matching the actual system behavior

**Rationale**: The Bank of Dad system applies interest on a schedule (weekly, biweekly, or monthly) at the rate stored in `interest_rate_bps`. The projection must mirror this to be accurate. Interest is applied as: `balance * (rate_bps / 10000) / periods_per_year` at each compounding event, matching how the backend scheduler processes it.

**Key details**:
- Interest rate is stored in basis points on the `children` table (e.g., 500 = 5.00%)
- Compounding frequency comes from `interest_schedules.frequency` (weekly, biweekly, monthly)
- Per-period rate = `interest_rate_bps / 10000 / periods_per_year`
  - Weekly: `rate / 52`
  - Biweekly: `rate / 26`
  - Monthly: `rate / 12`

**Alternatives considered**:
- **Continuous compounding (e^rt)**: Would give different results than the actual system, confusing children when projections don't match reality.
- **Simple interest**: Not how the system works; would understate growth.

## R-004: Projection Time Granularity

**Decision**: Weekly data points for all time horizons

**Rationale**: Weekly granularity provides smooth curves for all horizons (13 points for 3 months, 260 for 5 years) while keeping computation trivial. It aligns with the smallest compounding period (weekly). For biweekly/monthly compounding, interest is applied only at the appropriate intervals within the weekly step loop.

**Alternatives considered**:
- **Daily granularity**: 1,825 data points for 5 years is excessive for a line chart with no visual benefit. Would slow rendering on low-end devices.
- **Monthly granularity**: Too coarse for 3-month horizon (only 3 data points). Would miss weekly allowance deposits in the curve.

## R-005: Existing API Data Available (No New Endpoints)

**Decision**: Use three existing endpoints to gather all projection inputs

| Data Needed | Endpoint | Field |
|-------------|----------|-------|
| Current balance | `GET /children/{id}/balance` | `balance_cents` |
| Interest rate | `GET /children/{id}/balance` | `interest_rate_bps` |
| Allowance amount & frequency | `GET /children/{id}/allowance` | `amount_cents`, `frequency`, `status` |
| Interest compounding frequency | `GET /children/{id}/interest-schedule` | `frequency`, `status` |

All endpoints already require authentication and enforce child access to their own data.

## R-006: Navigation Placement

**Decision**: Add "Growth" nav item for child users in the sidebar, positioned after "Home", using the `TrendingUp` icon from lucide-react.

**Rationale**: The spec requires navigation after "Home." The `TrendingUp` icon visually represents growth/projections and is already available in the lucide-react dependency. Route will be `/child/growth`.
