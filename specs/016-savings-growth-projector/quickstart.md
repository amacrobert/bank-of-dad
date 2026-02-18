# Quickstart: Savings Growth Projector

**Feature**: 016-savings-growth-projector
**Date**: 2026-02-17

## Prerequisites

- Node.js and npm (for frontend)
- Running backend with test data (child account with balance, allowance, and/or interest configured)
- PostgreSQL with `bankofdad` database

## New Dependency

```bash
cd frontend && npm install recharts
```

## Files to Create

```
frontend/src/
├── pages/
│   └── GrowthPage.tsx              # Main page component
├── components/
│   ├── GrowthChart.tsx             # Recharts line chart wrapper
│   ├── ScenarioControls.tsx        # What-if input form
│   └── GrowthExplanation.tsx       # Plain-English summary
└── utils/
    └── projection.ts               # Pure projection calculation engine
```

## Files to Modify

```
frontend/src/App.tsx                # Add /child/growth route
frontend/src/components/Layout.tsx  # Add "Growth" nav item for children
```

## Development Flow

1. **Start with the projection engine** (`projection.ts`) — pure function, easily unit-testable
2. **Build the chart component** — wrap Recharts LineChart with projected data
3. **Build the scenario controls** — form inputs that feed into the projection
4. **Build the explanation** — derive text from ProjectionResult
5. **Wire up the page** — compose components, fetch data, manage state
6. **Add routing and navigation** — route in App.tsx, nav item in Layout.tsx

## Running the App

```bash
# Terminal 1: Backend
cd backend && go run ./cmd/server

# Terminal 2: Frontend
cd frontend && npm run dev
```

Navigate to `http://localhost:5173`, log in as a child, and click "Growth" in the sidebar.

## Testing

All projection math should be tested via unit tests on `projection.ts`. Use known compound interest calculations to verify accuracy.

```bash
cd frontend && npm test
```
