# Implementation Plan: CI Pipeline

**Branch**: `004-ci-github-actions` | **Date**: 2026-02-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-ci-github-actions/spec.md`

## Summary

Add a GitHub Actions CI pipeline that runs automatically on all pull requests. The pipeline runs backend (Go) and frontend (TypeScript/React) checks in parallel: unit tests, linting, static analysis, and build verification. Uses `golangci-lint` for Go linting/static analysis, ESLint for frontend linting, and native build tools for verification. Cancels in-progress runs on new pushes to the same PR.

## Technical Context

**Language/Version**: Go 1.24.0 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend)
**Primary Dependencies**: golangci-lint (new, backend CI only), ESLint + typescript-eslint (new, frontend CI only), GitHub Actions runners
**Storage**: N/A
**Testing**: `go test ./...` (backend, from `backend/` dir), `tsc` strict mode (frontend type checking)
**Target Platform**: GitHub Actions (ubuntu-latest runners)
**Project Type**: Web application (backend + frontend)
**Performance Goals**: All CI checks complete within 10 minutes
**Constraints**: Free-tier GitHub Actions minutes; no external services
**Scale/Scope**: 2 parallel jobs (backend, frontend), ~6 check steps total

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Research Gate

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | CI enforces test execution on every PR — directly supports this principle. The pipeline itself is configuration (YAML), not application code, so TDD doesn't apply to the workflow file itself. |
| II. Security-First Design | PASS | CI adds a quality gate that catches security-related issues (static analysis). No secrets or sensitive data are processed by the pipeline. |
| III. Simplicity | PASS | Single workflow file, two parallel jobs, standard tools. No custom scripts, no external services, no Docker-based CI (uses native runners for speed). |

### Quality Gates (from Constitution)

| Gate | How CI Addresses It |
|------|---------------------|
| All tests MUST pass before merge | CI runs `go test ./...` on every PR |
| No new linting errors or warnings | CI runs golangci-lint (backend) and ESLint (frontend) |
| Build MUST succeed | CI runs `go build ./...` and `tsc && vite build` |
| API contract changes require documentation updates | Out of scope for this feature (manual review concern) |

### Post-Design Re-check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | No application code is being written — this is infrastructure config |
| II. Security-First Design | PASS | No secrets stored in workflow; uses GitHub-provided runner environment |
| III. Simplicity | PASS | Minimal tooling: one config file, two standard linters, native build commands |

## Project Structure

### Documentation (this feature)

```text
specs/004-ci-github-actions/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
.github/
└── workflows/
    └── ci.yml               # GitHub Actions workflow (new)

backend/
├── .golangci.yml            # golangci-lint config (new)
└── ...                      # existing Go source + tests

frontend/
├── eslint.config.js         # ESLint flat config (new)
├── package.json             # updated: add lint script + ESLint devDeps
└── ...                      # existing React/TS source
```

**Structure Decision**: This feature adds configuration files only — no application source code changes. The GitHub Actions workflow lives at `.github/workflows/ci.yml` (standard location). Linter configs are co-located with their respective projects (backend/.golangci.yml, frontend/eslint.config.js).

## Design Decisions

### D-001: Workflow File Structure

Single file `.github/workflows/ci.yml` with two parallel jobs:

```
ci.yml
├── backend job (ubuntu-latest)
│   ├── Checkout
│   ├── Setup Go 1.24
│   ├── Cache Go modules
│   ├── Build verification (go build ./...)
│   ├── Lint + static analysis (golangci-lint)
│   └── Unit tests (go test ./...)
│
└── frontend job (ubuntu-latest)
    ├── Checkout
    ├── Setup Node 20
    ├── Cache node_modules
    ├── Install dependencies (npm ci)
    ├── Lint (eslint)
    ├── Type check (tsc --noEmit)
    └── Build verification (vite build)
```

### D-002: golangci-lint Configuration

Minimal `.golangci.yml` enabling the most valuable linters without being overly strict for a small project:

- **Enabled**: `govet`, `staticcheck`, `errcheck`, `gosimple`, `unused`, `ineffassign`, `gocritic`
- **Disabled by default**: Style-only linters that would generate noise without safety benefit
- **Timeout**: 5 minutes (generous for CI)

### D-003: ESLint Configuration

Flat config (`eslint.config.js`) with TypeScript and React support:

- **Plugins**: `@typescript-eslint`, `eslint-plugin-react-hooks`, `eslint-plugin-react-refresh`
- **Config base**: `@eslint/js` recommended + `typescript-eslint` recommended
- **Scope**: Only `src/` directory (skip config files, build output)

### D-004: Frontend Testing Strategy

No test framework is added in this feature. The frontend has zero test files today. The CI pipeline runs `tsc --noEmit` (type checking) and `vite build` (bundling) as quality gates. When frontend tests are added in a future feature, the CI pipeline will naturally pick them up by adding a test step.

### D-005: Concurrency

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.head_ref || github.run_id }}
  cancel-in-progress: true
```

This cancels in-progress CI runs when a new commit is pushed to the same PR branch.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
