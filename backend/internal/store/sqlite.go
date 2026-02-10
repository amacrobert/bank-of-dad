package store

import (
	"database/sql"
	"fmt"
	"runtime"
	"strings"

	_ "modernc.org/sqlite"
)

type DB struct {
	Write *sql.DB
	Read  *sql.DB
}

func Open(path string) (*DB, error) {
	writeDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open write db: %w", err)
	}
	writeDB.SetMaxOpenConns(1)

	if err := setPragmas(writeDB); err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("set write pragmas: %w", err)
	}

	readDB, err := sql.Open("sqlite", path)
	if err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("open read db: %w", err)
	}
	readDB.SetMaxOpenConns(runtime.NumCPU())

	if err := setPragmas(readDB); err != nil {
		writeDB.Close()
		readDB.Close()
		return nil, fmt.Errorf("set read pragmas: %w", err)
	}

	db := &DB{Write: writeDB, Read: readDB}

	if err := db.migrate(); err != nil {
		db.Close() //nolint:errcheck // closing on initialization failure
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	err1 := db.Write.Close()
	err2 := db.Read.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func setPragmas(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA foreign_keys = ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("exec %q: %w", p, err)
		}
	}
	return nil
}

func (db *DB) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS families (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			slug        TEXT    NOT NULL UNIQUE,
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS parents (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			google_id    TEXT    NOT NULL UNIQUE,
			email        TEXT    NOT NULL,
			display_name TEXT    NOT NULL,
			family_id    INTEGER NOT NULL DEFAULT 0,
			created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS children (
			id                    INTEGER PRIMARY KEY AUTOINCREMENT,
			family_id             INTEGER NOT NULL REFERENCES families(id),
			first_name            TEXT    NOT NULL,
			password_hash         TEXT    NOT NULL,
			is_locked             INTEGER NOT NULL DEFAULT 0,
			failed_login_attempts INTEGER NOT NULL DEFAULT 0,
			created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(family_id, first_name)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token      TEXT     PRIMARY KEY,
			user_type  TEXT     NOT NULL CHECK(user_type IN ('parent', 'child')),
			user_id    INTEGER  NOT NULL,
			family_id  INTEGER  NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)`,
		`CREATE TABLE IF NOT EXISTS auth_events (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			event_type  TEXT     NOT NULL,
			user_type   TEXT     NOT NULL,
			user_id     INTEGER,
			family_id   INTEGER,
			ip_address  TEXT     NOT NULL,
			details     TEXT,
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_auth_events_created ON auth_events(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_auth_events_family ON auth_events(family_id)`,
		// Account balances feature (002-account-balances)
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			child_id INTEGER NOT NULL,
			parent_id INTEGER NOT NULL,
			amount_cents INTEGER NOT NULL,
			transaction_type TEXT NOT NULL CHECK(transaction_type IN ('deposit', 'withdrawal', 'allowance')),
			note TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
			FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE RESTRICT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_child_created ON transactions(child_id, created_at DESC)`,
	}

	for _, stmt := range statements {
		if _, err := db.Write.Exec(stmt); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}

	// Add balance_cents column if it doesn't exist (idempotent migration for existing databases)
	if err := db.addColumnIfNotExists("children", "balance_cents", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("add balance_cents column: %w", err)
	}

	// Migrate transactions CHECK constraint to allow 'allowance' type (003-allowance-scheduling)
	// For existing databases, the old CHECK constraint only allows ('deposit', 'withdrawal').
	// SQLite doesn't support ALTER CHECK, so we recreate the table if needed.
	if err := db.migrateTransactionsCheckConstraint(); err != nil {
		return fmt.Errorf("migrate transactions check constraint: %w", err)
	}

	// Allowance scheduling feature (003-allowance-scheduling)
	allowanceStatements := []string{
		`CREATE TABLE IF NOT EXISTS allowance_schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			child_id INTEGER NOT NULL,
			parent_id INTEGER NOT NULL,
			amount_cents INTEGER NOT NULL,
			frequency TEXT NOT NULL CHECK(frequency IN ('weekly', 'biweekly', 'monthly')),
			day_of_week INTEGER CHECK(day_of_week >= 0 AND day_of_week <= 6),
			day_of_month INTEGER CHECK(day_of_month >= 1 AND day_of_month <= 31),
			note TEXT,
			status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'paused')),
			next_run_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
			FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE RESTRICT,
			CHECK(
				(frequency = 'weekly' AND day_of_week IS NOT NULL) OR
				(frequency = 'biweekly' AND day_of_week IS NOT NULL) OR
				(frequency = 'monthly' AND day_of_month IS NOT NULL)
			)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_due ON allowance_schedules(status, next_run_at)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_child ON allowance_schedules(child_id)`,
	}

	for _, stmt := range allowanceStatements {
		if _, err := db.Write.Exec(stmt); err != nil {
			return fmt.Errorf("exec allowance migration: %w", err)
		}
	}

	// Add schedule_id column to transactions (nullable FK for allowance tracking)
	if err := db.addColumnIfNotExists("transactions", "schedule_id", "INTEGER REFERENCES allowance_schedules(id) ON DELETE SET NULL"); err != nil {
		return fmt.Errorf("add schedule_id column: %w", err)
	}

	// Interest accrual feature (005-interest-accrual)
	if err := db.addColumnIfNotExists("children", "interest_rate_bps", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("add interest_rate_bps column: %w", err)
	}
	if err := db.addColumnIfNotExists("children", "last_interest_at", "DATETIME"); err != nil {
		return fmt.Errorf("add last_interest_at column: %w", err)
	}

	// Migrate transactions CHECK constraint to include 'interest' type (005-interest-accrual)
	if err := db.migrateTransactionsInterestType(); err != nil {
		return fmt.Errorf("migrate transactions interest type: %w", err)
	}

	// Account management enhancements (006-account-management-enhancements)

	// Interest accrual schedule table
	interestScheduleStatements := []string{
		`CREATE TABLE IF NOT EXISTS interest_schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			child_id INTEGER NOT NULL UNIQUE,
			parent_id INTEGER NOT NULL,
			frequency TEXT NOT NULL CHECK(frequency IN ('weekly', 'biweekly', 'monthly')),
			day_of_week INTEGER CHECK(day_of_week >= 0 AND day_of_week <= 6),
			day_of_month INTEGER CHECK(day_of_month >= 1 AND day_of_month <= 31),
			status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'paused')),
			next_run_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
			FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE RESTRICT,
			CHECK(
				(frequency = 'weekly' AND day_of_week IS NOT NULL) OR
				(frequency = 'biweekly' AND day_of_week IS NOT NULL) OR
				(frequency = 'monthly' AND day_of_month IS NOT NULL)
			)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_interest_schedules_due ON interest_schedules(status, next_run_at)`,
	}

	for _, stmt := range interestScheduleStatements {
		if _, err := db.Write.Exec(stmt); err != nil {
			return fmt.Errorf("exec interest schedule migration: %w", err)
		}
	}

	// Enforce one allowance per child (deduplicate then add UNIQUE index)
	if err := db.migrateAllowanceUniqueChild(); err != nil {
		return fmt.Errorf("migrate allowance unique child: %w", err)
	}

	return nil
}

// addColumnIfNotExists adds a column to a table if it doesn't already exist.
// This is used for idempotent migrations since SQLite doesn't support IF NOT EXISTS for ALTER TABLE.
func (db *DB) addColumnIfNotExists(table, column, definition string) error {
	// Check if column exists by querying table_info
	rows, err := db.Read.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("query table_info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("scan table_info: %w", err)
		}
		if name == column {
			// Column already exists
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate table_info: %w", err)
	}

	// Column doesn't exist, add it
	_, err = db.Write.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition))
	if err != nil {
		return fmt.Errorf("alter table: %w", err)
	}

	return nil
}

// migrateTransactionsCheckConstraint recreates the transactions table if its CHECK constraint
// doesn't include 'allowance'. This is needed because SQLite doesn't support ALTER CHECK.
func (db *DB) migrateTransactionsCheckConstraint() error {
	// Check the current table SQL to see if 'allowance' is already in the constraint
	var tableSql string
	err := db.Read.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='transactions'").Scan(&tableSql)
	if err == sql.ErrNoRows {
		// Table doesn't exist yet, will be created by main migration
		return nil
	}
	if err != nil {
		return fmt.Errorf("query sqlite_master: %w", err)
	}

	// If the constraint already includes 'allowance', no migration needed
	if strings.Contains(tableSql, "allowance") {
		return nil
	}

	// Need to recreate the table. Check if schedule_id column exists already.
	hasScheduleID := strings.Contains(tableSql, "schedule_id")

	// Temporarily disable foreign keys for the migration
	if _, err := db.Write.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}

	tx, err := db.Write.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Create new table with updated constraint
	newTableSQL := `CREATE TABLE transactions_new (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		child_id INTEGER NOT NULL,
		parent_id INTEGER NOT NULL,
		amount_cents INTEGER NOT NULL,
		transaction_type TEXT NOT NULL CHECK(transaction_type IN ('deposit', 'withdrawal', 'allowance')),
		note TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		schedule_id INTEGER REFERENCES allowance_schedules(id) ON DELETE SET NULL,
		FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
		FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE RESTRICT
	)`
	if _, err := tx.Exec(newTableSQL); err != nil {
		return fmt.Errorf("create new table: %w", err)
	}

	// Copy data from old table
	var copySQL string
	if hasScheduleID {
		copySQL = `INSERT INTO transactions_new (id, child_id, parent_id, amount_cents, transaction_type, note, created_at, schedule_id)
			SELECT id, child_id, parent_id, amount_cents, transaction_type, note, created_at, schedule_id FROM transactions`
	} else {
		copySQL = `INSERT INTO transactions_new (id, child_id, parent_id, amount_cents, transaction_type, note, created_at)
			SELECT id, child_id, parent_id, amount_cents, transaction_type, note, created_at FROM transactions`
	}
	if _, err := tx.Exec(copySQL); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}

	// Drop old table
	if _, err := tx.Exec("DROP TABLE transactions"); err != nil {
		return fmt.Errorf("drop old table: %w", err)
	}

	// Rename new table
	if _, err := tx.Exec("ALTER TABLE transactions_new RENAME TO transactions"); err != nil {
		return fmt.Errorf("rename table: %w", err)
	}

	// Recreate the index
	if _, err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_transactions_child_created ON transactions(child_id, created_at DESC)"); err != nil {
		return fmt.Errorf("recreate index: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	// Re-enable foreign keys
	if _, err := db.Write.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("re-enable foreign keys: %w", err)
	}

	return nil
}

// migrateTransactionsInterestType recreates the transactions table if its CHECK constraint
// doesn't include 'interest'. This is needed because SQLite doesn't support ALTER CHECK.
func (db *DB) migrateTransactionsInterestType() error {
	var tableSql string
	err := db.Read.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='transactions'").Scan(&tableSql)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("query sqlite_master: %w", err)
	}

	if strings.Contains(tableSql, "interest") {
		return nil
	}

	if _, err := db.Write.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}

	tx, err := db.Write.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	newTableSQL := `CREATE TABLE transactions_new (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		child_id INTEGER NOT NULL,
		parent_id INTEGER NOT NULL,
		amount_cents INTEGER NOT NULL,
		transaction_type TEXT NOT NULL CHECK(transaction_type IN ('deposit', 'withdrawal', 'allowance', 'interest')),
		note TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		schedule_id INTEGER REFERENCES allowance_schedules(id) ON DELETE SET NULL,
		FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
		FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE RESTRICT
	)`
	if _, err := tx.Exec(newTableSQL); err != nil {
		return fmt.Errorf("create new table: %w", err)
	}

	copySQL := `INSERT INTO transactions_new (id, child_id, parent_id, amount_cents, transaction_type, note, created_at, schedule_id)
		SELECT id, child_id, parent_id, amount_cents, transaction_type, note, created_at, schedule_id FROM transactions`
	if _, err := tx.Exec(copySQL); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}

	if _, err := tx.Exec("DROP TABLE transactions"); err != nil {
		return fmt.Errorf("drop old table: %w", err)
	}

	if _, err := tx.Exec("ALTER TABLE transactions_new RENAME TO transactions"); err != nil {
		return fmt.Errorf("rename table: %w", err)
	}

	if _, err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_transactions_child_created ON transactions(child_id, created_at DESC)"); err != nil {
		return fmt.Errorf("recreate index: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if _, err := db.Write.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("re-enable foreign keys: %w", err)
	}

	return nil
}

// migrateAllowanceUniqueChild adds a UNIQUE index on allowance_schedules.child_id.
// If duplicates exist, keeps the newest schedule (highest id) per child.
func (db *DB) migrateAllowanceUniqueChild() error {
	var count int
	err := db.Read.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_allowance_schedules_unique_child'`,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check unique index: %w", err)
	}
	if count > 0 {
		return nil
	}

	_, err = db.Write.Exec(`
		DELETE FROM allowance_schedules
		WHERE id NOT IN (
			SELECT MAX(id) FROM allowance_schedules GROUP BY child_id
		)
	`)
	if err != nil {
		return fmt.Errorf("deduplicate allowance schedules: %w", err)
	}

	_, err = db.Write.Exec(`CREATE UNIQUE INDEX idx_allowance_schedules_unique_child ON allowance_schedules(child_id)`)
	if err != nil {
		return fmt.Errorf("create unique index: %w", err)
	}

	return nil
}

