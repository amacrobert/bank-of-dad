# Tasks: CI Pipeline

**Input**: Design documents from `/specs/004-ci-github-actions/`
**Prerequisites**: plan.md (required), spec.md (required), research.md

**Tests**: Not explicitly requested — no test tasks generated. CI pipeline correctness is verified by running the pipeline itself (Phase 4).

**Organization**: Tasks are grouped by user story priority. Note that US3 (Linting) and US4 (Static Analysis) are combined in Phase 3 because `golangci-lint` serves both purposes for the backend, and `ESLint` + `tsc` serve both for the frontend.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Create directory structure and skeleton workflow file

- [x] T001 Create `.github/workflows/ci.yml` with PR trigger (`on: pull_request`), concurrency group (`cancel-in-progress: true` keyed on `github.head_ref`), and two empty job definitions (`backend` and `frontend`) on `ubuntu-latest`

**Checkpoint**: Workflow file exists and would trigger on PRs (no steps yet)

---

## Phase 2: US1 + US2 + US5 — Core CI Pipeline (Priority: P1) — MVP

**Goal**: A working CI pipeline that triggers on PRs, runs backend build + tests, and runs frontend build — covering the most critical quality gates.

**Independent Test**: Open a PR and verify two jobs appear, backend compiles and tests pass, frontend builds successfully. Then push a commit with a broken test and verify it fails.

**Covers**: US1 (PR triggers + status), US2 (unit tests), US5 (build verification)

- [x] T002 [US1] Configure backend job environment in `.github/workflows/ci.yml`: checkout (`actions/checkout`), Go 1.24 setup (`actions/setup-go`), Go module cache, working directory set to `backend/`
- [x] T003 [US1] Configure frontend job environment in `.github/workflows/ci.yml`: checkout (`actions/checkout`), Node 20 setup (`actions/setup-node`), npm cache, `npm ci` install step, working directory set to `frontend/`
- [x] T004 [US5] Add backend build verification step (`go build ./...`) to backend job in `.github/workflows/ci.yml`
- [x] T005 [US2] Add backend unit test step (`go test ./...`) to backend job in `.github/workflows/ci.yml`
- [x] T006 [US5] Add frontend type check (`tsc --noEmit`) and build verification (`vite build`) steps to frontend job in `.github/workflows/ci.yml`

**Checkpoint**: Pipeline triggers on PRs, backend compiles + tests run, frontend type checks + builds. This is a functional MVP.

---

## Phase 3: US3 + US4 — Linting & Static Analysis (Priority: P2)

**Goal**: Add linting and static analysis checks to both backend and frontend CI jobs. `golangci-lint` covers both linting (US3) and static analysis (US4) for Go. ESLint covers linting for TypeScript/React. TypeScript strict mode (already in Phase 2 via `tsc --noEmit`) covers frontend static analysis.

**Independent Test**: Push a PR with a linting violation (e.g., unused variable in Go, missing React hook dependency) and verify the pipeline catches it with a clear error message.

- [x] T007 [P] [US3] Create golangci-lint configuration in `backend/.golangci.yml` — enable linters: `govet`, `staticcheck`, `errcheck`, `gosimple`, `unused`, `ineffassign`, `gocritic`. Set run timeout to 5 minutes.
- [x] T008 [P] [US3] Install ESLint dev dependencies and add `lint` script to `frontend/package.json`. Dependencies: `eslint`, `@eslint/js`, `typescript-eslint`, `eslint-plugin-react-hooks`, `eslint-plugin-react-refresh`. Script: `"lint": "eslint src/"`.
- [x] T009 [P] [US3] Create ESLint flat config in `frontend/eslint.config.js` with `@eslint/js` recommended rules, `typescript-eslint` recommended rules, `react-hooks` plugin, and `react-refresh` plugin. Scope to `src/` directory only.
- [x] T010 [US3] Add golangci-lint step to backend job in `.github/workflows/ci.yml` using `golangci/golangci-lint-action` with `working-directory: backend/`
- [x] T011 [US3] Add ESLint step (`npm run lint`) to frontend job in `.github/workflows/ci.yml`

**Checkpoint**: Full CI pipeline with all checks — build, test, lint, static analysis for both backend and frontend.

---

## Phase 4: Polish & Verification

**Purpose**: Fix pre-existing issues caught by new linters and verify the complete pipeline end-to-end.

- [ ] T012 [P] Run `golangci-lint run ./...` locally in `backend/` and fix any pre-existing violations found in `backend/` source files
- [ ] T013 [P] Run `npx eslint src/` locally in `frontend/` and fix any pre-existing violations found in `frontend/src/` source files
- [ ] T014 Push branch and open test PR to verify the complete CI pipeline runs both jobs successfully with all checks passing

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Core CI (Phase 2)**: Depends on Phase 1 (T001 must exist before adding steps)
- **Linting (Phase 3)**: T007-T009 can start in parallel with Phase 2 (different files). T010-T011 depend on Phase 2 (modify ci.yml after job structure is set)
- **Polish (Phase 4)**: T012 depends on T007 (needs golangci-lint config). T013 depends on T008-T009 (needs ESLint config + deps). T014 depends on all prior tasks.

### User Story Dependencies

- **US1 (PR triggers)**: Phase 1 + Phase 2 environment setup — no dependencies on other stories
- **US2 (unit tests)**: Depends on US1 job structure — T005 adds test step after T002 sets up backend job
- **US3+US4 (linting/analysis)**: Independent of US2 — linter configs (T007-T009) can be created in parallel with test steps
- **US5 (build)**: Depends on US1 job structure — T004/T006 add build steps after jobs exist

### Parallel Opportunities

```
Phase 1: T001 (sequential — single file)

Phase 2: T002 + T003 (parallel — different sections of ci.yml, but same file so sequential is safer)
         T004, T005, T006 (sequential — all modify ci.yml)

Phase 3: T007 + T008 + T009 (parallel — three different files)
         T010, T011 (sequential — both modify ci.yml)

Phase 4: T012 + T013 (parallel — different directories)
         T014 (sequential — depends on everything)
```

---

## Implementation Strategy

### MVP First (Phase 1 + Phase 2)

1. Complete T001: Skeleton workflow
2. Complete T002-T006: Working CI with build + tests
3. **STOP and VALIDATE**: Push and verify pipeline triggers, backend tests pass, frontend builds
4. This alone satisfies US1, US2, US5 and provides immediate value

### Full Delivery (Add Phase 3 + Phase 4)

5. Complete T007-T009: Linter configurations (parallel)
6. Complete T010-T011: Add lint steps to CI
7. Complete T012-T013: Fix pre-existing lint issues (parallel)
8. Complete T014: Final end-to-end verification via test PR

---

## Notes

- All tasks modify configuration files only — no application source code changes (except fixing lint violations in Phase 4)
- The `ci.yml` file is modified by many tasks sequentially — avoid parallel edits to this file
- T007, T008, T009 create independent config files and CAN run in parallel
- T012, T013 fix code in separate directories and CAN run in parallel
- Frontend has no test files — the CI pipeline runs `tsc --noEmit` and `vite build` as quality gates instead of unit tests
