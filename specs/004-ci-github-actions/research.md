# Research: CI Pipeline

## R-001: Backend Linting & Static Analysis Tool

**Decision**: Use `golangci-lint` for Go linting and static analysis

**Rationale**: golangci-lint is the standard Go linting meta-runner. It bundles dozens of linters (including `go vet`, `staticcheck`, `errcheck`, `gosimple`, `unused`) into a single tool with a unified config file. It's well-supported by GitHub Actions with an official action (`golangci/golangci-lint-action`). Running `go vet` separately is unnecessary since golangci-lint includes it.

**Alternatives considered**:
- `go vet` alone: Too limited — only catches a narrow set of issues
- `staticcheck` standalone: Good but golangci-lint subsumes it and adds more
- Individual linters: Fragmented configuration; golangci-lint unifies them

## R-002: Frontend Linting Tool

**Decision**: Use ESLint for TypeScript/React linting

**Rationale**: ESLint is the standard linter for TypeScript and React projects. The project currently has no linter configured. ESLint catches unused variables, type issues, React-specific problems (hooks rules, JSX best practices), and enforces code consistency. The TypeScript compiler already has `strict`, `noUnusedLocals`, and `noUnusedParameters` enabled, so ESLint provides complementary checks.

**Alternatives considered**:
- TypeScript compiler only: Already configured but doesn't catch React-specific issues or code style
- Biome: Newer and faster, but less ecosystem support and plugin availability than ESLint

## R-003: Frontend Static Analysis

**Decision**: Rely on TypeScript strict mode + ESLint as frontend static analysis

**Rationale**: TypeScript's strict mode with `noUnusedLocals`/`noUnusedParameters` already provides strong static analysis. ESLint with TypeScript-aware rules adds further coverage. A separate static analysis tool is unnecessary for a React frontend of this size — this satisfies FR-007/FR-008 without adding complexity.

**Alternatives considered**:
- SonarQube/SonarCloud: Overkill for project size; adds external service dependency
- Separate TypeScript analysis tools: Redundant with strict TypeScript + ESLint

## R-004: Frontend Test Framework

**Decision**: Do not add a frontend test framework in this CI feature

**Rationale**: The frontend currently has zero test files and no test framework configured. Adding vitest/jest infrastructure is a separate feature concern (frontend testing setup), not a CI pipeline concern. The CI pipeline will run `tsc` (type checking) and `vite build` (build verification) which already catch compilation and bundling errors. If/when frontend tests are added later, the CI pipeline will pick them up naturally.

**Alternatives considered**:
- Add vitest now: Violates YAGNI — there are no tests to run. Adding test infrastructure with no tests adds complexity for no value
- Add jest now: Same YAGNI concern, plus jest requires more configuration with Vite

## R-005: GitHub Actions Workflow Structure

**Decision**: Single workflow file with parallel jobs for backend and frontend

**Rationale**: A single `.github/workflows/ci.yml` triggered on PRs keeps configuration simple. Backend and frontend checks run as separate parallel jobs since they are independent — this reduces total pipeline time. Each job installs only its own dependencies.

**Alternatives considered**:
- Separate workflow files per language: More files to maintain with no benefit
- Single sequential job: Slower — backend and frontend don't depend on each other
- Matrix strategy: Unnecessary — we have two distinct stacks, not variations of one

## R-006: Go Build Verification

**Decision**: `go build ./...` as a separate step (implicit in test execution but explicit for clarity)

**Rationale**: `go test ./...` implicitly compiles the code, but an explicit `go build ./...` step makes build failures immediately visible and separate from test failures. This satisfies FR-009 and provides clearer error reporting per FR-012.

**Alternatives considered**:
- Rely on `go test` to catch build errors: Works but conflates build and test failures in output

## R-007: Concurrency Control

**Decision**: Use GitHub Actions `concurrency` key with `cancel-in-progress: true`

**Rationale**: GitHub Actions natively supports concurrency groups. Setting the group to the PR ref with `cancel-in-progress: true` automatically cancels older runs when new commits are pushed, satisfying FR-013 with zero custom logic.

**Alternatives considered**:
- No concurrency control: Wastes runner minutes on stale commits
- Custom cancellation logic: Unnecessary — native support exists

## R-008: Existing Project State

**Findings**:
- Backend: Go 1.24.0, tests exist and use testify, no linter config, no Makefile
- Frontend: TypeScript 5.3.3, React 18.2.0, Vite 5.0.10, strict TS config, no ESLint, no test framework, no test files
- No `.github/workflows/` directory exists
- Docker setup exists but CI should use native runners (faster setup, no Docker build overhead)
- Frontend `package.json` scripts: `dev`, `build` (`tsc && vite build`), `preview` — no `lint` or `test` script
