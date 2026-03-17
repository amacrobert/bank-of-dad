package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookEventRepo_HasBeenProcessed_NewEvent(t *testing.T) {
	db := testDB(t)
	repo := NewWebhookEventRepo(db)

	processed, err := repo.HasBeenProcessed("evt_new_123")
	require.NoError(t, err)
	assert.False(t, processed)
}

func TestWebhookEventRepo_RecordEvent_ThenHasBeenProcessed(t *testing.T) {
	db := testDB(t)
	repo := NewWebhookEventRepo(db)

	err := repo.RecordEvent("evt_mark_123", "checkout.session.completed")
	require.NoError(t, err)

	processed, err := repo.HasBeenProcessed("evt_mark_123")
	require.NoError(t, err)
	assert.True(t, processed)
}

func TestWebhookEventRepo_DuplicateRecordEvent(t *testing.T) {
	db := testDB(t)
	repo := NewWebhookEventRepo(db)

	err := repo.RecordEvent("evt_dup_123", "checkout.session.completed")
	require.NoError(t, err)

	// Second call should not error (idempotent)
	err = repo.RecordEvent("evt_dup_123", "checkout.session.completed")
	require.NoError(t, err)

	processed, err := repo.HasBeenProcessed("evt_dup_123")
	require.NoError(t, err)
	assert.True(t, processed)
}
