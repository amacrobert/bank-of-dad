package repositories

import (
	"fmt"
	"testing"
	"time"

	"bank-of-dad/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthEventRepo_Log(t *testing.T) {
	db := testDB(t)
	repo := NewAuthEventRepo(db)

	event := models.AuthEvent{
		EventType: "login_success",
		UserType:  "parent",
		UserID:    1,
		FamilyID:  10,
		IPAddress: "192.168.1.1",
		Details:   "logged in via Google",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	err := repo.Log(event)
	require.NoError(t, err)

	events, err := repo.GetEventsByFamily(10)
	require.NoError(t, err)
	require.Len(t, events, 1)

	got := events[0]
	assert.Equal(t, "login_success", got.EventType)
	assert.Equal(t, "parent", got.UserType)
	assert.Equal(t, int64(1), got.UserID)
	assert.Equal(t, int64(10), got.FamilyID)
	assert.Equal(t, "192.168.1.1", got.IPAddress)
	assert.Equal(t, "logged in via Google", got.Details)
	assert.NotZero(t, got.ID)
}

func TestAuthEventRepo_GetEventsByFamily(t *testing.T) {
	db := testDB(t)
	repo := NewAuthEventRepo(db)

	now := time.Now().UTC().Truncate(time.Second)

	// Insert events for family 1
	require.NoError(t, repo.Log(models.AuthEvent{
		EventType: "login_success",
		UserType:  "parent",
		UserID:    1,
		FamilyID:  1,
		IPAddress: "10.0.0.1",
		Details:   "family 1 login",
		CreatedAt: now,
	}))

	// Insert events for family 2
	require.NoError(t, repo.Log(models.AuthEvent{
		EventType: "account_created",
		UserType:  "child",
		UserID:    2,
		FamilyID:  2,
		IPAddress: "10.0.0.2",
		Details:   "family 2 account",
		CreatedAt: now,
	}))

	events, err := repo.GetEventsByFamily(1)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "login_success", events[0].EventType)
	assert.Equal(t, int64(1), events[0].FamilyID)

	events, err = repo.GetEventsByFamily(2)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "account_created", events[0].EventType)
	assert.Equal(t, int64(2), events[0].FamilyID)
}

func TestAuthEventRepo_GetEventsByFamily_Empty(t *testing.T) {
	db := testDB(t)
	repo := NewAuthEventRepo(db)

	events, err := repo.GetEventsByFamily(9999)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestAuthEventRepo_GetEventsByFamily_OrderedByCreatedAtDesc(t *testing.T) {
	db := testDB(t)
	repo := NewAuthEventRepo(db)

	base := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	require.NoError(t, repo.Log(models.AuthEvent{
		EventType: "login_success",
		UserType:  "parent",
		UserID:    1,
		FamilyID:  5,
		IPAddress: "10.0.0.1",
		Details:   "oldest",
		CreatedAt: base,
	}))

	require.NoError(t, repo.Log(models.AuthEvent{
		EventType: "logout",
		UserType:  "parent",
		UserID:    1,
		FamilyID:  5,
		IPAddress: "10.0.0.1",
		Details:   "middle",
		CreatedAt: base.Add(1 * time.Hour),
	}))

	require.NoError(t, repo.Log(models.AuthEvent{
		EventType: "login_failure",
		UserType:  "child",
		UserID:    2,
		FamilyID:  5,
		IPAddress: "10.0.0.1",
		Details:   "newest",
		CreatedAt: base.Add(2 * time.Hour),
	}))

	events, err := repo.GetEventsByFamily(5)
	require.NoError(t, err)
	require.Len(t, events, 3)

	assert.Equal(t, "newest", events[0].Details)
	assert.Equal(t, "middle", events[1].Details)
	assert.Equal(t, "oldest", events[2].Details)

	assert.True(t, events[0].CreatedAt.After(events[1].CreatedAt))
	assert.True(t, events[1].CreatedAt.After(events[2].CreatedAt))
}

func TestAuthEventRepo_ListByUser(t *testing.T) {
	db := testDB(t)
	repo := NewAuthEventRepo(db)

	base := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Log(models.AuthEvent{
			EventType: "login_success",
			UserType:  "parent",
			UserID:    1,
			FamilyID:  10,
			IPAddress: "10.0.0.1",
			Details:   fmt.Sprintf("event-%d", i),
			CreatedAt: base.Add(time.Duration(i) * time.Hour),
		}))
	}

	// With limit
	events, err := repo.ListByUser(10, 3)
	require.NoError(t, err)
	require.Len(t, events, 3)
	// Should be ordered DESC (newest first)
	assert.Equal(t, "event-4", events[0].Details)
	assert.Equal(t, "event-3", events[1].Details)
	assert.Equal(t, "event-2", events[2].Details)
}
