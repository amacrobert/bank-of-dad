package store

import (
	"testing"
	"time"

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

func TestGetTimezone_Default(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	fam, err := fs.Create("tz-family")
	require.NoError(t, err)

	tz, err := fs.GetTimezone(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", tz)
}

func TestGetTimezone_NotFound(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	_, err := fs.GetTimezone(99999)
	assert.Error(t, err)
}

func TestUpdateTimezone(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	fam, err := fs.Create("tz-update-family")
	require.NoError(t, err)

	err = fs.UpdateTimezone(fam.ID, "America/Chicago")
	require.NoError(t, err)

	tz, err := fs.GetTimezone(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "America/Chicago", tz)
}

func TestUpdateTimezone_NonexistentFamily(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	err := fs.UpdateTimezone(99999, "America/Chicago")
	assert.Error(t, err)
}

func TestCreateFamily_HasDefaultTimezone(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	fam, err := fs.Create("tz-default-family")
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", fam.Timezone)
}

func TestGetByID_IncludesTimezone(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	created, err := fs.Create("tz-getbyid-family")
	require.NoError(t, err)

	err = fs.UpdateTimezone(created.ID, "Europe/London")
	require.NoError(t, err)

	found, err := fs.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Europe/London", found.Timezone)
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

func TestDeleteAll_RemovesEverything(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)
	ps := NewParentStore(db)
	cs := NewChildStore(db)
	ts := NewTransactionStore(db)
	ss := NewScheduleStore(db)
	rs := NewRefreshTokenStore(db)
	es := NewAuthEventStore(db)

	// Create family, parent, child
	fam, err := fs.Create("delete-test")
	require.NoError(t, err)

	parent, err := ps.Create("google-del-1", "del@test.com", "Del Parent")
	require.NoError(t, err)
	require.NoError(t, ps.SetFamilyID(parent.ID, fam.ID))

	child, err := cs.Create(fam.ID, "Kiddo", "password123", nil)
	require.NoError(t, err)

	// Create transaction
	_, _, err = ts.Deposit(child.ID, parent.ID, 1000, "test deposit")
	require.NoError(t, err)

	// Create allowance schedule
	dow := 1
	_, err = ss.Create(&AllowanceSchedule{
		ChildID:     child.ID,
		ParentID:    parent.ID,
		AmountCents: 500,
		Frequency:   "weekly",
		DayOfWeek:   &dow,
		Status:      ScheduleStatusActive,
	})
	require.NoError(t, err)

	// Create refresh tokens for parent and child
	_, err = rs.Create("parent", parent.ID, fam.ID, 24*time.Hour)
	require.NoError(t, err)
	_, err = rs.Create("child", child.ID, fam.ID, 24*time.Hour)
	require.NoError(t, err)

	// Create auth events for parent and child
	require.NoError(t, es.LogEvent(AuthEvent{
		EventType: "login", UserType: "parent", UserID: parent.ID,
		FamilyID: fam.ID, IPAddress: "127.0.0.1", CreatedAt: time.Now().UTC(),
	}))
	require.NoError(t, es.LogEvent(AuthEvent{
		EventType: "login", UserType: "child", UserID: child.ID,
		FamilyID: fam.ID, IPAddress: "127.0.0.1", CreatedAt: time.Now().UTC(),
	}))

	// Delete everything
	err = fs.DeleteAll(fam.ID, parent.ID)
	require.NoError(t, err)

	// Verify family gone
	found, err := fs.GetByID(fam.ID)
	require.NoError(t, err)
	assert.Nil(t, found)

	// Verify parent gone
	foundParent, err := ps.GetByID(parent.ID)
	require.NoError(t, err)
	assert.Nil(t, foundParent)

	// Verify child gone
	foundChild, err := cs.GetByID(child.ID)
	require.NoError(t, err)
	assert.Nil(t, foundChild)

	// Verify transactions gone
	txns, err := ts.ListByChild(child.ID)
	require.NoError(t, err)
	assert.Empty(t, txns)

	// Verify refresh tokens gone
	var tokenCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens WHERE family_id = $1`, fam.ID).Scan(&tokenCount)
	require.NoError(t, err)
	assert.Equal(t, 0, tokenCount)

	// Verify auth events gone
	var eventCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM auth_events WHERE family_id = $1`, fam.ID).Scan(&eventCount)
	require.NoError(t, err)
	assert.Equal(t, 0, eventCount)
}

func TestDeleteAll_DoesNotAffectOtherFamilies(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)
	ps := NewParentStore(db)
	cs := NewChildStore(db)
	ts := NewTransactionStore(db)

	// Family 1 (will be deleted)
	fam1, err := fs.Create("fam-delete")
	require.NoError(t, err)
	parent1, err := ps.Create("google-fam1", "fam1@test.com", "Parent One")
	require.NoError(t, err)
	require.NoError(t, ps.SetFamilyID(parent1.ID, fam1.ID))
	child1, err := cs.Create(fam1.ID, "Child1", "password123", nil)
	require.NoError(t, err)
	_, _, err = ts.Deposit(child1.ID, parent1.ID, 1000, "deposit")
	require.NoError(t, err)

	// Family 2 (should survive)
	fam2, err := fs.Create("fam-keep")
	require.NoError(t, err)
	parent2, err := ps.Create("google-fam2", "fam2@test.com", "Parent Two")
	require.NoError(t, err)
	require.NoError(t, ps.SetFamilyID(parent2.ID, fam2.ID))
	child2, err := cs.Create(fam2.ID, "Child2", "password123", nil)
	require.NoError(t, err)
	_, _, err = ts.Deposit(child2.ID, parent2.ID, 2000, "deposit")
	require.NoError(t, err)

	// Delete family 1
	err = fs.DeleteAll(fam1.ID, parent1.ID)
	require.NoError(t, err)

	// Family 2 still intact
	found, err := fs.GetByID(fam2.ID)
	require.NoError(t, err)
	assert.NotNil(t, found)

	foundParent, err := ps.GetByID(parent2.ID)
	require.NoError(t, err)
	assert.NotNil(t, foundParent)

	foundChild, err := cs.GetByID(child2.ID)
	require.NoError(t, err)
	assert.NotNil(t, foundChild)

	txns, err := ts.ListByChild(child2.ID)
	require.NoError(t, err)
	assert.Len(t, txns, 1)
}

func TestDeleteAll_EmptyFamily(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)
	ps := NewParentStore(db)

	fam, err := fs.Create("empty-fam")
	require.NoError(t, err)
	parent, err := ps.Create("google-empty", "empty@test.com", "Empty Parent")
	require.NoError(t, err)
	require.NoError(t, ps.SetFamilyID(parent.ID, fam.ID))

	err = fs.DeleteAll(fam.ID, parent.ID)
	require.NoError(t, err)

	found, err := fs.GetByID(fam.ID)
	require.NoError(t, err)
	assert.Nil(t, found)

	foundParent, err := ps.GetByID(parent.ID)
	require.NoError(t, err)
	assert.Nil(t, foundParent)
}

// --- Subscription store tests (024-stripe-subscription) ---

func TestGetSubscriptionByFamilyID_NewFamily(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	fam, err := fs.Create("sub-test-free")
	require.NoError(t, err)

	info, err := fs.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "free", info.AccountType)
	assert.False(t, info.StripeCustomerID.Valid)
	assert.False(t, info.StripeSubscriptionID.Valid)
	assert.False(t, info.SubscriptionStatus.Valid)
	assert.False(t, info.SubscriptionCurrentPeriodEnd.Valid)
	assert.False(t, info.SubscriptionCancelAtPeriodEnd)
}

func TestUpdateSubscriptionFromCheckout(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	fam, err := fs.Create("sub-checkout")
	require.NoError(t, err)

	periodEnd := time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)
	err = fs.UpdateSubscriptionFromCheckout(fam.ID, "cus_test123", "sub_test456", "active", periodEnd)
	require.NoError(t, err)

	info, err := fs.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "plus", info.AccountType)
	assert.Equal(t, "cus_test123", info.StripeCustomerID.String)
	assert.Equal(t, "sub_test456", info.StripeSubscriptionID.String)
	assert.Equal(t, "active", info.SubscriptionStatus.String)
	assert.True(t, info.SubscriptionCurrentPeriodEnd.Valid)
	assert.Equal(t, periodEnd, info.SubscriptionCurrentPeriodEnd.Time.UTC())
	assert.False(t, info.SubscriptionCancelAtPeriodEnd)
}

func TestGetFamilyByStripeCustomerID(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	fam, err := fs.Create("sub-cust-lookup")
	require.NoError(t, err)

	periodEnd := time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)
	err = fs.UpdateSubscriptionFromCheckout(fam.ID, "cus_lookup123", "sub_lookup456", "active", periodEnd)
	require.NoError(t, err)

	found, err := fs.GetFamilyByStripeCustomerID("cus_lookup123")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, fam.ID, found.ID)

	notFound, err := fs.GetFamilyByStripeCustomerID("cus_nonexistent")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestGetFamilyByStripeSubscriptionID(t *testing.T) {
	db := testDB(t)
	fs := NewFamilyStore(db)

	fam, err := fs.Create("sub-sid-lookup")
	require.NoError(t, err)

	periodEnd := time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)
	err = fs.UpdateSubscriptionFromCheckout(fam.ID, "cus_sid123", "sub_sid456", "active", periodEnd)
	require.NoError(t, err)

	found, err := fs.GetFamilyByStripeSubscriptionID("sub_sid456")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, fam.ID, found.ID)

	notFound, err := fs.GetFamilyByStripeSubscriptionID("sub_nonexistent")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}
