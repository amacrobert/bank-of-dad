# Research: Child Auto-Setup

## Decision 1: Frontend-only vs. Backend Composite Endpoint

**Decision**: Frontend-only — orchestrate existing endpoints from the client after child creation.

**Rationale**: All three operations (deposit, allowance, interest) already have fully tested backend endpoints with proper validation, auth, and error handling. Creating a composite backend endpoint would duplicate validation logic, add a new handler, and require new tests — all for a flow that only happens once per child.

**Alternatives considered**:
- **New composite `POST /api/children` with setup fields**: Would require significant backend changes (new request struct, importing stores from 3 different packages, transaction coordination). Rejected because the complexity far outweighs the benefit of atomicity for a one-time setup flow.
- **Backend orchestrator service**: Over-engineered for 3 sequential API calls. Rejected per Simplicity principle.

## Decision 2: Error Handling Strategy for Partial Failures

**Decision**: Best-effort sequential execution. If child creation succeeds but a subsequent call fails, show a warning naming the failed step. The child is kept, and successful operations are retained.

**Rationale**: The probability of failure after successful child creation is very low (same authenticated session, same server). If it does happen, the parent can configure the failed item from the child's settings page. Making this fully atomic would require a saga pattern or composite endpoint — both violate Simplicity.

**Alternatives considered**:
- **Rollback child on failure**: Destructive and confusing — the parent just created the child. Rejected.
- **Retry with exponential backoff**: Over-engineered for a one-time form submission. Rejected.

## Decision 3: Allowance Day-of-Week Default

**Decision**: Default to the current day of the week (JavaScript `new Date().getDay()`).

**Rationale**: The spec calls for a simple weekly allowance with no day picker. The most intuitive default is "same day each week starting from today." This matches user expectation: "I set up $10/week on Tuesday → allowance runs every Tuesday."

**Alternatives considered**:
- **Always Monday**: Arbitrary and potentially confusing. Rejected.
- **Let parent choose**: Adds UI complexity to a field meant to be quick. The dedicated allowance form already provides this — parents can adjust later. Rejected per Simplicity.

## Decision 4: Interest Schedule Fixed to Monthly on 1st

**Decision**: Hardcode frequency=monthly, day_of_month=1 when interest rate is provided.

**Rationale**: The spec explicitly states "monthly schedule on the first of the month." No user choice needed. The existing interest form allows changing frequency/day later if desired.

**Alternatives considered**: None — spec is explicit.

## Decision 5: Field Layout

**Decision**: Three fields in a horizontal row on desktop (using CSS grid `grid-cols-3`), stacking to vertical on mobile (`grid-cols-1` on small screens).

**Rationale**: The spec says "three fields next to each other." A 3-column grid with responsive breakpoint is the simplest approach and matches Tailwind patterns used elsewhere in the app.

**Alternatives considered**:
- **Accordion/collapsible section**: Hides the fields, reducing discoverability. Rejected.
- **Separate step in form**: Adds friction — defeats the "single form" purpose. Rejected.
