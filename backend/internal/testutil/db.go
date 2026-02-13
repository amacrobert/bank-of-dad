package testutil

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"testing"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/store"

	"github.com/stretchr/testify/require"
)

const defaultTestDSN = "postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable"

// SetupTestDB opens a connection to the test Postgres database, runs migrations,
// and registers cleanup that truncates all tables after the test completes.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	db, err := store.Open(dsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		// Truncate all tables in dependency order
		_, err := db.Exec(`TRUNCATE interest_schedules, transactions, allowance_schedules, auth_events, sessions, children, parents, families RESTART IDENTITY CASCADE`)
		if err != nil {
			t.Logf("cleanup truncate error: %v", err)
		}
		db.Close()
	})

	// Truncate before each test to ensure clean state
	_, err = db.Exec(`TRUNCATE interest_schedules, transactions, allowance_schedules, auth_events, sessions, children, parents, families RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	return db
}

// CreateTestFamily creates a family with the given slug for testing.
func CreateTestFamily(t *testing.T, db *sql.DB) *store.Family {
	t.Helper()
	fs := store.NewFamilyStore(db)
	f, err := fs.Create("test-family")
	require.NoError(t, err)
	return f
}

// CreateTestParent creates a parent associated with the given family for testing.
func CreateTestParent(t *testing.T, db *sql.DB, familyID int64) *store.Parent {
	t.Helper()
	ps := store.NewParentStore(db)
	p, err := ps.Create("google-id-123", "parent@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(p.ID, familyID)
	require.NoError(t, err)
	p.FamilyID = familyID
	return p
}

// CreateTestChild creates a child in the given family for testing.
func CreateTestChild(t *testing.T, db *sql.DB, familyID int64, name string) *store.Child {
	t.Helper()
	cs := store.NewChildStore(db)
	c, err := cs.Create(familyID, name, "password123", nil)
	require.NoError(t, err)
	return c
}

// SetRequestContext returns a new request with user type, user ID, and family ID in context.
func SetRequestContext(r *http.Request, userType string, userID, familyID int64) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, middleware.ContextKeyUserType, userType)
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, middleware.ContextKeyFamilyID, familyID)
	return r.WithContext(ctx)
}
