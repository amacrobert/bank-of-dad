# Research: Combine Transaction Cards

**Feature**: 014-combine-transactions
**Date**: 2026-02-16

## Research Summary

This feature is a straightforward frontend refactor with no unknowns requiring external research. All decisions are based on existing codebase patterns.

## Decision 1: Component Architecture

**Decision**: Create a single `TransactionsCard` component that owns both data fetching (upcoming) and rendering (upcoming + recent), accepting recent transactions as a prop.

**Rationale**: The `UpcomingPayments` component currently fetches its own data internally (upcoming allowances and interest schedule). The `TransactionHistory` component is a pure presentational component that receives transactions as a prop. The combined component should:
- Keep the upcoming data fetching internal (same pattern as current `UpcomingPayments`)
- Accept recent transactions as a prop (same pattern as current `TransactionHistory`)
- Accept `childId`, `balanceCents`, and `interestRateBps` props (same as current `UpcomingPayments`)

**Alternatives considered**:
- **Option A: Lift all data fetching to parent pages** — Rejected because it would require refactoring both `ChildDashboard.tsx` and `ManageChild.tsx` to fetch upcoming allowances/interest, changing more code than necessary.
- **Option B: Have the combined component fetch everything including recent transactions** — Rejected because `ChildDashboard` and `ManageChild` already fetch transactions for other purposes (e.g., balance calculation context). Duplicating that fetch would be wasteful.

## Decision 2: Section Rendering When Empty

**Decision**: Always render both "Upcoming" and "Recent" section headings. Show "No upcoming payments" or "No transactions yet." (matching existing empty states) when a section has no data.

**Rationale**: Consistent with spec requirement FR-005. Both sections are always visible so users understand the card structure regardless of data state. This matches the current behavior where `TransactionHistory` shows "No transactions yet." when empty. For `UpcomingPayments`, the current behavior is to return `null` (hide entirely) when no upcoming payments exist — the new combined card changes this to always show the section with an empty message for consistency.

**Alternatives considered**:
- **Hide empty sections entirely** — Rejected because it would make the card look like just a "Recent" or "Upcoming" card, losing the consolidated context.

## Decision 3: Card Title

**Decision**: Use "Transactions" as the card title.

**Rationale**: Directly from the feature description. Clean, concise, and contextually clear.

**Alternatives considered**: None — spec is explicit.

## Decision 4: Sub-heading Labels

**Decision**: Use "Upcoming" and "Recent" as sub-heading labels.

**Rationale**: Directly from the feature description. The current "Upcoming allowance and interest" title is context-dependent and varies based on data. Within a "Transactions" card, "Upcoming" is sufficient. Similarly, "Recent" replaces "Recent Activity" / "Transaction History".

**Alternatives considered**: None — spec is explicit.

## Decision 5: Preserving Visual Styling

**Decision**: Preserve all existing icon mappings, color coding, badge styling, amount formatting, and date formatting from both original components.

**Rationale**: Spec requirements FR-006 (no information loss) and SC-002 (100% data preservation). The visual treatments are already well-established:
- Upcoming: sky-blue badges for allowance, sage-green for interest, `~` prefix for estimates, "MMM D" dates
- Recent: type-specific icons (ArrowDownCircle, ArrowUpCircle, Calendar, TrendingUp), color coding (forest green, terracotta, amber), `+/-` amount prefixes, "MMM D, YYYY" dates

**Alternatives considered**: None — preservation is required.
