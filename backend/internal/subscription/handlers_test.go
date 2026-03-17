package subscription

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"bank-of-dad/internal/testutil"
	"bank-of-dad/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
)

func newTestHandlers(t *testing.T) (*Handlers, *repositories.FamilyRepo, *repositories.ParentRepo) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	familyRepo := repositories.NewFamilyRepo(db)
	parentRepo := repositories.NewParentRepo(db)
	childRepo := repositories.NewChildRepo(db)
	webhookEventRepo := repositories.NewWebhookEventRepo(db)
	h := NewHandlers(familyRepo, parentRepo, childRepo, webhookEventRepo, "sk_test_fake", "whsec_test_fake", "http://localhost:8000")
	return h, familyRepo, parentRepo
}

// T013: HandleGetSubscription returns free for new family
func TestHandleGetSubscription_FreeFamily(t *testing.T) {
	h, familyRepo, parentRepo := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-free-handler")
	require.NoError(t, err)
	parent, err := parentRepo.Create("google-sub-1", "sub1@test.com", "Sub Parent")
	require.NoError(t, err)
	require.NoError(t, parentRepo.SetFamilyID(parent.ID, fam.ID))

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
	h, familyRepo, parentRepo := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-plus-handler")
	require.NoError(t, err)
	parent, err := parentRepo.Create("google-sub-2", "sub2@test.com", "Plus Parent")
	require.NoError(t, err)
	require.NoError(t, parentRepo.SetFamilyID(parent.ID, fam.ID))

	// Simulate checkout completion
	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyRepo.UpdateSubscriptionFromCheckout(fam.ID, "cus_handler1", "sub_handler1", "active", periodEnd))

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
	h, familyRepo, parentRepo := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-already")
	require.NoError(t, err)
	parent, err := parentRepo.Create("google-sub-3", "sub3@test.com", "Already Sub")
	require.NoError(t, err)
	require.NoError(t, parentRepo.SetFamilyID(parent.ID, fam.ID))

	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyRepo.UpdateSubscriptionFromCheckout(fam.ID, "cus_already1", "sub_already1", "active", periodEnd))

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
	h, familyRepo, parentRepo := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-invalid-key")
	require.NoError(t, err)
	parent, err := parentRepo.Create("google-sub-4", "sub4@test.com", "Invalid Key")
	require.NoError(t, err)
	require.NoError(t, parentRepo.SetFamilyID(parent.ID, fam.ID))

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
	h, familyRepo, parentRepo := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-no-portal")
	require.NoError(t, err)
	parent, err := parentRepo.Create("google-sub-5", "sub5@test.com", "No Portal")
	require.NoError(t, err)
	require.NoError(t, parentRepo.SetFamilyID(parent.ID, fam.ID))

	req := httptest.NewRequest("POST", "/api/subscription/portal", nil)
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	w := httptest.NewRecorder()

	h.HandleCreatePortal(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "No active subscription to manage")
}

// handleSubscriptionUpdated updates status, period end, and cancel_at_period_end
func TestHandleSubscriptionUpdated(t *testing.T) {
	h, familyRepo, _ := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-wh-updated")
	require.NoError(t, err)

	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyRepo.UpdateSubscriptionFromCheckout(fam.ID, "cus_wh_upd1", "sub_wh_upd1", "active", periodEnd))

	// Simulate a subscription.updated webhook event
	newPeriodEnd := int64(1777420800) // 2026-04-26T00:00:00Z
	eventData := json.RawMessage(`{
		"id": "sub_wh_upd1",
		"status": "active",
		"current_period_end": ` + strconv.FormatInt(newPeriodEnd, 10) + `,
		"cancel_at_period_end": true,
		"customer": "cus_wh_upd1"
	}`)

	event := stripe.Event{
		Type: "customer.subscription.updated",
		Data: &stripe.EventData{Raw: eventData},
	}

	h.handleSubscriptionUpdated(event)

	info, err := familyRepo.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "plus", info.AccountType)
	require.NotNil(t, info.SubscriptionStatus)
	assert.Equal(t, "active", *info.SubscriptionStatus)
	assert.True(t, info.SubscriptionCancelAtPeriodEnd)
	require.NotNil(t, info.SubscriptionCurrentPeriodEnd)
	assert.Equal(t, time.Unix(newPeriodEnd, 0).UTC(), info.SubscriptionCurrentPeriodEnd.UTC())
}

// handleSubscriptionUpdated falls back to items-level current_period_end
func TestHandleSubscriptionUpdated_ItemsLevelPeriodEnd(t *testing.T) {
	h, familyRepo, _ := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-wh-items")
	require.NoError(t, err)

	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyRepo.UpdateSubscriptionFromCheckout(fam.ID, "cus_wh_items1", "sub_wh_items1", "active", periodEnd))

	// Simulate webhook where current_period_end is 0 at subscription level
	// but present in items.data[0].current_period_end
	itemPeriodEnd := int64(1777420800) // 2026-04-26T00:00:00Z
	eventData := json.RawMessage(`{
		"id": "sub_wh_items1",
		"status": "active",
		"current_period_end": 0,
		"cancel_at_period_end": true,
		"customer": "cus_wh_items1",
		"items": {
			"data": [{"current_period_end": ` + strconv.FormatInt(itemPeriodEnd, 10) + `}]
		}
	}`)

	event := stripe.Event{
		Type: "customer.subscription.updated",
		Data: &stripe.EventData{Raw: eventData},
	}

	h.handleSubscriptionUpdated(event)

	info, err := familyRepo.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	require.NotNil(t, info.SubscriptionStatus)
	assert.Equal(t, "active", *info.SubscriptionStatus)
	assert.True(t, info.SubscriptionCancelAtPeriodEnd)
	require.NotNil(t, info.SubscriptionCurrentPeriodEnd)
	assert.Equal(t, time.Unix(itemPeriodEnd, 0).UTC(), info.SubscriptionCurrentPeriodEnd.UTC())
}

// handleSubscriptionDeleted clears all subscription fields
func TestHandleSubscriptionDeleted(t *testing.T) {
	h, familyRepo, _ := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-wh-deleted")
	require.NoError(t, err)

	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyRepo.UpdateSubscriptionFromCheckout(fam.ID, "cus_wh_del1", "sub_wh_del1", "active", periodEnd))

	eventData := json.RawMessage(`{
		"id": "sub_wh_del1",
		"status": "canceled",
		"current_period_end": 0,
		"cancel_at_period_end": false,
		"customer": "cus_wh_del1"
	}`)

	event := stripe.Event{
		Type: "customer.subscription.deleted",
		Data: &stripe.EventData{Raw: eventData},
	}

	h.handleSubscriptionDeleted(event)

	info, err := familyRepo.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "free", info.AccountType)
	assert.Nil(t, info.StripeSubscriptionID)
	assert.Nil(t, info.SubscriptionStatus)
}

// handleCheckoutCompleted falls back to 30-day period when Stripe API call fails
func TestHandleCheckoutCompleted_FallbackPeriod(t *testing.T) {
	h, familyRepo, _ := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-wh-checkout")
	require.NoError(t, err)

	eventData := json.RawMessage(`{
		"client_reference_id": "` + strconv.FormatInt(fam.ID, 10) + `",
		"customer": "cus_wh_co1",
		"subscription": "sub_wh_co1"
	}`)

	event := stripe.Event{
		Type: "checkout.session.completed",
		Data: &stripe.EventData{Raw: eventData},
	}

	// Stripe API call will fail with fake key, triggering fallback
	h.handleCheckoutCompleted(event)

	info, err := familyRepo.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "plus", info.AccountType)
	require.NotNil(t, info.SubscriptionStatus)
	assert.Equal(t, "active", *info.SubscriptionStatus)
	require.NotNil(t, info.StripeCustomerID)
	assert.Equal(t, "cus_wh_co1", *info.StripeCustomerID)
	require.NotNil(t, info.StripeSubscriptionID)
	assert.Equal(t, "sub_wh_co1", *info.StripeSubscriptionID)
	// Fallback period should be ~30 days from now
	require.NotNil(t, info.SubscriptionCurrentPeriodEnd)
	expectedApprox := time.Now().UTC().Add(30 * 24 * time.Hour)
	assert.InDelta(t, expectedApprox.Unix(), info.SubscriptionCurrentPeriodEnd.Unix(), 60)
}

// handleCheckoutCompleted ignores events with missing fields
func TestHandleCheckoutCompleted_MissingFields(t *testing.T) {
	h, familyRepo, _ := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-wh-missing")
	require.NoError(t, err)

	// Missing client_reference_id — should be a no-op
	eventData := json.RawMessage(`{
		"client_reference_id": "",
		"customer": "cus_wh_miss1",
		"subscription": "sub_wh_miss1"
	}`)

	event := stripe.Event{
		Type: "checkout.session.completed",
		Data: &stripe.EventData{Raw: eventData},
	}

	h.handleCheckoutCompleted(event)

	info, err := familyRepo.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "free", info.AccountType) // unchanged
}

// handlePaymentFailed logs without error
func TestHandlePaymentFailed(t *testing.T) {
	h, _, _ := newTestHandlers(t)

	eventData := json.RawMessage(`{
		"subscription": "sub_fail1",
		"customer": "cus_fail1"
	}`)

	event := stripe.Event{
		Type: "invoice.payment_failed",
		Data: &stripe.EventData{Raw: eventData},
	}

	// Should not panic — just logs
	h.handlePaymentFailed(event)
}

// HandleCreateCheckout returns 400 for invalid JSON body
func TestHandleCreateCheckout_InvalidBody(t *testing.T) {
	h, familyRepo, parentRepo := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-bad-body")
	require.NoError(t, err)
	parent, err := parentRepo.Create("google-sub-bad", "bad@test.com", "Bad Body")
	require.NoError(t, err)
	require.NoError(t, parentRepo.SetFamilyID(parent.ID, fam.ID))

	req := httptest.NewRequest("POST", "/api/subscription/checkout", strings.NewReader("not json"))
	req = testutil.SetRequestContext(req, "parent", parent.ID, fam.ID)
	w := httptest.NewRecorder()

	h.HandleCreateCheckout(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "Invalid request body")
}

// T027: UpdateSubscriptionStatus and ClearSubscription repo tests
func TestUpdateSubscriptionStatus(t *testing.T) {
	_, familyRepo, _ := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-status-update")
	require.NoError(t, err)

	// Set up subscription first
	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyRepo.UpdateSubscriptionFromCheckout(fam.ID, "cus_status1", "sub_status1", "active", periodEnd))

	// Update subscription status
	newPeriodEnd := mustParseTime("2026-04-26T00:00:00Z")
	err = familyRepo.UpdateSubscriptionStatus("sub_status1", "active", newPeriodEnd, true)
	require.NoError(t, err)

	info, err := familyRepo.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	require.NotNil(t, info.SubscriptionStatus)
	assert.Equal(t, "active", *info.SubscriptionStatus)
	require.NotNil(t, info.SubscriptionCurrentPeriodEnd)
	assert.Equal(t, newPeriodEnd, info.SubscriptionCurrentPeriodEnd.UTC())
	assert.True(t, info.SubscriptionCancelAtPeriodEnd)
}

func TestClearSubscription(t *testing.T) {
	_, familyRepo, _ := newTestHandlers(t)

	fam, err := familyRepo.Create("sub-clear")
	require.NoError(t, err)

	// Set up subscription first
	periodEnd := mustParseTime("2026-03-26T00:00:00Z")
	require.NoError(t, familyRepo.UpdateSubscriptionFromCheckout(fam.ID, "cus_clear1", "sub_clear1", "active", periodEnd))

	// Clear subscription
	err = familyRepo.ClearSubscription("sub_clear1")
	require.NoError(t, err)

	info, err := familyRepo.GetSubscriptionByFamilyID(fam.ID)
	require.NoError(t, err)
	assert.Equal(t, "free", info.AccountType)
	assert.Nil(t, info.StripeCustomerID)
	assert.Nil(t, info.StripeSubscriptionID)
	assert.Nil(t, info.SubscriptionStatus)
	assert.Nil(t, info.SubscriptionCurrentPeriodEnd)
	assert.False(t, info.SubscriptionCancelAtPeriodEnd)
}
