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
	BalanceCents        int64
	Avatar              *string
	Theme               *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ChildStore struct {
	db *sql.DB
}

func NewChildStore(db *sql.DB) *ChildStore {
	return &ChildStore{db: db}
}

func (s *ChildStore) Create(familyID int64, firstName, password string, avatar *string) (*Child, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var id int64
	err = s.db.QueryRow(
		`INSERT INTO children (family_id, first_name, password_hash, avatar) VALUES ($1, $2, $3, $4) RETURNING id`,
		familyID, firstName, string(hash), avatar,
	).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("child named %q already exists in this family", firstName)
		}
		return nil, fmt.Errorf("insert child: %w", err)
	}

	return s.GetByID(id)
}

func (s *ChildStore) GetByID(id int64) (*Child, error) {
	var c Child
	var avatar, theme sql.NullString
	err := s.db.QueryRow(
		`SELECT id, family_id, first_name, password_hash, is_locked, failed_login_attempts, balance_cents, avatar, theme, created_at, updated_at
		 FROM children WHERE id = $1`, id,
	).Scan(&c.ID, &c.FamilyID, &c.FirstName, &c.PasswordHash, &c.IsLocked, &c.FailedLoginAttempts, &c.BalanceCents, &avatar, &theme, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get child by id: %w", err)
	}
	if avatar.Valid {
		c.Avatar = &avatar.String
	}
	if theme.Valid {
		c.Theme = &theme.String
	}
	return &c, nil
}

func (s *ChildStore) GetByFamilyAndName(familyID int64, firstName string) (*Child, error) {
	var c Child
	var avatar, theme sql.NullString
	err := s.db.QueryRow(
		`SELECT id, family_id, first_name, password_hash, is_locked, failed_login_attempts, balance_cents, avatar, theme, created_at, updated_at
		 FROM children WHERE family_id = $1 AND first_name = $2`, familyID, firstName,
	).Scan(&c.ID, &c.FamilyID, &c.FirstName, &c.PasswordHash, &c.IsLocked, &c.FailedLoginAttempts, &c.BalanceCents, &avatar, &theme, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get child by family and name: %w", err)
	}
	if avatar.Valid {
		c.Avatar = &avatar.String
	}
	if theme.Valid {
		c.Theme = &theme.String
	}
	return &c, nil
}

func (s *ChildStore) ListByFamily(familyID int64) ([]Child, error) {
	rows, err := s.db.Query(
		`SELECT id, family_id, first_name, password_hash, is_locked, failed_login_attempts, balance_cents, avatar, theme, created_at, updated_at
		 FROM children WHERE family_id = $1 ORDER BY id`, familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list children: %w", err)
	}
	defer rows.Close()

	var children []Child
	for rows.Next() {
		var c Child
		var avatar, theme sql.NullString
		if err := rows.Scan(&c.ID, &c.FamilyID, &c.FirstName, &c.PasswordHash, &c.IsLocked, &c.FailedLoginAttempts, &c.BalanceCents, &avatar, &theme, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan child: %w", err)
		}
		if avatar.Valid {
			c.Avatar = &avatar.String
		}
		if theme.Valid {
			c.Theme = &theme.String
		}
		children = append(children, c)
	}
	return children, rows.Err()
}

func (s *ChildStore) CheckPassword(child *Child, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(child.PasswordHash), []byte(password)) == nil
}

func (s *ChildStore) IncrementFailedAttempts(id int64) (int, error) {
	_, err := s.db.Exec(
		`UPDATE children SET failed_login_attempts = failed_login_attempts + 1, updated_at = NOW() WHERE id = $1`, id,
	)
	if err != nil {
		return 0, fmt.Errorf("increment failed attempts: %w", err)
	}
	var attempts int
	err = s.db.QueryRow(`SELECT failed_login_attempts FROM children WHERE id = $1`, id).Scan(&attempts)
	if err != nil {
		return 0, fmt.Errorf("read failed attempts: %w", err)
	}
	return attempts, nil
}

func (s *ChildStore) LockAccount(id int64) error {
	_, err := s.db.Exec(
		`UPDATE children SET is_locked = TRUE, updated_at = NOW() WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("lock account: %w", err)
	}
	return nil
}

func (s *ChildStore) ResetFailedAttempts(id int64) error {
	_, err := s.db.Exec(
		`UPDATE children SET failed_login_attempts = 0, updated_at = NOW() WHERE id = $1`, id,
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
	_, err = s.db.Exec(
		`UPDATE children SET password_hash = $1, is_locked = FALSE, failed_login_attempts = 0, updated_at = NOW() WHERE id = $2`,
		string(hash), id,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

// UpdateNameAndAvatar updates the child's name and optionally their avatar.
// avatarSet indicates whether the avatar field was provided in the request.
// When avatarSet is true, avatar is applied (nil clears, non-nil sets).
// When avatarSet is false, the avatar is left unchanged.
func (s *ChildStore) UpdateNameAndAvatar(id, familyID int64, newName string, avatar *string, avatarSet bool) error {
	var err error
	if avatarSet {
		_, err = s.db.Exec(
			`UPDATE children SET first_name = $1, avatar = $2, updated_at = NOW() WHERE id = $3 AND family_id = $4`,
			newName, avatar, id, familyID,
		)
	} else {
		_, err = s.db.Exec(
			`UPDATE children SET first_name = $1, updated_at = NOW() WHERE id = $2 AND family_id = $3`,
			newName, id, familyID,
		)
	}
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("child named %q already exists in this family", newName)
		}
		return fmt.Errorf("update name: %w", err)
	}
	return nil
}

// UpdateTheme sets the child's visual theme preference.
func (s *ChildStore) UpdateTheme(childID int64, theme string) error {
	result, err := s.db.Exec(
		`UPDATE children SET theme = $1, updated_at = NOW() WHERE id = $2`,
		theme, childID,
	)
	if err != nil {
		return fmt.Errorf("update theme: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update theme rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("child not found")
	}
	return nil
}

// UpdateAvatar sets the child's avatar emoji (or clears it if avatar is nil).
func (s *ChildStore) UpdateAvatar(childID int64, avatar *string) error {
	result, err := s.db.Exec(
		`UPDATE children SET avatar = $1, updated_at = NOW() WHERE id = $2`,
		avatar, childID,
	)
	if err != nil {
		return fmt.Errorf("update avatar: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update avatar rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("child not found")
	}
	return nil
}

// Delete permanently removes a child and all associated data in a single
// atomic transaction. Cascading foreign keys handle transactions, allowance
// schedules, and interest schedules. Sessions and auth events are deleted
// explicitly since they reference user_id without a foreign key.
func (s *ChildStore) Delete(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	if _, err := tx.Exec(`DELETE FROM refresh_tokens WHERE user_type = 'child' AND user_id = $1`, id); err != nil {
		return fmt.Errorf("delete child refresh tokens: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM auth_events WHERE user_type = 'child' AND user_id = $1`, id); err != nil {
		return fmt.Errorf("delete child auth events: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM children WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete child: %w", err)
	}

	return tx.Commit()
}

// GetBalance returns the current balance in cents for a child.
func (s *ChildStore) GetBalance(id int64) (int64, error) {
	var balance int64
	err := s.db.QueryRow(`SELECT balance_cents FROM children WHERE id = $1`, id).Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("child not found")
	}
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}
	return balance, nil
}
