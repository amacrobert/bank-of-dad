<!--
SYNC IMPACT REPORT
==================
Version Change: N/A → 1.0.0 (Initial ratification)
Modified Principles: N/A (new document)
Added Sections:
  - Core Principles (3): Test-First, Security-First, Simplicity
  - Technology Stack
  - Development Workflow
  - Governance

Templates Requiring Updates:
  - .specify/templates/plan-template.md: ✅ Compatible (Constitution Check section exists)
  - .specify/templates/spec-template.md: ✅ Compatible (requirements structure aligns)
  - .specify/templates/tasks-template.md: ✅ Compatible (test-first workflow supported)

Follow-up TODOs: None
-->

# Bank of Dad Constitution

## Core Principles

### I. Test-First Development

All feature development MUST follow Test-Driven Development (TDD) practices:

- **Write tests before implementation**: Tests MUST be written and verified to FAIL before any implementation code is written
- **Red-Green-Refactor cycle**: Strictly enforce the cycle of failing test → passing implementation → refactor
- **Test coverage requirements**:
  - Contract tests for all API endpoints
  - Integration tests for user journeys
  - Unit tests for business logic with non-trivial complexity
- **No untested code in production**: All code paths affecting user data or financial calculations MUST have corresponding tests

**Rationale**: Bank of Dad handles financial data for families. Bugs in interest calculations or balance tracking erode trust. TDD catches defects early and documents expected behavior.

### II. Security-First Design

All development MUST prioritize security given the sensitive nature of family financial data:

- **Data protection**: User credentials and financial data MUST be encrypted at rest and in transit
- **Authentication**: All API endpoints except health checks MUST require authentication
- **Authorization**: Users MUST only access their own family's data; parent/child roles MUST be enforced
- **Input validation**: All user input MUST be validated and sanitized before processing
- **Dependency management**: Dependencies MUST be kept up-to-date; known vulnerabilities MUST be addressed within 7 days of disclosure
- **Logging**: Security events (login attempts, permission changes, financial transactions) MUST be logged without exposing sensitive data

**Rationale**: This application manages real money concepts for families. Security breaches would compromise trust and potentially expose children's information.

### III. Simplicity

All design and implementation decisions MUST favor simplicity:

- **YAGNI (You Aren't Gonna Need It)**: Do not implement features, abstractions, or configurations for hypothetical future requirements
- **Minimal dependencies**: Add dependencies only when they provide clear value; prefer standard library solutions
- **Clear over clever**: Code MUST be readable and maintainable by developers unfamiliar with the codebase
- **Kid-friendly UX**: User interfaces MUST be intuitive enough for children to understand their account balances and interest earnings
- **Single responsibility**: Each component, service, and function SHOULD do one thing well
- **No premature optimization**: Optimize only when measurements indicate a performance problem

**Rationale**: Bank of Dad is an educational tool. Complexity obscures the learning experience for kids and increases maintenance burden. Simple systems are more secure and reliable.

## Technology Stack

**Backend**: Go HTTP server
**Frontend**: React + Vite + TypeScript
**Infrastructure**: Docker, Docker Compose
**Storage**: To be determined based on feature requirements

All technology choices MUST align with the Simplicity principle. New technologies require justification demonstrating clear benefit over existing stack.

## Development Workflow

### Code Review Requirements

- All changes MUST be reviewed before merging to main
- Reviews MUST verify compliance with all three Core Principles
- Security-sensitive changes require explicit security review notation

### Quality Gates

- All tests MUST pass before merge
- No new linting errors or warnings
- Build MUST succeed in Docker environment
- API contract changes require documentation updates

### Commit Standards

- Commits SHOULD be atomic and focused on a single change
- Commit messages MUST describe the "why" not just the "what"
- Breaking changes MUST be clearly noted in commit messages

## Governance

### Amendment Process

1. Proposed amendments MUST be documented with rationale
2. Amendments MUST include a migration plan for existing code if applicable
3. Version number MUST be updated according to semantic versioning:
   - **MAJOR**: Principle removal or fundamental redefinition
   - **MINOR**: New principle added or material guidance expansion
   - **PATCH**: Clarifications, typo fixes, non-semantic refinements

### Compliance

- All pull requests MUST verify alignment with Core Principles
- Constitution violations require explicit justification in the Complexity Tracking section of implementation plans
- This constitution supersedes conflicting practices in other documentation

### Review Cadence

- Constitution SHOULD be reviewed quarterly or when major features are planned
- Outdated guidance MUST be updated or removed

**Version**: 1.0.0 | **Ratified**: 2026-01-20 | **Last Amended**: 2026-01-20
