package store

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const defaultTestDSN = "postgres://bankofdad:bankofdad@localhost:5432/bankofdad_test?sslmode=disable"

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}
	db, err := Open(dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Exec(`TRUNCATE stripe_webhook_events, interest_schedules, transactions, allowance_schedules, auth_events, refresh_tokens, children, parents, families RESTART IDENTITY CASCADE`)
		db.Close()
	})
	_, err = db.Exec(`TRUNCATE stripe_webhook_events, interest_schedules, transactions, allowance_schedules, auth_events, refresh_tokens, children, parents, families RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
	return db
}
