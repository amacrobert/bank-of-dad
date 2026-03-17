package repositories

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFamily(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	f, err := fr.Create("smith-family")
	require.NoError(t, err)
	assert.Equal(t, "smith-family", f.Slug)
	assert.NotZero(t, f.ID)
	assert.False(t, f.CreatedAt.IsZero())
}

func TestCreateFamily_DuplicateSlug(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	_, err := fr.Create("smith-family")
	require.NoError(t, err)

	_, err = fr.Create("smith-family")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slug already taken")
}

func TestGetBySlug(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	created, err := fr.Create("jones-bank")
	require.NoError(t, err)

	found, err := fr.GetBySlug("jones-bank")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "jones-bank", found.Slug)
}

func TestGetBySlug_NotFound(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	found, err := fr.GetBySlug("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestSlugExists(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	exists, err := fr.SlugExists("test-slug")
	require.NoError(t, err)
	assert.False(t, exists)

	_, err = fr.Create("test-slug")
	require.NoError(t, err)

	exists, err = fr.SlugExists("test-slug")
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
		{"ab", false},                                // too short
		{"a", false},                                 // too short
		{"", false},                                  // empty
		{"-abc", false},                              // starts with hyphen
		{"abc-", false},                              // ends with hyphen
		{"ABC", false},                               // uppercase
		{"abc def", false},                           // space
		{"abc_def", false},                           // underscore
		{"abcdefghij1234567890abcdefghij1", false},   // > 30 chars
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
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("tz-family")
	require.NoError(t, err)

	tz, err := fr.GetTimezone(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", tz)
}

func TestGetTimezone_NotFound(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	_, err := fr.GetTimezone(99999)
	assert.Error(t, err)
}

func TestUpdateTimezone(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("tz-update-family")
	require.NoError(t, err)

	err = fr.UpdateTimezone(fam.ID, "America/Chicago")
	require.NoError(t, err)

	tz, err := fr.GetTimezone(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "America/Chicago", tz)
}

func TestUpdateTimezone_NonexistentFamily(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	err := fr.UpdateTimezone(99999, "America/Chicago")
	assert.Error(t, err)
}

func TestCreateFamily_HasDefaultTimezone(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("tz-default-family")
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", fam.Timezone)
}

func TestGetByID_IncludesTimezone(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	created, err := fr.Create("tz-getbyid-family")
	require.NoError(t, err)

	err = fr.UpdateTimezone(created.ID, "Europe/London")
	require.NoError(t, err)

	found, err := fr.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Europe/London", found.Timezone)
}

func TestSuggestSlugs(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	suggestions := fr.SuggestSlugs("smith-family")
	assert.NotEmpty(t, suggestions)
	for _, s := range suggestions {
		assert.NoError(t, ValidateSlug(s))
	}
}

func TestDeleteAll_RemovesEverything(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)
	pr := NewParentRepo(db)

	// Create family, parent
	fam, err := fr.Create("delete-test")
	require.NoError(t, err)

	parent, err := pr.Create("google-del-1", "del@test.com", "Del Parent")
	require.NoError(t, err)
	require.NoError(t, pr.SetFamilyID(parent.ID, fam.ID))

	// Create a child via raw SQL (to avoid depending on child repo)
	err = db.Exec(`INSERT INTO children (family_id, first_name, password_hash) VALUES (?, 'Kiddo', 'hash123')`, fam.ID).Error
	require.NoError(t, err)

	// Delete everything
	err = fr.DeleteAll(fam.ID, parent.ID)
	require.NoError(t, err)

	// Verify family gone
	found, err := fr.GetByID(fam.ID)
	require.NoError(t, err)
	assert.Nil(t, found)

	// Verify parent gone
	foundParent, err := pr.GetByID(parent.ID)
	require.NoError(t, err)
	assert.Nil(t, foundParent)

	// Verify child gone
	var childCount int64
	db.Raw(`SELECT COUNT(*) FROM children WHERE family_id = ?`, fam.ID).Scan(&childCount)
	assert.Equal(t, int64(0), childCount)
}

func TestDeleteAll_DoesNotAffectOtherFamilies(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)
	pr := NewParentRepo(db)

	// Family 1 (will be deleted)
	fam1, err := fr.Create("fam-delete")
	require.NoError(t, err)
	parent1, err := pr.Create("google-fam1", "fam1@test.com", "Parent One")
	require.NoError(t, err)
	require.NoError(t, pr.SetFamilyID(parent1.ID, fam1.ID))

	// Family 2 (should survive)
	fam2, err := fr.Create("fam-keep")
	require.NoError(t, err)
	parent2, err := pr.Create("google-fam2", "fam2@test.com", "Parent Two")
	require.NoError(t, err)
	require.NoError(t, pr.SetFamilyID(parent2.ID, fam2.ID))

	// Delete family 1
	err = fr.DeleteAll(fam1.ID, parent1.ID)
	require.NoError(t, err)

	// Family 2 still intact
	found, err := fr.GetByID(fam2.ID)
	require.NoError(t, err)
	assert.NotNil(t, found)

	foundParent, err := pr.GetByID(parent2.ID)
	require.NoError(t, err)
	assert.NotNil(t, foundParent)
}

func TestDeleteAll_EmptyFamily(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)
	pr := NewParentRepo(db)

	fam, err := fr.Create("empty-fam")
	require.NoError(t, err)
	parent, err := pr.Create("google-empty", "empty@test.com", "Empty Parent")
	require.NoError(t, err)
	require.NoError(t, pr.SetFamilyID(parent.ID, fam.ID))

	err = fr.DeleteAll(fam.ID, parent.ID)
	require.NoError(t, err)

	found, err := fr.GetByID(fam.ID)
	require.NoError(t, err)
	assert.Nil(t, found)

	foundParent, err := pr.GetByID(parent.ID)
	require.NoError(t, err)
	assert.Nil(t, foundParent)
}

// --- Subscription store tests ---

func TestGetSubscriptionByFamilyID_NewFamily(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("sub-test-free")
	require.NoError(t, err)

	info, err := fr.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "free", info.AccountType)
	assert.Nil(t, info.StripeCustomerID)
	assert.Nil(t, info.StripeSubscriptionID)
	assert.Nil(t, info.SubscriptionStatus)
	assert.Nil(t, info.SubscriptionCurrentPeriodEnd)
	assert.False(t, info.SubscriptionCancelAtPeriodEnd)
}

func TestUpdateSubscriptionFromCheckout(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("sub-checkout")
	require.NoError(t, err)

	periodEnd := time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)
	err = fr.UpdateSubscriptionFromCheckout(fam.ID, "cus_test123", "sub_test456", "active", periodEnd)
	require.NoError(t, err)

	info, err := fr.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "plus", info.AccountType)
	assert.NotNil(t, info.StripeCustomerID)
	assert.Equal(t, "cus_test123", *info.StripeCustomerID)
	assert.NotNil(t, info.StripeSubscriptionID)
	assert.Equal(t, "sub_test456", *info.StripeSubscriptionID)
	assert.NotNil(t, info.SubscriptionStatus)
	assert.Equal(t, "active", *info.SubscriptionStatus)
	assert.NotNil(t, info.SubscriptionCurrentPeriodEnd)
	assert.Equal(t, periodEnd, info.SubscriptionCurrentPeriodEnd.UTC())
	assert.False(t, info.SubscriptionCancelAtPeriodEnd)
}

func TestGetFamilyByStripeCustomerID(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("sub-cust-lookup")
	require.NoError(t, err)

	periodEnd := time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)
	err = fr.UpdateSubscriptionFromCheckout(fam.ID, "cus_lookup123", "sub_lookup456", "active", periodEnd)
	require.NoError(t, err)

	found, err := fr.GetFamilyByStripeCustomerID("cus_lookup123")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, fam.ID, found.ID)

	notFound, err := fr.GetFamilyByStripeCustomerID("cus_nonexistent")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestGetFamilyByStripeSubscriptionID(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("sub-sid-lookup")
	require.NoError(t, err)

	periodEnd := time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)
	err = fr.UpdateSubscriptionFromCheckout(fam.ID, "cus_sid123", "sub_sid456", "active", periodEnd)
	require.NoError(t, err)

	found, err := fr.GetFamilyByStripeSubscriptionID("sub_sid456")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, fam.ID, found.ID)

	notFound, err := fr.GetFamilyByStripeSubscriptionID("sub_nonexistent")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestGetBankName_Default(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("bank-name-family")
	require.NoError(t, err)

	name, err := fr.GetBankName(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "Dad", name)
}

func TestUpdateBankName(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("bank-name-update")
	require.NoError(t, err)

	err = fr.UpdateBankName(fam.ID, "Mom")
	require.NoError(t, err)

	name, err := fr.GetBankName(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "Mom", name)
}

func TestUpdateBankName_NonexistentFamily(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	err := fr.UpdateBankName(99999, "Mom")
	assert.Error(t, err)
}

func TestUpdateSubscriptionStatus(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("sub-status-update")
	require.NoError(t, err)

	periodEnd := time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)
	err = fr.UpdateSubscriptionFromCheckout(fam.ID, "cus_status1", "sub_status1", "active", periodEnd)
	require.NoError(t, err)

	newPeriodEnd := time.Date(2026, 4, 26, 0, 0, 0, 0, time.UTC)
	err = fr.UpdateSubscriptionStatus("sub_status1", "past_due", newPeriodEnd, true)
	require.NoError(t, err)

	info, err := fr.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "past_due", *info.SubscriptionStatus)
	assert.Equal(t, newPeriodEnd, info.SubscriptionCurrentPeriodEnd.UTC())
	assert.True(t, info.SubscriptionCancelAtPeriodEnd)
}

func TestUpdateSubscriptionStatus_NotFound(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	err := fr.UpdateSubscriptionStatus("sub_nonexistent", "active", time.Now(), false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscription not found")
}

func TestClearSubscription(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	fam, err := fr.Create("sub-clear")
	require.NoError(t, err)

	periodEnd := time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)
	err = fr.UpdateSubscriptionFromCheckout(fam.ID, "cus_clear1", "sub_clear1", "active", periodEnd)
	require.NoError(t, err)

	err = fr.ClearSubscription("sub_clear1")
	require.NoError(t, err)

	info, err := fr.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "free", info.AccountType)
	assert.Nil(t, info.StripeCustomerID)
	assert.Nil(t, info.StripeSubscriptionID)
	assert.Nil(t, info.SubscriptionStatus)
	assert.Nil(t, info.SubscriptionCurrentPeriodEnd)
	assert.False(t, info.SubscriptionCancelAtPeriodEnd)
}

func TestClearSubscription_NotFound(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	err := fr.ClearSubscription("sub_nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscription not found")
}

func TestUpdateSubscriptionFromCheckout_NonexistentFamily(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	err := fr.UpdateSubscriptionFromCheckout(99999, "cus_x", "sub_x", "active", time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "family not found")
}

func TestGetSubscriptionByFamilyID_NotFound(t *testing.T) {
	db := testDB(t)
	fr := NewFamilyRepo(db)

	info, err := fr.GetSubscriptionByFamilyID(99999)
	require.NoError(t, err)
	assert.Nil(t, info)
}
