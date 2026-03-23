# bank-of-dad Development Guidelines

## Tech Stack
- Backend: Go 1.24, PostgreSQL 17 (pgx/v5), golang-migrate, testify
- Frontend: TypeScript 5.3.3, React 18.2.0, Vite, Tailwind CSS 4, Recharts, lucide-react
- Auth: JWT access tokens (15-min, HS256) + opaque refresh tokens (DB-stored)
- OAuth: Google sign-in for parents, password auth for children

## Project Structure
```
backend/
├── main.go                  # HTTP server, route definitions
├── internal/
│   ├── auth/                # JWT, Google OAuth, child login
│   ├── allowance/           # Allowance scheduling + processor
│   ├── interest/            # Interest scheduling + accrual
│   ├── balance/             # Deposits, withdrawals, balances
│   ├── family/              # Family + child CRUD
│   ├── settings/            # Timezone, parent prefs
│   ├── middleware/          # JWT auth, CORS, logging, rate limiting
│   ├── store/               # PostgreSQL data access layer
│   ├── config/              # Env var loading
│   └── testutil/            # Shared test helpers
├── migrations/              # SQL migrations (001–005)
└── go.mod

frontend/
├── src/
│   ├── pages/               # Route page components
│   ├── components/          # Reusable UI components
│   │   └── ui/              # Base components (Button, Input, Card, etc.)
│   ├── context/             # ThemeContext, TimezoneContext
│   ├── hooks/               # Custom React hooks
│   ├── utils/               # Utilities (projection, familyPreference)
│   ├── api.ts               # Fetch wrapper with JWT injection + auto-refresh
│   ├── auth.ts              # Token storage (localStorage)
│   ├── themes.ts            # Child theme definitions
│   ├── types.ts             # TypeScript interfaces
│   └── App.tsx              # Routes
└── package.json

specs/                        # Feature specifications (001–020)
```

## Commands
```bash
# Backend tests (must cd into backend/, -p 1 required for shared test DB)
cd backend && go test -p 1 ./...

# Frontend type check + build
cd frontend && npx tsc --noEmit && npm run build

# Frontend lint
cd frontend && npm run lint
```

## Key Patterns
- Store: `NewXxxStore(db *sql.DB)`, PostgreSQL `$1`/`$2` placeholders
- Handler: struct with store deps, `writeJSON` helper, `ErrorResponse` type
- Routes: `mux.Handle("METHOD /path", middleware(handler.Method))`
- Money: int64 cents throughout
- Migrations: golang-migrate, `backend/migrations/`
- Auth context: helpers in `auth` package, re-exported by `middleware`
- Schedulers: goroutine + `time.NewTicker` + stop channel

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->

## Active Technologies
- Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `github.com/stripe/stripe-go/v82` (new), pgx/v5 (existing), Vite + Tailwind CSS 4 (existing) (024-stripe-subscription)
- PostgreSQL 17 — extend `families` table, add `stripe_webhook_events` table (024-stripe-subscription)
- Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + pgx/v5 (existing), Vite + Tailwind CSS 4 (existing), lucide-react (existing), Recharts (existing) (025-savings-goals)
- PostgreSQL 17 — 2 new tables (`savings_goals`, `goal_allocations`), no changes to existing tables (025-savings-goals)
- Go 1.24 (backend, unchanged), TypeScript 5.3.3 + React 18.2.0 (frontend) + Vite, Tailwind CSS 4, lucide-react (all existing) (026-child-auto-setup)
- PostgreSQL 17 (existing, no schema changes) (026-child-auto-setup)
- Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + pgx/v5 (DB), stripe-go/v82 (webhooks), Vite + Tailwind CSS 4, lucide-reac (028-child-account-limits)
- PostgreSQL 17 — 1 new column (`is_disabled`) on existing `children` table (028-child-account-limits)
- Go 1.24 + GORM (`gorm.io/gorm`), GORM PostgreSQL driver (`gorm.io/driver/postgres`), existing `golang-migrate/migrate/v4` (retained), `jackc/pgx/v5` (retained as underlying driver) (029-gorm-backend-refactor)
- PostgreSQL 17 — 12 tables, schema managed by `golang-migrate` (unchanged) (029-gorm-backend-refactor)
- Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + GORM (gorm.io/gorm), pgx/v5 (driver), Vite, Tailwind CSS 4, lucide-reac (031-chore-system)
- PostgreSQL 17 — 3 new tables (`chores`, `chore_assignments`, `chore_instances`), 1 new transaction type (031-chore-system)

## Recent Changes
- 024-stripe-subscription: Added Go 1.24 (backend), TypeScript 5.3.3 + React 18.2.0 (frontend) + `github.com/stripe/stripe-go/v82` (new), pgx/v5 (existing), Vite + Tailwind CSS 4 (existing)
