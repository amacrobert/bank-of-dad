package store

import (
	"database/sql"
	"fmt"
	"runtime"

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
		db.Close()
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
	}

	for _, stmt := range statements {
		if _, err := db.Write.Exec(stmt); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}

	return nil
}
