package repositories

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const defaultTestDSN = "postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable"

var (
	sharedDB   *gorm.DB
	sharedOnce sync.Once
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()

	sharedOnce.Do(func() {
		dsn := os.Getenv("TEST_DATABASE_URL")
		if dsn == "" {
			dsn = defaultTestDSN
		}
		db, err := Open(dsn)
		if err != nil {
			panic("failed to open test db: " + err.Error())
		}
		sqlDB, err := db.DB()
		if err != nil {
			panic("failed to get underlying sql.DB: " + err.Error())
		}
		sqlDB.SetMaxOpenConns(5)
		sharedDB = db
	})

	result := sharedDB.Exec(`TRUNCATE withdrawal_requests, chore_instances, chore_assignments, chores, goal_allocations, savings_goals, stripe_webhook_events, interest_schedules, transactions, allowance_schedules, auth_events, refresh_tokens, children, parents, families RESTART IDENTITY CASCADE`)
	require.NoError(t, result.Error)

	return sharedDB
}
