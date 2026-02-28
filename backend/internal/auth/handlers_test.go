package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"bank-of-dad/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultTestDSN = "postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable"

func setupAuthTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}
	db, err := store.Open(dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Exec(`TRUNCATE stripe_webhook_events, interest_schedules, transactions, allowance_schedules, auth_events, refresh_tokens, children, parents, families RESTART IDENTITY CASCADE`)
		db.Close()
	})
	_, err = db.Exec(`TRUNCATE stripe_webhook_events, interest_schedules, transactions, allowance_schedules, auth_events, refresh_tokens, children, parents, families RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
	return db
}

func setAuthRequestContext(r *http.Request, userType string, userID, familyID int64) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, ContextKeyUserType, userType)
	ctx = context.WithValue(ctx, ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, ContextKeyFamilyID, familyID)
	return r.WithContext(ctx)
}

func newTestAuthHandlers(t *testing.T) (*Handlers, *store.FamilyStore, *store.ChildStore) {
	t.Helper()
	db := setupAuthTestDB(t)
	parentStore := store.NewParentStore(db)
	familyStore := store.NewFamilyStore(db)
	childStore := store.NewChildStore(db)
	refreshTokenStore := store.NewRefreshTokenStore(db)
	eventStore := store.NewAuthEventStore(db)
	h := NewHandlers(parentStore, familyStore, childStore, refreshTokenStore, eventStore, []byte("test-key"))
	return h, familyStore, childStore
}

// T007: HandleGetMe includes theme field for child users
func TestHandleGetMe_ChildIncludesTheme(t *testing.T) {
	h, familyStore, childStore := newTestAuthHandlers(t)

	fam, err := familyStore.Create("test-family")
	require.NoError(t, err)

	child, err := childStore.Create(fam.ID, "Alice", "password123", nil)
	require.NoError(t, err)

	err = childStore.UpdateTheme(child.ID, "sparkle")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	req = setAuthRequestContext(req, "child", child.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleGetMe(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "child", resp["user_type"])
	assert.Equal(t, "Alice", resp["first_name"])
	assert.Equal(t, "sparkle", resp["theme"])
	assert.Equal(t, "test-family", resp["family_slug"])
}

func TestHandleGetMe_ChildNilTheme(t *testing.T) {
	h, familyStore, childStore := newTestAuthHandlers(t)

	fam, err := familyStore.Create("test-family")
	require.NoError(t, err)

	child, err := childStore.Create(fam.ID, "Bob", "password123", nil)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	req = setAuthRequestContext(req, "child", child.ID, fam.ID)
	rr := httptest.NewRecorder()

	h.HandleGetMe(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "child", resp["user_type"])
	// theme should be null when not set
	assert.Nil(t, resp["theme"])
}
