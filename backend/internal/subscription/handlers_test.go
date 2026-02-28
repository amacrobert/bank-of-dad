package subscription

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bank-of-dad/internal/store"
	"bank-of-dad/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHandlers(t *testing.T) (*Handlers, *store.FamilyStore, *store.ParentStore) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	familyStore := store.NewFamilyStore(db)
	parentStore := store.NewParentStore(db)
	webhookEventStore := store.NewWebhookEventStore(db)
	h := NewHandlers(familyStore, parentStore, webhookEventStore, "sk_test_fake", "whsec_test_fake", "http://localhost:8000")
	return h, familyStore, parentStore
}

// T013: HandleGetSubscription returns free for new family
func TestHandleGetSubscription_FreeFamily(t *testing.T) {
	h, familyStore, parentStore := newTestHandlers(t)

	fam, err := familyStore.Create("sub-free-handler")
	require.NoError(t, err)
	parent, err := parentStore.Create("google-sub-1", "sub1@test.com", "Sub Parent")
	require.NoError(t, err)
	require.NoError(t, parentStore.SetFamilyID(parent.ID, fam.ID))

	req := httptest.NewRequest("GET", "/api/subscription", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	w := httptest.NewRecorder()

	h.HandleGetSubscription(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "free", resp["account_type"])
	assert.Nil(t, resp["subscription_status"])
	assert.Nil(t, resp["current_period_end"])
	assert.Equal(t, false, resp["cancel_at_period_end"])
}

// T013: HandleGetSubscription returns subscription fields for subscribed family
func TestHandleGetSubscription_SubscribedFamily(t *testing.T) {
	h, familyStore, parentStore := newTestHandlers(t)

	fam, err := familyStore.Create("sub-plus-handler")
	require.NoError(t, err)
	parent, err := parentStore.Create("google-sub-2", "sub2@test.com", "Plus Parent")
	require.NoError(t, err)
	require.NoError(t, parentStore.SetFamilyID(parent.ID, fam.ID))

	// Simulate checkout completion
	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyStore.UpdateSubscriptionFromCheckout(fam.ID, "cus_handler1", "sub_handler1", "active", periodEnd))

	req := httptest.NewRequest("GET", "/api/subscription", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	w := httptest.NewRecorder()

	h.HandleGetSubscription(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "plus", resp["account_type"])
	assert.Equal(t, "active", resp["subscription_status"])
	assert.NotNil(t, resp["current_period_end"])
	assert.Equal(t, false, resp["cancel_at_period_end"])
}

// T014: HandleCreateCheckout returns 400 for already-subscribed family
func TestHandleCreateCheckout_AlreadySubscribed(t *testing.T) {
	h, familyStore, parentStore := newTestHandlers(t)

	fam, err := familyStore.Create("sub-already")
	require.NoError(t, err)
	parent, err := parentStore.Create("google-sub-3", "sub3@test.com", "Already Sub")
	require.NoError(t, err)
	require.NoError(t, parentStore.SetFamilyID(parent.ID, fam.ID))

	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyStore.UpdateSubscriptionFromCheckout(fam.ID, "cus_already1", "sub_already1", "active", periodEnd))

	body := `{"price_lookup_key":"plus_monthly"}`
	req := httptest.NewRequest("POST", "/api/subscription/checkout", strings.NewReader(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	w := httptest.NewRecorder()

	h.HandleCreateCheckout(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "already has an active subscription")
}

// T014: HandleCreateCheckout returns 400 for invalid price_lookup_key
func TestHandleCreateCheckout_InvalidPriceKey(t *testing.T) {
	h, familyStore, parentStore := newTestHandlers(t)

	fam, err := familyStore.Create("sub-invalid-key")
	require.NoError(t, err)
	parent, err := parentStore.Create("google-sub-4", "sub4@test.com", "Invalid Key")
	require.NoError(t, err)
	require.NoError(t, parentStore.SetFamilyID(parent.ID, fam.ID))

	body := `{"price_lookup_key":"invalid_key"}`
	req := httptest.NewRequest("POST", "/api/subscription/checkout", strings.NewReader(body))
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	w := httptest.NewRecorder()

	h.HandleCreateCheckout(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "Invalid price lookup key")
}

// T026: HandleCreatePortal returns 400 if family has no stripe_customer_id
func TestHandleCreatePortal_NoCustomer(t *testing.T) {
	h, familyStore, parentStore := newTestHandlers(t)

	fam, err := familyStore.Create("sub-no-portal")
	require.NoError(t, err)
	parent, err := parentStore.Create("google-sub-5", "sub5@test.com", "No Portal")
	require.NoError(t, err)
	require.NoError(t, parentStore.SetFamilyID(parent.ID, fam.ID))

	req := httptest.NewRequest("POST", "/api/subscription/portal", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	w := httptest.NewRecorder()

	h.HandleCreatePortal(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "No active subscription to manage")
}

// T027: UpdateSubscriptionStatus and ClearSubscription store tests
func TestUpdateSubscriptionStatus(t *testing.T) {
	_, familyStore, _ := newTestHandlers(t)

	fam, err := familyStore.Create("sub-status-update")
	require.NoError(t, err)

	// Set up subscription first
	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyStore.UpdateSubscriptionFromCheckout(fam.ID, "cus_status1", "sub_status1", "active", periodEnd))

	// Update subscription status
	newPeriodEnd := mustParseTime("2026-04-26T00:00:00Z")
	err = familyStore.UpdateSubscriptionStatus("sub_status1", "active", newPeriodEnd, true)
	require.NoError(t, err)

	info, err := familyStore.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", info.SubscriptionStatus.String)
	assert.Equal(t, newPeriodEnd, info.SubscriptionCurrentPeriodEnd.Time.UTC())
	assert.True(t, info.SubscriptionCancelAtPeriodEnd)
}

func TestClearSubscription(t *testing.T) {
	_, familyStore, _ := newTestHandlers(t)

	fam, err := familyStore.Create("sub-clear")
	require.NoError(t, err)

	// Set up subscription first
	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyStore.UpdateSubscriptionFromCheckout(fam.ID, "cus_clear1", "sub_clear1", "active", periodEnd))

	// Clear subscription
	err = familyStore.ClearSubscription("sub_clear1")
	require.NoError(t, err)

	info, err := familyStore.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "free", info.AccountType)
	assert.False(t, info.StripeCustomerID.Valid)
	assert.False(t, info.StripeSubscriptionID.Valid)
	assert.False(t, info.SubscriptionStatus.Valid)
	assert.False(t, info.SubscriptionCurrentPeriodEnd.Valid)
	assert.False(t, info.SubscriptionCancelAtPeriodEnd)
}
