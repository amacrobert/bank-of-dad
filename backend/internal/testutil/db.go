package testutil

import (
	"context"
	"net/http"
	"os"
	"testing"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/models"
	"bank-of-dad/repositories"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const defaultTestDSN = "postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable"

// SetupTestDB opens a connection to the test Postgres database, runs migrations,
// and registers cleanup that truncates all tables after the test completes.
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	db, err := repositories.Open(dsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		// Truncate all tables in dependency order
		result := db.Exec(`TRUNCATE withdrawal_requests, chore_instances, chore_assignments, chores, goal_allocations, savings_goals, stripe_webhook_events, interest_schedules, transactions, allowance_schedules, auth_events, refresh_tokens, children, parents, families RESTART IDENTITY CASCADE`)
		if result.Error != nil {
			t.Logf("cleanup truncate error: %v", result.Error)
		}
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	// Truncate before each test to ensure clean state
	result := db.Exec(`TRUNCATE withdrawal_requests, chore_instances, chore_assignments, chores, goal_allocations, savings_goals, stripe_webhook_events, interest_schedules, transactions, allowance_schedules, auth_events, refresh_tokens, children, parents, families RESTART IDENTITY CASCADE`)
	require.NoError(t, result.Error)

	return db
}

// CreateTestFamily creates a family with the given slug for testing.
func CreateTestFamily(t *testing.T, db *gorm.DB) *models.Family {
	t.Helper()
	fs := repositories.NewFamilyRepo(db)
	f, err := fs.Create("test-family")
	require.NoError(t, err)
	return f
}

// CreateTestParent creates a parent associated with the given family for testing.
func CreateTestParent(t *testing.T, db *gorm.DB, familyID int64) *models.Parent {
	t.Helper()
	ps := repositories.NewParentRepo(db)
	p, err := ps.Create("google-id-123", "parent@test.com", "Test Parent")
	require.NoError(t, err)
	err = ps.SetFamilyID(p.ID, familyID)
	require.NoError(t, err)
	p.FamilyID = familyID
	return p
}

// CreateTestChild creates a child in the given family for testing.
func CreateTestChild(t *testing.T, db *gorm.DB, familyID int64, name string) *models.Child {
	t.Helper()
	cs := repositories.NewChildRepo(db)
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
