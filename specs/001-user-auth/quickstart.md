# Quickstart: User Authentication

**Feature**: 001-user-auth
**Purpose**: Manual test scenarios to validate each user story end-to-end

## Prerequisites

1. Application running via `docker compose up --build`
2. Google OAuth credentials configured (client ID + secret in environment)
3. Frontend accessible at `http://localhost:8000`
4. Backend API at `http://localhost:8001`

## Test Scenario 1: Parent Registration (US1)

**Goal**: Verify a new parent can register via Google and create a family bank URL.

1. Navigate to `http://localhost:8000`
2. Click "Sign in with Google"
3. Authenticate with a Google account
4. **Verify**: You are prompted to choose a family bank URL slug
5. Enter `test-family` as the slug
6. **Verify**: Family bank is created, you are redirected to the parent dashboard
7. **Verify**: Dashboard shows your Google display name and family URL

### Slug Validation

8. Open a new browser/incognito, register a second parent
9. Try to claim `test-family` as the slug
10. **Verify**: Error message appears saying slug is taken, with suggestions

### Duplicate Registration

11. Log out
12. Click "Sign in with Google" with the same Google account
13. **Verify**: You are logged in to your existing account (not prompted to register again)

## Test Scenario 2: Parent Login (US2)

**Goal**: Verify a registered parent can log in and out.

1. Navigate to `http://localhost:8000` (logged out)
2. Click "Sign in with Google"
3. Authenticate with the previously registered Google account
4. **Verify**: Redirected to parent dashboard
5. **Verify**: Dashboard shows correct family info
6. Click "Log out"
7. **Verify**: Redirected to home page
8. **Verify**: Cannot access dashboard without logging in again

### Unregistered Login

9. Click "Sign in with Google" with a new Google account that has never registered
10. **Verify**: Redirected to slug selection (registration flow)

## Test Scenario 3: Create Child Account (US3)

**Goal**: Verify a parent can create a child account.

1. Log in as a parent
2. Click "Add Child"
3. Enter first name: `Tommy`, password: `secret123`
4. **Verify**: Child account created successfully
5. **Verify**: Login URL and credentials are displayed for sharing with the child

### Password Validation

6. Try creating another child with password: `abc` (too short)
7. **Verify**: Error message about minimum 6 characters

### Unique Name Validation

8. Try creating another child named `Tommy`
9. **Verify**: Error message about duplicate name in family

## Test Scenario 4: Child Login (US4)

**Goal**: Verify a child can log in via the family bank URL.

1. Open a new browser/incognito window
2. Navigate to `http://localhost:8000/test-family`
3. **Verify**: Family login page is shown
4. Enter first name: `Tommy`, password: `secret123`
5. **Verify**: Logged in, child dashboard shown with account balance
6. Click "Log out"
7. **Verify**: Returned to family bank login page

### Wrong Credentials

8. Enter first name: `Tommy`, password: `wrongpassword`
9. **Verify**: Friendly error message shown

### Account Lockout

10. Enter wrong password 4 more times (total 5 failures)
11. **Verify**: Account locked message shown
12. **Verify**: Parent receives notification (check parent dashboard or logs)

### Non-existent Family URL

13. Navigate to `http://localhost:8000/nonexistent-family`
14. **Verify**: 404 page shown with link to create your own bank

## Test Scenario 5: Manage Child Credentials (US5)

**Goal**: Verify a parent can reset a child's password and update their name.

### Password Reset

1. Log in as the parent
2. Select child `Tommy`
3. Click "Reset Password"
4. Enter new password: `newpass789`
5. **Verify**: Password updated confirmation
6. **Verify**: If account was locked, it is now unlocked
7. Open new browser, navigate to family URL
8. Log in as Tommy with old password `secret123`
9. **Verify**: Login fails
10. Log in with new password `newpass789`
11. **Verify**: Login succeeds

### Name Update

12. Log in as the parent
13. Select child `Tommy`
14. Update name to `Thomas`
15. **Verify**: Name changed throughout the application
16. **Verify**: Child can log in with name `Thomas`

## Test Scenario 6: Session Persistence (US6)

**Goal**: Verify sessions persist across browser restarts.

### Parent Session (7-day TTL)

1. Log in as a parent
2. Close the browser completely
3. Reopen and navigate to `http://localhost:8000`
4. **Verify**: Still logged in, dashboard accessible without re-authenticating

### Child Session (24-hour TTL)

5. Log in as a child
6. Close the browser completely
7. Reopen and navigate to the family URL
8. **Verify**: Still logged in (within 24 hours)

### Session Expiry

9. (Requires adjusting session TTL for testing or waiting)
10. After session expires, attempt to access dashboard
11. **Verify**: Redirected to appropriate login page

## Security Checks

### Cross-Family Access

1. Create two families with children
2. Log in as child in Family A
3. Try navigating to Family B's URL
4. **Verify**: Cannot access Family B's data

### Auth Event Logging

1. Perform various login/logout actions
2. Check server logs
3. **Verify**: All auth events logged (login success, failure, logout, account creation)
4. **Verify**: No passwords or session tokens appear in logs
