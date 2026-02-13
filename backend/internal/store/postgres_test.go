package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	db := testDB(t)
	require.NotNil(t, db)

	// Verify the connection is usable
	var result int
	err := db.QueryRow("SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestOpen_Idempotent(t *testing.T) {
	// First call creates migrations
	db1 := testDB(t)
	require.NotNil(t, db1)

	// Second Open on the same database should succeed (migrations already applied)
	dsn := getDSN()
	db2, err := Open(dsn)
	require.NoError(t, err)
	require.NotNil(t, db2)
	defer db2.Close()

	// Both connections should be usable
	var result int
	err = db2.QueryRow("SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestOpen_MigrationsApplied(t *testing.T) {
	db := testDB(t)

	// Verify core tables exist by querying them
	tables := []string{
		"families", "parents", "children", "sessions",
		"auth_events", "transactions", "allowance_schedules",
		"interest_schedules",
	}

	for _, table := range tables {
		var exists bool
		err := db.QueryRow(
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table,
		).Scan(&exists)
		require.NoError(t, err, "checking table %s", table)
		assert.True(t, exists, "table %s should exist", table)
	}
}

func getDSN() string {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}
	return dsn
}
