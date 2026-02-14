# Feature Specification: Stateless Authentication

**Feature Branch**: `012-stateless-auth`
**Created**: 2026-02-14
**Status**: Draft
**Input**: User description: "Replace cookie authentication with stateless authentication. In production, the backend and the frontend are not necessarily going to have the same domain."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cross-Domain Parent Login via Google (Priority: P1)

A parent visits the frontend application (hosted on one domain) and clicks "Sign in with Google." After completing Google sign-in, they are redirected back and receive an authentication token. On all subsequent page loads and API calls, the frontend sends this token to the backend (hosted on a different domain) to prove the parent's identity. The parent remains logged in across page refreshes until the token expires or they log out.

**Why this priority**: This is the primary authentication flow. Without cross-domain token-based auth working for parent login, the application cannot function in a production environment where frontend and backend are on separate domains.

**Independent Test**: Can be fully tested by completing a Google OAuth login, verifying a token is returned, and confirming the token grants access to authenticated parent endpoints from a different origin.

**Acceptance Scenarios**:

1. **Given** a parent is not logged in, **When** they complete Google OAuth sign-in, **Then** the system returns an authentication token that the frontend can store and use for subsequent requests.
2. **Given** a parent has a valid token, **When** they make an API request with the token in the Authorization header, **Then** the backend validates the token and processes the request as that parent.
3. **Given** a parent has a valid token and refreshes the browser, **When** the page reloads, **Then** the frontend retrieves the stored token and the parent remains authenticated without re-logging in.
4. **Given** a parent's token has expired, **When** they make an API request, **Then** the backend returns a 401 response and the frontend redirects to the login page.

---

### User Story 2 - Cross-Domain Child Login (Priority: P1)

A child visits the frontend and enters their family slug, first name, and password. Upon successful login, they receive an authentication token. The child uses the app normally, with the frontend sending the token on every request to the backend on its separate domain.

**Why this priority**: Child login is equally critical to parent login. Both must work cross-domain for the app to be usable in production.

**Independent Test**: Can be fully tested by submitting child credentials, verifying a token is returned, and confirming the token grants access to authenticated child endpoints.

**Acceptance Scenarios**:

1. **Given** a child is not logged in, **When** they submit valid credentials (family slug, first name, password), **Then** the system returns an authentication token.
2. **Given** a child has a valid token, **When** they make an API request with the token in the Authorization header, **Then** the backend identifies them correctly and processes the request.
3. **Given** a child submits invalid credentials, **When** the login request is made, **Then** the system returns an error without issuing a token and the failed attempt is recorded.
4. **Given** a child's account is locked after 5 failed attempts, **When** they try to log in, **Then** the system rejects the request without issuing a token.

---

### User Story 3 - Token Refresh Without Re-Login (Priority: P2)

A logged-in user (parent or child) continues using the application over an extended period. Before their current token expires, the system provides a way to obtain a fresh token so the user does not get unexpectedly logged out mid-session.

**Why this priority**: Without token refresh, users would be abruptly logged out when their token expires, resulting in a poor experience. This is important but secondary to basic login working.

**Independent Test**: Can be tested by authenticating, waiting until near token expiry, requesting a token refresh, and confirming the new token is valid.

**Acceptance Scenarios**:

1. **Given** a user has a valid (non-expired) token, **When** they request a token refresh, **Then** the system issues a new token with a fresh expiration and the old token remains valid until its original expiry.
2. **Given** a user has an expired token, **When** they request a token refresh, **Then** the system rejects the request with a 401 response.
3. **Given** a user's token is within a configurable window of expiry, **When** the frontend detects this, **Then** the frontend automatically requests a new token without interrupting the user's workflow.

---

### User Story 4 - Logout Invalidation (Priority: P2)

A user clicks "Log out." Their current token is invalidated so it can no longer be used, even if it has not yet expired. This ensures that logging out is a definitive action.

**Why this priority**: Logout must actually revoke access. Without server-side invalidation, a stolen token could continue to be used after the user believes they've logged out.

**Independent Test**: Can be tested by logging in, obtaining a token, logging out, and confirming the token is rejected on subsequent requests.

**Acceptance Scenarios**:

1. **Given** a user is logged in with a valid token, **When** they log out, **Then** the token is invalidated and any subsequent request using that token returns 401.
2. **Given** a user logs out on one device, **When** they use the same token from another device, **Then** the request is rejected.

---

### Edge Cases

- What happens when a user sends a request with a malformed or tampered token? The system rejects it with a 401 response.
- What happens when a user sends a request with no token at all to a protected endpoint? The system returns a 401 response.
- What happens if the token contains a user ID for a user that no longer exists? The system returns a 401 response.
- What happens during the migration period if a request arrives with an old-style session cookie but no token? The system rejects it with 401 — the migration is a clean cutover, not a gradual transition.
- What happens if the frontend and backend clocks are slightly out of sync? Token expiry should have enough tolerance to handle minor clock skew (up to a few minutes).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST authenticate users via stateless tokens sent in the HTTP Authorization header, replacing the current cookie-based session mechanism.
- **FR-002**: System MUST issue a token upon successful Google OAuth login (parent) or credential-based login (child).
- **FR-003**: System MUST validate tokens on every authenticated request without requiring a database lookup for standard validation (stateless verification).
- **FR-004**: System MUST support token refresh so users can extend their session without re-authenticating.
- **FR-005**: System MUST support server-side token revocation on logout, ensuring logged-out tokens cannot be reused.
- **FR-006**: System MUST work when the frontend and backend are hosted on different domains (no same-origin requirement).
- **FR-007**: System MUST preserve existing user identity information in the token: user type (parent/child), user ID, and family ID.
- **FR-008**: System MUST maintain existing session durations — 7 days for parents, 24 hours for children.
- **FR-009**: System MUST continue to enforce the existing child account lockout policy (5 failed attempts).
- **FR-010**: System MUST continue to log authentication events (login, logout, failure, lockout) to the audit trail.
- **FR-011**: Frontend MUST store the token client-side and include it in the Authorization header of every API request.
- **FR-012**: Frontend MUST handle token expiry gracefully by redirecting the user to the login page when a 401 response is received.
- **FR-013**: Frontend MUST persist the token across page refreshes so users are not logged out on reload.
- **FR-014**: System MUST remove all cookie-based session logic (session cookie setting, reading, clearing) after migration.
- **FR-015**: CORS configuration MUST allow the frontend origin to make authenticated requests with the Authorization header to the backend on a different domain.

### Key Entities

- **Authentication Token**: A self-contained credential issued by the backend upon successful login. Contains user identity (user type, user ID, family ID), issuance time, and expiration time. Sent by the frontend in the Authorization header on every request.
- **Token Revocation Record**: A server-side record of tokens that have been explicitly invalidated (e.g., via logout) before their natural expiry. Checked during token validation to enforce logout.

## Assumptions

- The existing Google OAuth flow (redirect-based) will continue to work, but the final step will return a token instead of setting a cookie. The redirect after OAuth callback will deliver the token to the frontend via a URL parameter or similar mechanism suitable for cross-domain use.
- The existing child login endpoint will return the token in the response body instead of setting a cookie.
- The frontend will store the token in browser storage (e.g., localStorage or sessionStorage) rather than relying on cookies.
- Token revocation (for logout) will require minimal server-side storage — only tracking actively revoked tokens until their natural expiry, not all issued tokens.
- The migration is a clean cutover: all existing sessions will be invalidated when this feature is deployed. Users will need to log in again.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully log in and use the application when the frontend and backend are hosted on different domains.
- **SC-002**: Authenticated API requests complete without requiring cookies — the Authorization header is the sole authentication mechanism.
- **SC-003**: Users remain logged in across page refreshes for the duration of their token's validity (7 days for parents, 24 hours for children).
- **SC-004**: Logged-out tokens are rejected within seconds of the logout action.
- **SC-005**: All existing authentication features continue to work: Google OAuth login, child username/password login, account lockout, and audit logging.
- **SC-006**: Token refresh allows active users to maintain their session beyond the initial token lifetime without re-entering credentials.
