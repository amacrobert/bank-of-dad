package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFamily(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	f, err := fs.Create("smith-family")
	require.NoError(t, err)
	assert.Equal(t, "smith-family", f.Slug)
	assert.NotZero(t, f.ID)
	assert.False(t, f.CreatedAt.IsZero())
}

func TestCreateFamily_DuplicateSlug(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	_, err := fs.Create("smith-family")
	require.NoError(t, err)

	_, err = fs.Create("smith-family")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slug already taken")
}

func TestGetBySlug(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	created, err := fs.Create("jones-bank")
	require.NoError(t, err)

	found, err := fs.GetBySlug("jones-bank")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "jones-bank", found.Slug)
}

func TestGetBySlug_NotFound(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	found, err := fs.GetBySlug("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestSlugExists(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	exists, err := fs.SlugExists("test-slug")
	require.NoError(t, err)
	assert.False(t, exists)

	_, err = fs.Create("test-slug")
	require.NoError(t, err)

	exists, err = fs.SlugExists("test-slug")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		slug  string
		valid bool
	}{
		{"abc", true},
		{"smith-family", true},
		{"my-bank-123", true},
		{"a1b", true},
		{"ab", false},                         // too short
		{"a", false},                          // too short
		{"", false},                           // empty
		{"-abc", false},                       // starts with hyphen
		{"abc-", false},                       // ends with hyphen
		{"ABC", false},                        // uppercase
		{"abc def", false},                    // space
		{"abc_def", false},                    // underscore
		{"abcdefghij1234567890abcdefghij1", false}, // > 30 chars
	}

	for _, tt := range tests {
		err := ValidateSlug(tt.slug)
		if tt.valid {
			assert.NoError(t, err, "slug %q should be valid", tt.slug)
		} else {
			assert.Error(t, err, "slug %q should be invalid", tt.slug)
		}
	}
}

func TestSuggestSlugs(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	suggestions := fs.SuggestSlugs("smith-family")
	assert.NotEmpty(t, suggestions)
	for _, s := range suggestions {
		assert.NoError(t, ValidateSlug(s))
	}
}
