package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Child struct {
	ID                  int64
	FamilyID            int64
	FirstName           string
	PasswordHash        string
	IsLocked            bool
	FailedLoginAttempts int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ChildStore struct {
	db *DB
}

func NewChildStore(db *DB) *ChildStore {
	return &ChildStore{db: db}
}

func (s *ChildStore) Create(familyID int64, firstName, password string) (*Child, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	res, err := s.db.Write.Exec(
		`INSERT INTO children (family_id, first_name, password_hash) VALUES (?, ?, ?)`,
		familyID, firstName, string(hash),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("child named %q already exists in this family", firstName)
		}
		return nil, fmt.Errorf("insert child: %w", err)
	}

	id, _ := res.LastInsertId()
	return s.GetByID(id)
}

func (s *ChildStore) GetByID(id int64) (*Child, error) {
	var c Child
	var isLocked int
	var createdAt, updatedAt string
	err := s.db.Read.QueryRow(
		`SELECT id, family_id, first_name, password_hash, is_locked, failed_login_attempts, created_at, updated_at
		 FROM children WHERE id = ?`, id,
	).Scan(&c.ID, &c.FamilyID, &c.FirstName, &c.PasswordHash, &isLocked, &c.FailedLoginAttempts, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get child by id: %w", err)
	}
	c.IsLocked = isLocked != 0
	c.CreatedAt, _ = parseTime(createdAt)
	c.UpdatedAt, _ = parseTime(updatedAt)
	return &c, nil
}

func (s *ChildStore) GetByFamilyAndName(familyID int64, firstName string) (*Child, error) {
	var c Child
	var isLocked int
	var createdAt, updatedAt string
	err := s.db.Read.QueryRow(
		`SELECT id, family_id, first_name, password_hash, is_locked, failed_login_attempts, created_at, updated_at
		 FROM children WHERE family_id = ? AND first_name = ?`, familyID, firstName,
	).Scan(&c.ID, &c.FamilyID, &c.FirstName, &c.PasswordHash, &isLocked, &c.FailedLoginAttempts, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get child by family and name: %w", err)
	}
	c.IsLocked = isLocked != 0
	c.CreatedAt, _ = parseTime(createdAt)
	c.UpdatedAt, _ = parseTime(updatedAt)
	return &c, nil
}

func (s *ChildStore) ListByFamily(familyID int64) ([]Child, error) {
	rows, err := s.db.Read.Query(
		`SELECT id, family_id, first_name, password_hash, is_locked, failed_login_attempts, created_at, updated_at
		 FROM children WHERE family_id = ? ORDER BY first_name`, familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list children: %w", err)
	}
	defer rows.Close()

	var children []Child
	for rows.Next() {
		var c Child
		var isLocked int
		var createdAt, updatedAt string
		if err := rows.Scan(&c.ID, &c.FamilyID, &c.FirstName, &c.PasswordHash, &isLocked, &c.FailedLoginAttempts, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan child: %w", err)
		}
		c.IsLocked = isLocked != 0
		c.CreatedAt, _ = parseTime(createdAt)
		c.UpdatedAt, _ = parseTime(updatedAt)
		children = append(children, c)
	}
	return children, rows.Err()
}

func (s *ChildStore) CheckPassword(child *Child, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(child.PasswordHash), []byte(password)) == nil
}

func (s *ChildStore) IncrementFailedAttempts(id int64) (int, error) {
	_, err := s.db.Write.Exec(
		`UPDATE children SET failed_login_attempts = failed_login_attempts + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id,
	)
	if err != nil {
		return 0, fmt.Errorf("increment failed attempts: %w", err)
	}
	var attempts int
	err = s.db.Read.QueryRow(`SELECT failed_login_attempts FROM children WHERE id = ?`, id).Scan(&attempts)
	if err != nil {
		return 0, fmt.Errorf("read failed attempts: %w", err)
	}
	return attempts, nil
}

func (s *ChildStore) LockAccount(id int64) error {
	_, err := s.db.Write.Exec(
		`UPDATE children SET is_locked = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id,
	)
	if err != nil {
		return fmt.Errorf("lock account: %w", err)
	}
	return nil
}

func (s *ChildStore) ResetFailedAttempts(id int64) error {
	_, err := s.db.Write.Exec(
		`UPDATE children SET failed_login_attempts = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id,
	)
	if err != nil {
		return fmt.Errorf("reset failed attempts: %w", err)
	}
	return nil
}

func (s *ChildStore) UpdatePassword(id int64, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = s.db.Write.Exec(
		`UPDATE children SET password_hash = ?, is_locked = 0, failed_login_attempts = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		string(hash), id,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (s *ChildStore) UpdateName(id, familyID int64, newName string) error {
	_, err := s.db.Write.Exec(
		`UPDATE children SET first_name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND family_id = ?`,
		newName, id, familyID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("child named %q already exists in this family", newName)
		}
		return fmt.Errorf("update name: %w", err)
	}
	return nil
}
