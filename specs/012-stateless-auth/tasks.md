# Tasks: Stateless Authentication

**Input**: Design documents from `/specs/012-stateless-auth/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/auth-api.md, quickstart.md

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Install dependencies, create migration files, update configuration

- [ ] T001 Install `golang-jwt/jwt/v5` dependency in backend/go.mod
- [ ] T002 [P] Create database migration `backend/migrations/002_stateless_auth.up.sql`: create `refresh_tokens` table (id SERIAL PK, token_hash TEXT NOT NULL UNIQUE, user_type TEXT NOT NULL CHECK IN parent/child, user_id INTEGER NOT NULL, family_id INTEGER NOT NULL, expires_at TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ DEFAULT NOW()) with indexes on expires_at and (user_type, user_id); drop sessions table. Create matching `backend/migrations/002_stateless_auth.down.sql` rollback per data-model.md
- [ ] T003 [P] Update `backend/internal/config/config.go`: add `JWTSecret string` field loaded from `JWT_SECRET` env var (required, base64-encoded, decoded to []byte, minimum 32 bytes after decode — error at startup if missing or too short). Keep CookieDomain and CookieSecure fields for now (removed in Polish phase)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core JWT and refresh token infrastructure that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Create JWT utilities in `backend/internal/auth/jwt.go`: define `Claims` struct (UserType string, UserID int, FamilyID int + jwt.RegisteredClaims); implement `GenerateAccessToken(jwtKey []byte, userType string, userID int, familyID int) (string, error)` that creates a 15-minute HS256 JWT with sub="parent:123" or "child:456", user_type, user_id, family_id claims; implement `ValidateAccessToken(jwtKey []byte, tokenString string) (*Claims, error)` that parses and validates the JWT, enforces HS256 signing method to prevent algorithm confusion, returns claims or error
- [ ] T005 Write tests for JWT utilities in `backend/internal/auth/jwt_test.go`: test signing and validating a token round-trip; test expired token rejection; test tampered token rejection; test wrong signing method rejection; test claims are correctly extracted (user_type, user_id, family_id)
- [ ] T006 [P] Create refresh token store in `backend/internal/store/refresh_token.go`: define `RefreshToken` struct (ID, TokenHash, UserType, UserID, FamilyID, ExpiresAt, CreatedAt); implement `NewRefreshTokenStore(db *sql.DB)`; implement `Create(userType string, userID int, familyID int, ttl time.Duration) (rawToken string, err error)` that generates 32-byte random token, stores SHA-256 hash in DB, returns raw base64url token; implement `Validate(rawToken string) (*RefreshToken, error)` that hashes the token and queries DB checking expires_at > NOW(); implement `DeleteByHash(tokenHash string) error`; implement `DeleteExpired() error`; implement `DeleteByUser(userType string, userID int) error`; implement `StartCleanupLoop(interval time.Duration, stop chan struct{})` matching existing session cleanup pattern
- [ ] T007 [P] Write tests for refresh token store in `backend/internal/store/refresh_token_test.go`: test Create generates unique tokens and stores hash; test Validate succeeds for valid token; test Validate fails for expired token; test Validate fails for non-existent token; test DeleteByHash removes token; test DeleteExpired removes only expired tokens; test DeleteByUser removes all tokens for a user. Use testDB helper with TRUNCATE refresh_tokens RESTART IDENTITY CASCADE in setup
- [ ] T008 Update auth middleware in `backend/internal/middleware/auth.go`: change `RequireAuth` to extract token from `Authorization: Bearer <token>` header instead of `session` cookie; parse and validate JWT using `auth.ValidateAccessToken(jwtKey, token)`; set same context keys (ContextKeyUserType, ContextKeyUserID, ContextKeyFamilyID) from JWT claims; on invalid/expired token return 401 JSON error (no cookie clearing). Update `RequireAuth` signature to accept `jwtKey []byte` instead of `SessionValidator` interface. Update `RequireParent` accordingly. Remove `SessionValidator` interface
- [ ] T009 [P] Update CORS middleware in `backend/internal/middleware/cors.go`: change `CORS` function signature to `CORS(allowedOrigin string)` removing the `allowCredentials` parameter; remove `Access-Control-Allow-Credentials` header (no longer needed since auth is via Authorization header, not cookies)
- [ ] T010 Create frontend auth module in `frontend/src/auth.ts`: implement `getAccessToken(): string | null` reading from localStorage key `access_token`; implement `getRefreshToken(): string | null` reading from localStorage key `refresh_token`; implement `setTokens(accessToken: string, refreshToken: string): void` storing both to localStorage; implement `clearTokens(): void` removing both from localStorage; implement `isLoggedIn(): boolean` checking if access token exists (not expired)
- [ ] T011 Update frontend API client in `frontend/src/api.ts`: remove `credentials: "include"` from fetch calls; add `Authorization: Bearer ${getAccessToken()}` header to every request using the auth module; on 401 response, call `clearTokens()` and redirect to home page `/` (basic handling — auto-refresh added in US3)

**Checkpoint**: JWT signing/validation works, refresh token store passes all tests, middleware reads Bearer tokens, frontend sends Bearer tokens. Backend compilation requires updating handler constructors and main.go wiring (done in Phase 3).

---

## Phase 3: User Story 1 — Cross-Domain Parent Login via Google (Priority: P1)

**Goal**: Parent completes Google OAuth and receives JWT + refresh token. Frontend stores tokens and uses them for all subsequent API calls. Parent flow works end-to-end.

**Independent Test**: Complete Google OAuth login → verify tokens returned in redirect URL → verify GET /api/auth/me succeeds with Bearer token → verify page refresh preserves login.

### Implementation for User Story 1

- [ ] T012 [US1] Update `backend/internal/auth/google.go`: change GoogleAuth struct to replace `sessionStore *store.SessionStore` with `refreshTokenStore *store.RefreshTokenStore` and replace `cookieSecure bool` with `jwtKey []byte`; update `NewGoogleAuth` constructor; in `HandleCallback`: replace `sessionStore.Create()` + `http.SetCookie()` with `GenerateAccessToken(jwtKey, "parent", parent.ID, parent.FamilyID)` + `refreshTokenStore.Create("parent", parent.ID, parent.FamilyID, 7*24*time.Hour)`; redirect to `FRONTEND_URL/auth/callback?access_token=<jwt>&refresh_token=<token>&redirect=/setup` (new user) or `&redirect=/dashboard` (existing user). Keep oauth_state cookie logic unchanged (backend-only CSRF)
- [ ] T013 [US1] Update `backend/internal/family/handlers.go`: in `HandleCreateFamily`, replace cookie reading + `sessionStore.UpdateFamilyID()` with generating a new access token via `auth.GenerateAccessToken(jwtKey, "parent", parentID, newFamilyID)` and including `access_token` field in the JSON response. Add `jwtKey []byte` field to family Handlers struct and update constructor
- [ ] T014 [US1] Update `backend/main.go`: create `refreshTokenStore := store.NewRefreshTokenStore(db)`; decode JWT secret from config `cfg.JWTSecret`; pass `jwtKey` and `refreshTokenStore` to `auth.NewGoogleAuth`, `auth.NewChildAuth`, `auth.NewHandlers`, and `family.NewHandlers`; update `middleware.RequireAuth(jwtKey)` and `middleware.RequireParent(jwtKey)` calls; update `middleware.CORS(cfg.FrontendURL)` call (remove credentials bool); start `refreshTokenStore.StartCleanupLoop(1*time.Hour, stopCleanup)` replacing session cleanup; keep `sessionStore` temporarily (removed in Polish)
- [ ] T015 [US1] Update `frontend/src/pages/GoogleCallback.tsx`: instead of calling GET /api/auth/me immediately, extract `access_token`, `refresh_token`, and `redirect` from URL query parameters using `URLSearchParams`; call `setTokens(accessToken, refreshToken)` from auth module; clear tokens from URL using `window.history.replaceState`; navigate to the `redirect` param value (`/setup` or `/dashboard`)
- [ ] T016 [P] [US1] Update `frontend/src/pages/SetupPage.tsx`: after successful POST /api/families response, read `access_token` from response JSON and call `setTokens(newAccessToken, getRefreshToken())` to store the updated token with new family_id
- [ ] T017 [P] [US1] Update `frontend/src/pages/ParentDashboard.tsx`: no functional changes needed if api.ts already sends Bearer token. Verify GET /api/auth/me works with Bearer token and auth redirect on 401 works correctly

**Checkpoint**: Parent can log in via Google OAuth, receive tokens, access the dashboard, create a family, and remain logged in across page refreshes.

---

## Phase 4: User Story 2 — Cross-Domain Child Login (Priority: P1)

**Goal**: Child logs in with credentials and receives JWT + refresh token. Child flow works end-to-end with Bearer token auth.

**Independent Test**: Submit child credentials → verify response contains access_token + refresh_token + user info → verify GET /api/auth/me succeeds → verify failed login increments attempts and lockout works.

### Implementation for User Story 2

- [ ] T018 [US2] Update `backend/internal/auth/child.go`: change ChildAuth struct to replace `sessionStore *store.SessionStore` with `refreshTokenStore *store.RefreshTokenStore` and replace `cookieSecure bool` with `jwtKey []byte`; update `NewChildAuth` constructor; in `HandleChildLogin`: replace `sessionStore.Create()` + `http.SetCookie()` with `GenerateAccessToken(jwtKey, "child", child.ID, fam.ID)` + `refreshTokenStore.Create("child", child.ID, fam.ID, 24*time.Hour)`; return JSON `{"access_token": "...", "refresh_token": "...", "user": {"user_type": "child", "user_id": ..., "family_id": ..., "first_name": "...", "family_slug": "..."}}` — note user info is now nested under `user` key per contracts
- [ ] T019 [US2] Update `frontend/src/pages/FamilyLogin.tsx`: after successful POST /api/auth/child/login, extract `access_token` and `refresh_token` from response JSON; call `setTokens(accessToken, refreshToken)` from auth module; navigate to `/child/dashboard`
- [ ] T020 [P] [US2] Update `frontend/src/pages/ChildDashboard.tsx`: no functional changes needed if api.ts already sends Bearer token. Verify GET /api/auth/me works with Bearer token and auth redirect on 401 works correctly

**Checkpoint**: Child can log in with credentials, receive tokens, and access the child dashboard. Failed logins and account lockout still work with audit logging.

---

## Phase 5: User Story 3 — Token Refresh Without Re-Login (Priority: P2)

**Goal**: Users can silently refresh their tokens before expiry. Active sessions extend automatically without user intervention.

**Independent Test**: Authenticate → wait for access token to near expiry → call POST /api/auth/refresh with refresh token → verify new access_token + refresh_token returned → verify old refresh token is invalidated (rotation).

### Implementation for User Story 3

- [ ] T021 [US3] Implement `HandleRefresh` method on Handlers struct in `backend/internal/auth/handlers.go`: parse `{"refresh_token": "..."}` from request body (no Authorization header required); call `refreshTokenStore.Validate(rawToken)` to verify the refresh token is valid and not expired; if invalid, return 401 `{"error": "Invalid or expired refresh token"}`; if valid, delete old refresh token via `refreshTokenStore.DeleteByHash(refreshToken.TokenHash)`; look up current user data from parentStore or childStore based on refreshToken.UserType/UserID to get current family_id; create new refresh token via `refreshTokenStore.Create(userType, userID, familyID, ttl)` with 7-day (parent) or 24-hour (child) TTL; generate new access token via `GenerateAccessToken(jwtKey, userType, userID, familyID)`; return `{"access_token": "...", "refresh_token": "..."}`
- [ ] T022 [US3] Register POST /api/auth/refresh route in `backend/main.go`: add `mux.Handle("POST /api/auth/refresh", refreshRateLimit(http.HandlerFunc(authHandlers.HandleRefresh)))` using a rate limiter of 10 req/min per IP (same as child login). This route must NOT require auth middleware (access token may be expired)
- [ ] T023 [US3] Add automatic token refresh logic in `frontend/src/auth.ts` and `frontend/src/api.ts`: in auth.ts, add `refreshTokens()` async function that calls POST /api/auth/refresh with the stored refresh token and updates localStorage with new tokens; in api.ts, update the 401 handler: before redirecting to login, attempt to call `refreshTokens()` — if refresh succeeds, retry the original request with the new access token; if refresh fails (401), then clear tokens and redirect to login. Add a flag to prevent concurrent refresh attempts (e.g., a promise-based lock)

**Checkpoint**: Users with valid refresh tokens can silently extend their sessions. Access token expiry is transparent to the user.

---

## Phase 6: User Story 4 — Logout Invalidation (Priority: P2)

**Goal**: Logging out immediately revokes the refresh token so no new access tokens can be obtained.

**Independent Test**: Log in → obtain tokens → log out (send refresh token in body) → attempt POST /api/auth/refresh with old refresh token → verify 401.

### Implementation for User Story 4

- [ ] T024 [US4] Update `HandleLogout` in `backend/internal/auth/handlers.go`: parse `{"refresh_token": "..."}` from request body; compute SHA-256 hash; call `refreshTokenStore.DeleteByHash(hash)` to revoke it; log `logout` audit event via eventStore; return `{"message": "Logged out successfully"}`. Remove all cookie-clearing logic (no more http.SetCookie with MaxAge=-1)
- [ ] T025 [US4] Update `frontend/src/components/Layout.tsx`: in `handleLogout`, send POST /api/auth/logout with `{"refresh_token": getRefreshToken()}` in request body; after response (success or failure), call `clearTokens()` to remove both tokens from localStorage; then navigate to home page

**Checkpoint**: Logout revokes the refresh token server-side and clears client-side tokens. Old refresh tokens cannot be used to obtain new access tokens.

---

## Phase 7: Polish & Cleanup

**Purpose**: Remove legacy code, update deployment config, final validation

- [ ] T026 Delete `backend/internal/store/session.go` and `backend/internal/store/session_test.go` — replaced by refresh_token.go and jwt.go
- [ ] T027 [P] Remove `CookieDomain` and `CookieSecure` fields from config struct in `backend/internal/config/config.go`; remove `COOKIE_DOMAIN` and `COOKIE_SECURE` env vars from `docker-compose.yaml`; add `JWT_SECRET` env var to docker-compose.yaml backend service
- [ ] T028 [P] Remove any remaining `sessionStore` references from `backend/main.go` (session store creation, cleanup goroutine stop channel if session-specific, passing to constructors). Verify no imports of session store remain
- [ ] T029 Run database migration on test DB and execute full backend test suite with `cd backend && go test -p 1 ./...` — fix any failures
- [ ] T030 Validate all auth flows end-to-end per quickstart.md: parent Google OAuth login → dashboard; child credential login → dashboard; token refresh after access token expiry; logout → tokens revoked; page refresh preserves login; cross-origin requests work with Authorization header

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — first story to make backend compilable with new auth
- **US2 (Phase 4)**: Depends on Phase 3 (main.go wiring must be done first)
- **US3 (Phase 5)**: Depends on Phase 3 (needs working JWT + refresh token infrastructure in main.go)
- **US4 (Phase 6)**: Depends on Phase 3 (needs Handlers struct with refreshTokenStore)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Must be implemented first — includes main.go wiring that all other stories depend on
- **US2 (P1)**: Can start after US1 completes (shares main.go wiring)
- **US3 (P2)**: Can start after US1 completes (independent of US2)
- **US4 (P2)**: Can start after US1 completes (independent of US2 and US3)
- **US3 and US4**: Can be developed in parallel after US1

### Within Each User Story

- Backend changes before frontend changes (frontend depends on API contract)
- main.go route registration after handler implementation

### Parallel Opportunities

- T002 + T003 (Setup phase: migration files + config)
- T006/T007 + T004/T005 (Foundational: refresh token store + JWT utilities)
- T009 (CORS update, independent)
- T010 (frontend auth module, independent of backend)
- T016 + T017 (US1: SetupPage + ParentDashboard, different files)
- T019 + T020 (US2: FamilyLogin + ChildDashboard, different files — but T019 is quick)
- US3 + US4 (can be developed in parallel after US1)
- T027 + T028 + T029 (Polish: independent cleanup tasks)

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: US1 (parent login)
4. Complete Phase 4: US2 (child login)
5. **STOP and VALIDATE**: Both login flows work with Bearer tokens cross-domain
6. Deploy if ready — users can log in and use the app

### Incremental Delivery

1. Setup + Foundational → Infrastructure ready
2. Add US1 → Parent login works → Test independently
3. Add US2 → Child login works → Test independently → **MVP deployed**
4. Add US3 → Token refresh works → Sessions extend silently
5. Add US4 → Logout invalidation works → Full security
6. Polish → Clean codebase, all tests passing

---

## Notes

- [P] tasks = different files, no dependencies on each other
- [Story] label maps task to specific user story (US1-US4)
- US1 is the critical path — it includes main.go wiring that all other stories depend on
- Backend changes within a story should be done before frontend changes for that story
- The migration is a clean cutover: all existing sessions are invalidated on deploy
- `go test -p 1` required because all packages share the test database
- JWT_SECRET must be set for tests — generate with `openssl rand -base64 64`
