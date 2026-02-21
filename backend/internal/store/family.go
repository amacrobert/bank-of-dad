package store

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

type Family struct {
	ID        int64
	Slug      string
	Timezone  string
	CreatedAt time.Time
}

type FamilyStore struct {
	db *sql.DB
}

func NewFamilyStore(db *sql.DB) *FamilyStore {
	return &FamilyStore{db: db}
}

func (s *FamilyStore) Create(slug string) (*Family, error) {
	var id int64
	err := s.db.QueryRow(`INSERT INTO families (slug) VALUES ($1) RETURNING id`, slug).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("slug already taken: %s", slug)
		}
		return nil, fmt.Errorf("insert family: %w", err)
	}
	return s.GetByID(id)
}

func (s *FamilyStore) GetByID(id int64) (*Family, error) {
	var f Family
	err := s.db.QueryRow(
		`SELECT id, slug, timezone, created_at FROM families WHERE id = $1`, id,
	).Scan(&f.ID, &f.Slug, &f.Timezone, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by id: %w", err)
	}
	return &f, nil
}

func (s *FamilyStore) GetBySlug(slug string) (*Family, error) {
	var f Family
	err := s.db.QueryRow(
		`SELECT id, slug, timezone, created_at FROM families WHERE slug = $1`, slug,
	).Scan(&f.ID, &f.Slug, &f.Timezone, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by slug: %w", err)
	}
	return &f, nil
}

func (s *FamilyStore) GetTimezone(familyID int64) (string, error) {
	var tz string
	err := s.db.QueryRow(
		`SELECT timezone FROM families WHERE id = $1`, familyID,
	).Scan(&tz)
	if err != nil {
		return "", fmt.Errorf("get timezone: %w", err)
	}
	return tz, nil
}

func (s *FamilyStore) UpdateTimezone(familyID int64, timezone string) error {
	result, err := s.db.Exec(
		`UPDATE families SET timezone = $1 WHERE id = $2`, timezone, familyID,
	)
	if err != nil {
		return fmt.Errorf("update timezone: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update timezone rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("family not found: %d", familyID)
	}
	return nil
}

func (s *FamilyStore) SlugExists(slug string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM families WHERE slug = $1`, slug,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check slug exists: %w", err)
	}
	return count > 0, nil
}

func (s *FamilyStore) SuggestSlugs(base string) []string {
	suggestions := []string{
		base + "-1",
		base + "-2",
		"the-" + base,
	}

	var available []string
	for _, sug := range suggestions {
		if ValidateSlug(sug) == nil {
			exists, err := s.SlugExists(sug)
			if err == nil && !exists {
				available = append(available, sug)
			}
		}
	}
	return available
}

func (s *FamilyStore) DeleteAll(familyID, parentID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	// Delete refresh tokens for all children in the family
	if _, err := tx.Exec(`DELETE FROM refresh_tokens WHERE user_type = 'child' AND user_id IN (SELECT id FROM children WHERE family_id = $1)`, familyID); err != nil {
		return fmt.Errorf("delete child refresh tokens: %w", err)
	}

	// Delete auth events for all children in the family
	if _, err := tx.Exec(`DELETE FROM auth_events WHERE user_type = 'child' AND user_id IN (SELECT id FROM children WHERE family_id = $1)`, familyID); err != nil {
		return fmt.Errorf("delete child auth events: %w", err)
	}

	// Delete children (cascades transactions, allowance_schedules, interest_schedules)
	if _, err := tx.Exec(`DELETE FROM children WHERE family_id = $1`, familyID); err != nil {
		return fmt.Errorf("delete children: %w", err)
	}

	// Delete refresh tokens for the parent
	if _, err := tx.Exec(`DELETE FROM refresh_tokens WHERE user_type = 'parent' AND user_id = $1`, parentID); err != nil {
		return fmt.Errorf("delete parent refresh tokens: %w", err)
	}

	// Delete auth events for the parent and family-scoped events
	if _, err := tx.Exec(`DELETE FROM auth_events WHERE (user_type = 'parent' AND user_id = $1) OR family_id = $2`, parentID, familyID); err != nil {
		return fmt.Errorf("delete parent auth events: %w", err)
	}

	// Delete the parent
	if _, err := tx.Exec(`DELETE FROM parents WHERE id = $1`, parentID); err != nil {
		return fmt.Errorf("delete parent: %w", err)
	}

	// Delete the family
	if _, err := tx.Exec(`DELETE FROM families WHERE id = $1`, familyID); err != nil {
		return fmt.Errorf("delete family: %w", err)
	}

	return tx.Commit()
}

func ValidateSlug(slug string) error {
	if len(slug) < 3 || len(slug) > 30 {
		return fmt.Errorf("slug must be between 3 and 30 characters")
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen")
	}
	return nil
}
