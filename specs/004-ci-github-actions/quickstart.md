# Quickstart: CI Pipeline

## Overview

This feature adds a GitHub Actions CI pipeline that automatically validates code quality on every pull request. No application code changes are needed — only configuration files.

## Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `.github/workflows/ci.yml` | Create | GitHub Actions workflow definition |
| `backend/.golangci.yml` | Create | Go linter configuration |
| `frontend/eslint.config.js` | Create | TypeScript/React linter configuration |
| `frontend/package.json` | Modify | Add `lint` script and ESLint dev dependencies |

## Verification

After implementation, open a PR and verify:

1. Two jobs appear in the PR checks: `backend` and `frontend`
2. Backend job runs: build, lint, tests
3. Frontend job runs: lint, type check, build
4. Both jobs complete within 10 minutes
5. Push another commit — verify the previous in-progress run is cancelled

## Key Commands

```bash
# Backend checks (run locally)
cd backend && go build ./... && go test ./...

# Frontend checks (run locally)
cd frontend && npm ci && npx eslint src/ && npx tsc --noEmit && npx vite build

# golangci-lint locally (after installing)
cd backend && golangci-lint run ./...
```

## Notes

- No data model or API contracts needed — this feature is purely CI configuration
- No frontend test framework is added (no test files exist yet)
- The `contracts/` directory is not created for this feature as there are no API changes
