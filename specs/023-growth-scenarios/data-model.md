# Data Model: Growth Projector Scenarios

## Entities

### ScenarioConfig (new)

Represents a single "what if" scenario's user-configurable inputs. Replaces the per-field `weeklySpendingCents` / `oneTimeDepositCents` / `oneTimeWithdrawalCents` with combined fields plus direction toggles.

| Field              | Type                          | Description                                              |
|--------------------|-------------------------------|----------------------------------------------------------|
| `id`               | `string`                      | Unique identifier for this scenario (e.g., `"s0"`, `"s1"`) |
| `weeklyAmountCents`| `number`                      | Weekly amount in cents (0 = no weekly change)            |
| `weeklyDirection`  | `'spending' \| 'saving'`      | Whether the weekly amount is spent or saved              |
| `oneTimeAmountCents`| `number`                     | One-time amount in cents (0 = no one-time change)        |
| `oneTimeDirection` | `'deposit' \| 'withdrawal'`   | Whether the one-time amount is a deposit or withdrawal   |
| `color`            | `string`                      | Hex color for graph line and UI indicator                |

**Validation rules**:
- `weeklyAmountCents` >= 0
- `oneTimeAmountCents` >= 0
- When `oneTimeDirection === 'withdrawal'`, `oneTimeAmountCents` <= current balance
- `id` must be unique within a scenario set

### ScenarioInputs (updated)

The existing `ScenarioInputs` interface used by `calculateProjection` is updated to support the saving direction.

| Field                    | Type     | Description                                  | Change     |
|--------------------------|----------|----------------------------------------------|------------|
| `weeklySpendingCents`    | `number` | Weekly spending amount (cents)               | Existing   |
| `weeklySavingsCents`     | `number` | Weekly savings amount (cents)                | **New**    |
| `oneTimeDepositCents`    | `number` | One-time deposit amount (cents)              | Existing   |
| `oneTimeWithdrawalCents` | `number` | One-time withdrawal amount (cents)           | Existing   |
| `horizonMonths`          | `number` | Projection time horizon                      | Existing   |

**Mapping from ScenarioConfig to ScenarioInputs**:
- If `weeklyDirection === 'spending'`: `weeklySpendingCents = weeklyAmountCents`, `weeklySavingsCents = 0`
- If `weeklyDirection === 'saving'`: `weeklySpendingCents = 0`, `weeklySavingsCents = weeklyAmountCents`
- If `oneTimeDirection === 'deposit'`: `oneTimeDepositCents = oneTimeAmountCents`, `oneTimeWithdrawalCents = 0`
- If `oneTimeDirection === 'withdrawal'`: `oneTimeDepositCents = 0`, `oneTimeWithdrawalCents = oneTimeAmountCents`

### MultiScenarioProjectionData (new)

The merged data structure for the multi-line chart. Each data point has balance values for all scenarios.

| Field                    | Type     | Description                                  |
|--------------------------|----------|----------------------------------------------|
| `weekIndex`              | `number` | Week number (0 = today)                      |
| `date`                   | `string` | ISO date string for this week                |
| `balanceCents_0`         | `number` | Balance for scenario 0 at this week          |
| `balanceCents_1`         | `number` | Balance for scenario 1 at this week          |
| `balanceCents_N`         | `number` | Balance for scenario N (dynamic keys)        |

### ScenarioTitleContext (new)

Input for the title generation function.

| Field                   | Type                          | Description                                    |
|-------------------------|-------------------------------|------------------------------------------------|
| `hasAllowance`          | `boolean`                     | Whether the child has an active allowance       |
| `weeklyAllowanceCents`  | `number`                      | Weekly-equivalent allowance amount in cents     |
| `weeklyAmountCents`     | `number`                      | The scenario's weekly amount                    |
| `weeklyDirection`       | `'spending' \| 'saving'`      | Spending or saving                              |
| `oneTimeAmountCents`    | `number`                      | The scenario's one-time amount                  |
| `oneTimeDirection`      | `'deposit' \| 'withdrawal'`   | Deposit or withdrawal                           |

## URL Serialization Schema

**Query parameters**:
- `scenarios`: Base64-encoded JSON array of serialized scenario configs
- `h`: Horizon in months (number)

**Serialized scenario format** (minimal, for URL compactness):
```json
[
  { "w": 500, "wd": "s", "o": 0, "od": "d" },
  { "w": 0, "wd": "s", "o": 5000, "od": "w" }
]
```

Field abbreviations:
- `w` = weeklyAmountCents
- `wd` = weeklyDirection (`"s"` = spending, `"v"` = saving)
- `o` = oneTimeAmountCents
- `od` = oneTimeDirection (`"d"` = deposit, `"w"` = withdrawal)

## Relationships

```
ScenarioConfig[] (1-5 items)
    │
    ├── maps to → ScenarioInputs (one per scenario, for projection engine)
    │                 │
    │                 └── fed to → calculateProjection() → ProjectionResult
    │
    ├── maps to → ScenarioTitleContext → generateScenarioTitle() → string
    │
    └── serialized to/from → URL query parameters

MultiScenarioProjectionData[] ← merged from all ProjectionResult.dataPoints
    │
    └── fed to → GrowthChart (one <Line> per scenario)
```

## Color Palette (fixed)

| Index | Color     | Hex       |
|-------|-----------|-----------|
| 0     | Blue      | `#2563eb` |
| 1     | Red       | `#dc2626` |
| 2     | Green     | `#16a34a` |
| 3     | Purple    | `#9333ea` |
| 4     | Orange    | `#ea580c` |
