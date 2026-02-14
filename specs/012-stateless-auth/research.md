# Research: Stateless Authentication

**Feature**: 012-stateless-auth
**Date**: 2026-02-14

## Decision 1: JWT Library

**Decision**: Use `golang-jwt/jwt/v5`

**Rationale**: Community-standard Go JWT library with 13,400+ importers. Actively maintained (latest release Jan 2026). Clean v5 API with `ParserOption` functions and improved `Claims` interface. Widely documented and battle-tested.

**Alternatives considered**:
- `lestrrat-go/jwx`: Full JOSE suite — overkill for our single-backend use case
- `kataras/jwt`: Far less adoption (~200 importers), smaller community

## Decision 2: Signing Algorithm

**Decision**: HMAC-SHA256 (HS256) with symmetric key

**Rationale**: Single-backend application where the same server issues and validates tokens. Symmetric signing is simpler (one shared secret vs key pair), faster, and has no key distribution problem since the secret never leaves the process.

**Alternatives considered**:
- RS256/ES256 (asymmetric): Appropriate for microservices where separate services verify tokens with a public key. Unnecessary complexity for a monolith.

## Decision 3: Secret Key Management

**Decision**: 64-byte cryptographically random key, base64-encoded, loaded from `JWT_SECRET` environment variable. Minimum 32 bytes enforced at startup.

**Rationale**: Environment variable is the standard approach for containerized deployments (Docker, Fly.io). 64 bytes exceeds the RFC 7518 minimum of 32 bytes for HS256, providing a security margin. Base64 encoding allows safe storage in env vars.

**Key generation**: `openssl rand -base64 64`

## Decision 4: Token Architecture

**Decision**: Short-lived access tokens (15 min JWT) + database-tracked refresh tokens (opaque, 7-day parent / 24-hour child)

**Rationale**: This satisfies both FR-003 (stateless validation on every request) and FR-005 (server-side revocation on logout):
- **Access token**: JWT validated without DB lookup (stateless). 15-minute lifetime limits exposure.
- **Refresh token**: Opaque random string, SHA-256 hash stored in `refresh_tokens` DB table. DB lookup only on refresh (~once per 15 min), not on every request. Deleted on logout for immediate revocation.

**Alternatives considered**:
- Token blacklist: Requires DB check on every request — defeats stateless purpose
- Single long-lived JWT with no refresh: No way to revoke on logout
- JTI claim tracking: Equivalent to blacklist; DB check on every request

**Tradeoff**: After logout, the access token remains valid for up to 15 minutes. For a family banking app, this is acceptable. The refresh token is immediately invalidated, so no new access tokens can be obtained.

## Decision 5: OAuth Callback Token Delivery

**Decision**: Backend redirects to `FRONTEND_URL/auth/callback?access_token=<jwt>&refresh_token=<opaque>` after Google OAuth completion.

**Rationale**: Simplest cross-domain approach. The tokens briefly appear in the URL but:
- HTTPS encrypts the URL in transit
- The frontend extracts tokens immediately and replaces the URL (no browser history leak)
- Access token is short-lived (15 min), limiting exposure
- The SPA doesn't navigate to external URLs from the callback page (no Referrer leak)

**Alternatives considered**:
- HttpOnly cookie for refresh token: Not reliable cross-domain (Safari blocks third-party cookies, Chrome restricting them). Contradicts the cross-domain requirement.
- One-time code exchange: More secure but adds another DB table, endpoint, and round-trip. Overkill for this app's threat model.
- URL fragment (#): Client-side only — cannot be read by a server-rendered page, but works for SPAs. Adds complexity with hash routing.

## Decision 6: Frontend Token Storage

**Decision**: Both access token and refresh token stored in `localStorage`.

**Rationale**: Cross-domain requirement eliminates HttpOnly cookies as an option (third-party cookie restrictions). localStorage survives page refresh (FR-013) and tab close, meeting the 7-day parent / 24-hour child session requirements (FR-008). In-memory storage would lose tokens on refresh.

**XSS mitigation**: localStorage is vulnerable to XSS. Mitigations:
- Access token is short-lived (15 min) — stolen token has limited utility
- Content Security Policy headers
- React's built-in XSS protection (JSX escaping)
- No third-party scripts loaded

**Alternatives considered**:
- In-memory (JS variable): Loses tokens on page refresh — users would re-login constantly
- sessionStorage: Doesn't survive tab close — 7-day parent sessions impossible
- HttpOnly cookie: Not viable cross-domain

## Decision 7: Refresh Token Rotation

**Decision**: Rotate refresh tokens on every refresh — issue a new refresh token and invalidate the old one.

**Rationale**: Limits the window of exposure if a refresh token is stolen. If someone tries to use an already-rotated refresh token, it signals compromise.

## Decision 8: Family ID Updates After Family Creation

**Decision**: The family creation endpoint (`POST /api/families`) returns a new access token with the updated `family_id` claim.

**Rationale**: JWT claims are set at issuance time. When a parent creates a family during setup, their original access token has `family_id=0`. Rather than requiring a separate refresh call, the family creation response includes a new access token. Simplest user experience.

## Decision 9: OAuth State Cookie

**Decision**: Keep the existing `oauth_state` cookie for CSRF protection during OAuth flow. No change needed.

**Rationale**: The oauth_state cookie is set and read by the backend on the same domain (`api.example.com`). It's a short-lived (10 min) cookie between the browser and the backend only — not cross-domain. The OAuth flow is: browser → backend → Google → backend → frontend. The cookie lives on the backend domain throughout.
