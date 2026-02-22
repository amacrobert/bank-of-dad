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
