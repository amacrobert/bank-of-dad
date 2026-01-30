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
	CreatedAt time.Time
}

type FamilyStore struct {
	db *DB
}

func NewFamilyStore(db *DB) *FamilyStore {
	return &FamilyStore{db: db}
}

func (s *FamilyStore) Create(slug string) (*Family, error) {
	res, err := s.db.Write.Exec(`INSERT INTO families (slug) VALUES (?)`, slug)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("slug already taken: %s", slug)
		}
		return nil, fmt.Errorf("insert family: %w", err)
	}
	id, _ := res.LastInsertId()
	return s.GetByID(id)
}

func (s *FamilyStore) GetByID(id int64) (*Family, error) {
	var f Family
	var createdAt string
	err := s.db.Read.QueryRow(
		`SELECT id, slug, created_at FROM families WHERE id = ?`, id,
	).Scan(&f.ID, &f.Slug, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by id: %w", err)
	}
	f.CreatedAt, _ = parseTime(createdAt)
	return &f, nil
}

func (s *FamilyStore) GetBySlug(slug string) (*Family, error) {
	var f Family
	var createdAt string
	err := s.db.Read.QueryRow(
		`SELECT id, slug, created_at FROM families WHERE slug = ?`, slug,
	).Scan(&f.ID, &f.Slug, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by slug: %w", err)
	}
	f.CreatedAt, _ = parseTime(createdAt)
	return &f, nil
}

func (s *FamilyStore) SlugExists(slug string) (bool, error) {
	var count int
	err := s.db.Read.QueryRow(
		`SELECT COUNT(*) FROM families WHERE slug = ?`, slug,
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

func ValidateSlug(slug string) error {
	if len(slug) < 3 || len(slug) > 30 {
		return fmt.Errorf("slug must be between 3 and 30 characters")
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen")
	}
	return nil
}
