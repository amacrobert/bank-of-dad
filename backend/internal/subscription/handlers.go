package subscription

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"bank-of-dad/internal/auth"
	"bank-of-dad/repositories"

	"github.com/stripe/stripe-go/v82"
	portalsession "github.com/stripe/stripe-go/v82/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/price"
	stripesub "github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

type Handlers struct {
	familyRepo          *repositories.FamilyRepo
	parentRepo          *repositories.ParentRepo
	childRepo           *repositories.ChildRepo
	webhookEventRepo    *repositories.WebhookEventRepo
	stripeSecretKey     string
	stripeWebhookSecret string
	frontendURL         string
}

func NewHandlers(
	familyRepo *repositories.FamilyRepo,
	parentRepo *repositories.ParentRepo,
	childRepo *repositories.ChildRepo,
	webhookEventRepo *repositories.WebhookEventRepo,
	stripeSecretKey string,
	stripeWebhookSecret string,
	frontendURL string,
) *Handlers {
	return &Handlers{
		familyRepo:          familyRepo,
		parentRepo:          parentRepo,
		childRepo:           childRepo,
		webhookEventRepo:    webhookEventRepo,
		stripeSecretKey:     stripeSecretKey,
		stripeWebhookSecret: stripeWebhookSecret,
		frontendURL:         frontendURL,
	}
}

var validPriceLookupKeys = map[string]bool{
	"plus_monthly": true,
	"plus_annual":  true,
}

func (h *Handlers) HandleGetSubscription(w http.ResponseWriter, r *http.Request) {
	familyID := auth.GetFamilyID(r)

	info, err := h.familyRepo.GetSubscriptionByFamilyID(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if info == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Family not found"})
		return
	}

	var subscriptionStatus interface{}
	if info.SubscriptionStatus != nil {
		subscriptionStatus = *info.SubscriptionStatus
	}

	var currentPeriodEnd interface{}
	if info.SubscriptionCurrentPeriodEnd != nil {
		currentPeriodEnd = info.SubscriptionCurrentPeriodEnd.UTC().Format(time.RFC3339)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"account_type":         info.AccountType,
		"subscription_status":  subscriptionStatus,
		"current_period_end":   currentPeriodEnd,
		"cancel_at_period_end": info.SubscriptionCancelAtPeriodEnd,
	})
}

func (h *Handlers) HandleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	familyID := auth.GetFamilyID(r)
	userID := auth.GetUserID(r)

	var req struct {
		PriceLookupKey string `json:"price_lookup_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if !validPriceLookupKeys[req.PriceLookupKey] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid price lookup key"})
		return
	}

	// Check if family already has active subscription
	info, err := h.familyRepo.GetSubscriptionByFamilyID(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if info != nil && info.SubscriptionStatus != nil && *info.SubscriptionStatus == "active" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Family already has an active subscription"})
		return
	}

	// Get parent email for Stripe checkout
	parent, err := h.parentRepo.GetByID(userID)
	if err != nil || parent == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	// Look up the Stripe Price by lookup key
	stripe.Key = h.stripeSecretKey

	priceParams := &stripe.PriceListParams{}
	priceParams.LookupKeys = []*string{stripe.String(req.PriceLookupKey)}
	priceIter := price.List(priceParams)

	var priceID string
	for priceIter.Next() {
		priceID = priceIter.Price().ID
		break
	}
	if err := priceIter.Err(); err != nil {
		log.Printf("Error looking up price: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to look up price"})
		return
	}
	if priceID == "" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Price not found for lookup key"})
		return
	}

	successURL := fmt.Sprintf("%s/settings/subscription?success=true", h.frontendURL)
	cancelURL := fmt.Sprintf("%s/settings/subscription?canceled=true", h.frontendURL)

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		CustomerEmail:     stripe.String(parent.Email),
		ClientReferenceID: stripe.String(strconv.FormatInt(familyID, 10)),
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
	}

	session, err := checkoutsession.New(params)
	if err != nil {
		log.Printf("Error creating checkout session: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create checkout session"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"checkout_url": session.URL})
}

func (h *Handlers) HandleCreatePortal(w http.ResponseWriter, r *http.Request) {
	familyID := auth.GetFamilyID(r)

	info, err := h.familyRepo.GetSubscriptionByFamilyID(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if info == nil || info.StripeCustomerID == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No active subscription to manage"})
		return
	}

	stripe.Key = h.stripeSecretKey

	returnURL := fmt.Sprintf("%s/settings/subscription", h.frontendURL)
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(*info.StripeCustomerID),
		ReturnURL: stripe.String(returnURL),
	}

	session, err := portalsession.New(params)
	if err != nil {
		log.Printf("Error creating portal session: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create portal session"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"portal_url": session.URL})
}

func (h *Handlers) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 65536))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to read body"})
		return
	}

	event, err := webhook.ConstructEventWithOptions(body, r.Header.Get("Stripe-Signature"), h.stripeWebhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid signature"})
		return
	}

	// Idempotency check
	processed, err := h.webhookEventRepo.HasBeenProcessed(event.ID)
	if err != nil {
		log.Printf("Error checking webhook idempotency: %v", err)
		writeJSON(w, http.StatusOK, map[string]string{"status": "error"})
		return
	}
	if processed {
		writeJSON(w, http.StatusOK, map[string]string{"status": "already_processed"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutCompleted(event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(event)
	case "invoice.payment_failed":
		h.handlePaymentFailed(event)
	default:
		log.Printf("Unrecognized webhook event type: %s", event.Type)
	}

	// Mark event as processed
	if err := h.webhookEventRepo.RecordEvent(event.ID, string(event.Type)); err != nil {
		log.Printf("Error marking webhook event as processed: %v", err)
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// subscriptionRaw is used to parse subscription fields from raw webhook JSON,
// since stripe-go v82 moved current_period_end to the item level in the Go struct
// but the Stripe API still sends it at the subscription level in webhook events.
type subscriptionRaw struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	CurrentPeriodEnd  int64  `json:"current_period_end"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
	Customer          string `json:"customer"`
	Items             struct {
		Data []struct {
			CurrentPeriodEnd int64 `json:"current_period_end"`
		} `json:"data"`
	} `json:"items"`
}

func (s *subscriptionRaw) getPeriodEnd() int64 {
	if s.CurrentPeriodEnd != 0 {
		return s.CurrentPeriodEnd
	}
	if len(s.Items.Data) > 0 && s.Items.Data[0].CurrentPeriodEnd != 0 {
		return s.Items.Data[0].CurrentPeriodEnd
	}
	return 0
}

// checkoutSessionRaw parses the checkout.session.completed event data.
type checkoutSessionRaw struct {
	ClientReferenceID string `json:"client_reference_id"`
	Customer          string `json:"customer"`
	Subscription      string `json:"subscription"`
}

func (h *Handlers) handleCheckoutCompleted(event stripe.Event) {
	var session checkoutSessionRaw
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Error parsing checkout session: %v", err)
		return
	}

	familyIDStr := session.ClientReferenceID
	if familyIDStr == "" {
		log.Printf("checkout.session.completed: missing client_reference_id")
		return
	}

	familyID, err := strconv.ParseInt(familyIDStr, 10, 64)
	if err != nil {
		log.Printf("checkout.session.completed: invalid client_reference_id: %s", familyIDStr)
		return
	}

	customerID := session.Customer
	subscriptionID := session.Subscription

	if customerID == "" || subscriptionID == "" {
		log.Printf("checkout.session.completed: missing customer or subscription ID")
		return
	}

	// Fetch the subscription from Stripe to get the current_period_end
	stripe.Key = h.stripeSecretKey
	sub, err := stripesub.Get(subscriptionID, nil)
	if err != nil {
		log.Printf("checkout.session.completed: error fetching subscription: %v", err)
		// Use current time + 30 days as fallback
		periodEnd := time.Now().UTC().Add(30 * 24 * time.Hour)
		if err := h.familyRepo.UpdateSubscriptionFromCheckout(familyID, customerID, subscriptionID, "active", periodEnd); err != nil {
			log.Printf("checkout.session.completed: error updating family: %v", err)
			return
		}
		// Enable all disabled children now that family is Plus
		if err := h.childRepo.EnableAllChildren(familyID); err != nil {
			log.Printf("checkout.session.completed: error enabling children for family %d: %v", familyID, err)
		}
		return
	}

	// Get current_period_end from the first item
	var periodEndUnix int64
	if sub.Items != nil && len(sub.Items.Data) > 0 {
		periodEndUnix = sub.Items.Data[0].CurrentPeriodEnd
	}
	if periodEndUnix == 0 {
		periodEndUnix = time.Now().UTC().Add(30 * 24 * time.Hour).Unix()
	}

	periodEnd := time.Unix(periodEndUnix, 0).UTC()
	if err := h.familyRepo.UpdateSubscriptionFromCheckout(familyID, customerID, subscriptionID, string(sub.Status), periodEnd); err != nil {
		log.Printf("checkout.session.completed: error updating family: %v", err)
		return
	}

	// Enable all disabled children now that family is Plus
	if err := h.childRepo.EnableAllChildren(familyID); err != nil {
		log.Printf("checkout.session.completed: error enabling children for family %d: %v", familyID, err)
	}
}

func (h *Handlers) handleSubscriptionUpdated(event stripe.Event) {
	var sub subscriptionRaw
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error parsing subscription: %v", err)
		return
	}

	periodEnd := time.Unix(sub.getPeriodEnd(), 0).UTC()
	if err := h.familyRepo.UpdateSubscriptionStatus(sub.ID, sub.Status, periodEnd, sub.CancelAtPeriodEnd); err != nil {
		log.Printf("customer.subscription.updated: family not found for subscription %s: %v", sub.ID, err)
	}
}

func (h *Handlers) handleSubscriptionDeleted(event stripe.Event) {
	var sub subscriptionRaw
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error parsing subscription: %v", err)
		return
	}

	// Look up family BEFORE clearing subscription (ClearSubscription NULLs the subscription ID)
	fam, err := h.familyRepo.GetFamilyByStripeSubscriptionID(sub.ID)
	if err != nil {
		log.Printf("customer.subscription.deleted: error looking up family for subscription %s: %v", sub.ID, err)
	}

	if err := h.familyRepo.ClearSubscription(sub.ID); err != nil {
		log.Printf("customer.subscription.deleted: family not found for subscription %s: %v", sub.ID, err)
		return
	}

	// Disable excess children now that family is back to free tier (limit: 2)
	if fam != nil {
		if err := h.childRepo.ReconcileChildLimits(fam.ID, 2); err != nil {
			log.Printf("customer.subscription.deleted: error reconciling children for family %d: %v", fam.ID, err)
		}
	}
}

type invoiceRaw struct {
	Subscription string `json:"subscription"`
	Customer     string `json:"customer"`
}

func (h *Handlers) handlePaymentFailed(event stripe.Event) {
	var invoice invoiceRaw
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error parsing invoice: %v", err)
		return
	}

	log.Printf("invoice.payment_failed: subscription=%s customer=%s", invoice.Subscription, invoice.Customer)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
