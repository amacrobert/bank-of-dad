package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Parent struct {
	ID          int64
	GoogleID    string
	Email       string
	DisplayName string
	FamilyID    int64
	CreatedAt   time.Time
}

type ParentStore struct {
	db *sql.DB
}

func NewParentStore(db *sql.DB) *ParentStore {
	return &ParentStore{db: db}
}

func (s *ParentStore) Create(googleID, email, displayName string) (*Parent, error) {
	var id int64
	err := s.db.QueryRow(
		`INSERT INTO parents (google_id, email, display_name) VALUES ($1, $2, $3) RETURNING id`,
		googleID, email, displayName,
	).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("google account already registered")
		}
		return nil, fmt.Errorf("insert parent: %w", err)
	}
	return s.GetByID(id)
}

func (s *ParentStore) GetByGoogleID(googleID string) (*Parent, error) {
	var p Parent
	err := s.db.QueryRow(
		`SELECT id, google_id, email, display_name, family_id, created_at
		 FROM parents WHERE google_id = $1`, googleID,
	).Scan(&p.ID, &p.GoogleID, &p.Email, &p.DisplayName, &p.FamilyID, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get parent by google id: %w", err)
	}
	return &p, nil
}

func (s *ParentStore) GetByID(id int64) (*Parent, error) {
	var p Parent
	err := s.db.QueryRow(
		`SELECT id, google_id, email, display_name, family_id, created_at
		 FROM parents WHERE id = $1`, id,
	).Scan(&p.ID, &p.GoogleID, &p.Email, &p.DisplayName, &p.FamilyID, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get parent by id: %w", err)
	}
	return &p, nil
}

func (s *ParentStore) SetFamilyID(parentID, familyID int64) error {
	_, err := s.db.Exec(
		`UPDATE parents SET family_id = $1 WHERE id = $2`,
		familyID, parentID,
	)
	if err != nil {
		return fmt.Errorf("set family id: %w", err)
	}
	return nil
}
