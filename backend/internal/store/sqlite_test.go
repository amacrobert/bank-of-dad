package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper that opens a DB backed by a temporary file and registers cleanup.
func openTestDB(t *testing.T) *DB {
	t.Helper()

	tmp, err := os.CreateTemp("", "bank-of-dad-test-*.db")
	require.NoError(t, err)
	path := tmp.Name()
	tmp.Close()

	t.Cleanup(func() {
		os.Remove(path)
		os.Remove(path + "-wal")
		os.Remove(path + "-shm")
	})

	db, err := Open(path)
	require.NoError(t, err, "Open should succeed for a valid temp path")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestOpen_Success(t *testing.T) {
	db := openTestDB(t)

	assert.NotNil(t, db.Write, "Write connection should not be nil")
	assert.NotNil(t, db.Read, "Read connection should not be nil")

	// Both connections should be alive.
	assert.NoError(t, db.Write.Ping(), "Write connection should be pingable")
	assert.NoError(t, db.Read.Ping(), "Read connection should be pingable")
}

func TestOpen_CreatesAllTables(t *testing.T) {
	db := openTestDB(t)

	expectedTables := []string{
		"families",
		"parents",
		"children",
		"sessions",
		"auth_events",
	}

	for _, table := range expectedTables {
		var name string
		err := db.Read.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&name)
		require.NoError(t, err, "table %q should exist", table)
		assert.Equal(t, table, name)
	}
}

func TestOpen_WALModeActive(t *testing.T) {
	db := openTestDB(t)

	// Check WAL mode on the Write connection.
	var writeMode string
	err := db.Write.QueryRow("PRAGMA journal_mode").Scan(&writeMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", writeMode, "Write connection should use WAL journal mode")

	// Check WAL mode on the Read connection.
	var readMode string
	err = db.Read.QueryRow("PRAGMA journal_mode").Scan(&readMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", readMode, "Read connection should use WAL journal mode")
}

func TestOpen_ForeignKeysEnabled(t *testing.T) {
	db := openTestDB(t)

	// Check foreign_keys on the Write connection.
	var writeFk int
	err := db.Write.QueryRow("PRAGMA foreign_keys").Scan(&writeFk)
	require.NoError(t, err)
	assert.Equal(t, 1, writeFk, "Write connection should have foreign_keys enabled")

	// Check foreign_keys on the Read connection.
	var readFk int
	err = db.Read.QueryRow("PRAGMA foreign_keys").Scan(&readFk)
	require.NoError(t, err)
	assert.Equal(t, 1, readFk, "Read connection should have foreign_keys enabled")
}

func TestDB_Close(t *testing.T) {
	tmp, err := os.CreateTemp("", "bank-of-dad-close-test-*.db")
	require.NoError(t, err)
	path := tmp.Name()
	tmp.Close()

	t.Cleanup(func() {
		os.Remove(path)
		os.Remove(path + "-wal")
		os.Remove(path + "-shm")
	})

	db, err := Open(path)
	require.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err, "Close should not return an error")

	// After closing, Ping should fail on both connections.
	assert.Error(t, db.Write.Ping(), "Write Ping should fail after Close")
	assert.Error(t, db.Read.Ping(), "Read Ping should fail after Close")
}
