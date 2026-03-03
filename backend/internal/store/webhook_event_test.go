package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookEventStore_IsProcessed_NewEvent(t *testing.T) {
	db := testDB(t)
	ws := NewWebhookEventStore(db)

	processed, err := ws.IsProcessed("evt_new_123")
	require.NoError(t, err)
	assert.False(t, processed)
}

func TestWebhookEventStore_MarkProcessed_ThenIsProcessed(t *testing.T) {
	db := testDB(t)
	ws := NewWebhookEventStore(db)

	err := ws.MarkProcessed("evt_mark_123", "checkout.session.completed")
	require.NoError(t, err)

	processed, err := ws.IsProcessed("evt_mark_123")
	require.NoError(t, err)
	assert.True(t, processed)
}

func TestWebhookEventStore_DuplicateMarkProcessed(t *testing.T) {
	db := testDB(t)
	ws := NewWebhookEventStore(db)

	err := ws.MarkProcessed("evt_dup_123", "checkout.session.completed")
	require.NoError(t, err)

	// Second call should not error (ON CONFLICT DO NOTHING)
	err = ws.MarkProcessed("evt_dup_123", "checkout.session.completed")
	require.NoError(t, err)

	processed, err := ws.IsProcessed("evt_dup_123")
	require.NoError(t, err)
	assert.True(t, processed)
}
