package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogEvent(t *testing.T) {
	db := testDB(t)
	s := NewAuthEventStore(db)

	event := AuthEvent{
		EventType: "login_success",
		UserType:  "parent",
		UserID:    1,
		FamilyID:  10,
		IPAddress: "192.168.1.1",
		Details:   "logged in via Google",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	err := s.LogEvent(event)
	require.NoError(t, err)

	events, err := s.GetEventsByFamily(10)
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

func TestGetEventsByFamily(t *testing.T) {
	db := testDB(t)
	s := NewAuthEventStore(db)

	now := time.Now().UTC().Truncate(time.Second)

	// Insert events for family 1
	require.NoError(t, s.LogEvent(AuthEvent{
		EventType: "login_success",
		UserType:  "parent",
		UserID:    1,
		FamilyID:  1,
		IPAddress: "10.0.0.1",
		Details:   "family 1 login",
		CreatedAt: now,
	}))

	// Insert events for family 2
	require.NoError(t, s.LogEvent(AuthEvent{
		EventType: "account_created",
		UserType:  "child",
		UserID:    2,
		FamilyID:  2,
		IPAddress: "10.0.0.2",
		Details:   "family 2 account",
		CreatedAt: now,
	}))

	events, err := s.GetEventsByFamily(1)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "login_success", events[0].EventType)
	assert.Equal(t, int64(1), events[0].FamilyID)

	events, err = s.GetEventsByFamily(2)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "account_created", events[0].EventType)
	assert.Equal(t, int64(2), events[0].FamilyID)
}

func TestGetEventsByFamily_Empty(t *testing.T) {
	db := testDB(t)
	s := NewAuthEventStore(db)

	events, err := s.GetEventsByFamily(9999)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestGetEventsByFamily_OrderedByCreatedAtDesc(t *testing.T) {
	db := testDB(t)
	s := NewAuthEventStore(db)

	base := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	require.NoError(t, s.LogEvent(AuthEvent{
		EventType: "login_success",
		UserType:  "parent",
		UserID:    1,
		FamilyID:  5,
		IPAddress: "10.0.0.1",
		Details:   "oldest",
		CreatedAt: base,
	}))

	require.NoError(t, s.LogEvent(AuthEvent{
		EventType: "logout",
		UserType:  "parent",
		UserID:    1,
		FamilyID:  5,
		IPAddress: "10.0.0.1",
		Details:   "middle",
		CreatedAt: base.Add(1 * time.Hour),
	}))

	require.NoError(t, s.LogEvent(AuthEvent{
		EventType: "login_failure",
		UserType:  "child",
		UserID:    2,
		FamilyID:  5,
		IPAddress: "10.0.0.1",
		Details:   "newest",
		CreatedAt: base.Add(2 * time.Hour),
	}))

	events, err := s.GetEventsByFamily(5)
	require.NoError(t, err)
	require.Len(t, events, 3)

	assert.Equal(t, "newest", events[0].Details)
	assert.Equal(t, "middle", events[1].Details)
	assert.Equal(t, "oldest", events[2].Details)

	assert.True(t, events[0].CreatedAt.After(events[1].CreatedAt))
	assert.True(t, events[1].CreatedAt.After(events[2].CreatedAt))
}
