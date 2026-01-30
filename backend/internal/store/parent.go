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
	db *DB
}

func NewParentStore(db *DB) *ParentStore {
	return &ParentStore{db: db}
}

func (s *ParentStore) Create(googleID, email, displayName string) (*Parent, error) {
	res, err := s.db.Write.Exec(
		`INSERT INTO parents (google_id, email, display_name) VALUES (?, ?, ?)`,
		googleID, email, displayName,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("google account already registered")
		}
		return nil, fmt.Errorf("insert parent: %w", err)
	}
	id, _ := res.LastInsertId()
	return s.GetByID(id)
}

func (s *ParentStore) GetByGoogleID(googleID string) (*Parent, error) {
	var p Parent
	var createdAt string
	err := s.db.Read.QueryRow(
		`SELECT id, google_id, email, display_name, family_id, created_at
		 FROM parents WHERE google_id = ?`, googleID,
	).Scan(&p.ID, &p.GoogleID, &p.Email, &p.DisplayName, &p.FamilyID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get parent by google id: %w", err)
	}
	p.CreatedAt, _ = parseTime(createdAt)
	return &p, nil
}

func (s *ParentStore) GetByID(id int64) (*Parent, error) {
	var p Parent
	var createdAt string
	err := s.db.Read.QueryRow(
		`SELECT id, google_id, email, display_name, family_id, created_at
		 FROM parents WHERE id = ?`, id,
	).Scan(&p.ID, &p.GoogleID, &p.Email, &p.DisplayName, &p.FamilyID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get parent by id: %w", err)
	}
	p.CreatedAt, _ = parseTime(createdAt)
	return &p, nil
}

func (s *ParentStore) SetFamilyID(parentID, familyID int64) error {
	_, err := s.db.Write.Exec(
		`UPDATE parents SET family_id = ? WHERE id = ?`,
		familyID, parentID,
	)
	if err != nil {
		return fmt.Errorf("set family id: %w", err)
	}
	return nil
}
