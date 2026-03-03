package store

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// SubscriptionInfo holds the subscription-related fields from the families table.
type SubscriptionInfo struct {
	AccountType                  string
	StripeCustomerID             sql.NullString
	StripeSubscriptionID         sql.NullString
	SubscriptionStatus           sql.NullString
	SubscriptionCurrentPeriodEnd sql.NullTime
	SubscriptionCancelAtPeriodEnd bool
}

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

type Family struct {
	ID        int64
	Slug      string
	Timezone  string
	CreatedAt time.Time
	AccountType                  string
	StripeCustomerID             sql.NullString
	StripeSubscriptionID         sql.NullString
	SubscriptionStatus           sql.NullString
	SubscriptionCurrentPeriodEnd sql.NullTime
	SubscriptionCancelAtPeriodEnd bool
}

type FamilyStore struct {
	db *sql.DB
}

func NewFamilyStore(db *sql.DB) *FamilyStore {
	return &FamilyStore{db: db}
}

func (s *FamilyStore) Create(slug string) (*Family, error) {
	var id int64
	err := s.db.QueryRow(`INSERT INTO families (slug) VALUES ($1) RETURNING id`, slug).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("slug already taken: %s", slug)
		}
		return nil, fmt.Errorf("insert family: %w", err)
	}
	return s.GetByID(id)
}

func (s *FamilyStore) GetByID(id int64) (*Family, error) {
	var f Family
	err := s.db.QueryRow(
		`SELECT id, slug, timezone, created_at, account_type, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, subscription_cancel_at_period_end FROM families WHERE id = $1`, id,
	).Scan(&f.ID, &f.Slug, &f.Timezone, &f.CreatedAt, &f.AccountType, &f.StripeCustomerID, &f.StripeSubscriptionID, &f.SubscriptionStatus, &f.SubscriptionCurrentPeriodEnd, &f.SubscriptionCancelAtPeriodEnd)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by id: %w", err)
	}
	return &f, nil
}

func (s *FamilyStore) GetBySlug(slug string) (*Family, error) {
	var f Family
	err := s.db.QueryRow(
		`SELECT id, slug, timezone, created_at, account_type, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, subscription_cancel_at_period_end FROM families WHERE slug = $1`, slug,
	).Scan(&f.ID, &f.Slug, &f.Timezone, &f.CreatedAt, &f.AccountType, &f.StripeCustomerID, &f.StripeSubscriptionID, &f.SubscriptionStatus, &f.SubscriptionCurrentPeriodEnd, &f.SubscriptionCancelAtPeriodEnd)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by slug: %w", err)
	}
	return &f, nil
}

func (s *FamilyStore) GetTimezone(familyID int64) (string, error) {
	var tz string
	err := s.db.QueryRow(
		`SELECT timezone FROM families WHERE id = $1`, familyID,
	).Scan(&tz)
	if err != nil {
		return "", fmt.Errorf("get timezone: %w", err)
	}
	return tz, nil
}

func (s *FamilyStore) UpdateTimezone(familyID int64, timezone string) error {
	result, err := s.db.Exec(
		`UPDATE families SET timezone = $1 WHERE id = $2`, timezone, familyID,
	)
	if err != nil {
		return fmt.Errorf("update timezone: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update timezone rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("family not found: %d", familyID)
	}
	return nil
}

func (s *FamilyStore) SlugExists(slug string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM families WHERE slug = $1`, slug,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check slug exists: %w", err)
	}
	return count > 0, nil
}

func (s *FamilyStore) SuggestSlugs(base string) []string {
	suggestions := []string{
		base + "-1",
		base + "-2",
		"the-" + base,
	}

	var available []string
	for _, sug := range suggestions {
		if ValidateSlug(sug) == nil {
			exists, err := s.SlugExists(sug)
			if err == nil && !exists {
				available = append(available, sug)
			}
		}
	}
	return available
}

func (s *FamilyStore) DeleteAll(familyID, parentID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	// Delete refresh tokens for all children in the family
	if _, err := tx.Exec(`DELETE FROM refresh_tokens WHERE user_type = 'child' AND user_id IN (SELECT id FROM children WHERE family_id = $1)`, familyID); err != nil {
		return fmt.Errorf("delete child refresh tokens: %w", err)
	}

	// Delete auth events for all children in the family
	if _, err := tx.Exec(`DELETE FROM auth_events WHERE user_type = 'child' AND user_id IN (SELECT id FROM children WHERE family_id = $1)`, familyID); err != nil {
		return fmt.Errorf("delete child auth events: %w", err)
	}

	// Delete children (cascades transactions, allowance_schedules, interest_schedules)
	if _, err := tx.Exec(`DELETE FROM children WHERE family_id = $1`, familyID); err != nil {
		return fmt.Errorf("delete children: %w", err)
	}

	// Delete refresh tokens for the parent
	if _, err := tx.Exec(`DELETE FROM refresh_tokens WHERE user_type = 'parent' AND user_id = $1`, parentID); err != nil {
		return fmt.Errorf("delete parent refresh tokens: %w", err)
	}

	// Delete auth events for the parent and family-scoped events
	if _, err := tx.Exec(`DELETE FROM auth_events WHERE (user_type = 'parent' AND user_id = $1) OR family_id = $2`, parentID, familyID); err != nil {
		return fmt.Errorf("delete parent auth events: %w", err)
	}

	// Delete the parent
	if _, err := tx.Exec(`DELETE FROM parents WHERE id = $1`, parentID); err != nil {
		return fmt.Errorf("delete parent: %w", err)
	}

	// Delete the family
	if _, err := tx.Exec(`DELETE FROM families WHERE id = $1`, familyID); err != nil {
		return fmt.Errorf("delete family: %w", err)
	}

	return tx.Commit()
}

// GetSubscriptionByFamilyID returns the subscription info for a family.
func (s *FamilyStore) GetSubscriptionByFamilyID(familyID int64) (*SubscriptionInfo, error) {
	var info SubscriptionInfo
	err := s.db.QueryRow(
		`SELECT account_type, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, subscription_cancel_at_period_end FROM families WHERE id = $1`, familyID,
	).Scan(&info.AccountType, &info.StripeCustomerID, &info.StripeSubscriptionID, &info.SubscriptionStatus, &info.SubscriptionCurrentPeriodEnd, &info.SubscriptionCancelAtPeriodEnd)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription by family id: %w", err)
	}
	return &info, nil
}

// GetFamilyByStripeCustomerID looks up a family by its Stripe Customer ID.
func (s *FamilyStore) GetFamilyByStripeCustomerID(customerID string) (*Family, error) {
	var f Family
	err := s.db.QueryRow(
		`SELECT id, slug, timezone, created_at, account_type, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, subscription_cancel_at_period_end FROM families WHERE stripe_customer_id = $1`, customerID,
	).Scan(&f.ID, &f.Slug, &f.Timezone, &f.CreatedAt, &f.AccountType, &f.StripeCustomerID, &f.StripeSubscriptionID, &f.SubscriptionStatus, &f.SubscriptionCurrentPeriodEnd, &f.SubscriptionCancelAtPeriodEnd)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by stripe customer id: %w", err)
	}
	return &f, nil
}

// GetFamilyByStripeSubscriptionID looks up a family by its Stripe Subscription ID.
func (s *FamilyStore) GetFamilyByStripeSubscriptionID(subscriptionID string) (*Family, error) {
	var f Family
	err := s.db.QueryRow(
		`SELECT id, slug, timezone, created_at, account_type, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, subscription_cancel_at_period_end FROM families WHERE stripe_subscription_id = $1`, subscriptionID,
	).Scan(&f.ID, &f.Slug, &f.Timezone, &f.CreatedAt, &f.AccountType, &f.StripeCustomerID, &f.StripeSubscriptionID, &f.SubscriptionStatus, &f.SubscriptionCurrentPeriodEnd, &f.SubscriptionCancelAtPeriodEnd)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by stripe subscription id: %w", err)
	}
	return &f, nil
}

// UpdateSubscriptionFromCheckout updates the family's subscription fields after a successful Stripe Checkout.
func (s *FamilyStore) UpdateSubscriptionFromCheckout(familyID int64, stripeCustomerID, stripeSubscriptionID, status string, periodEnd time.Time) error {
	result, err := s.db.Exec(
		`UPDATE families SET account_type = 'plus', stripe_customer_id = $1, stripe_subscription_id = $2, subscription_status = $3, subscription_current_period_end = $4 WHERE id = $5`,
		stripeCustomerID, stripeSubscriptionID, status, periodEnd, familyID,
	)
	if err != nil {
		return fmt.Errorf("update subscription from checkout: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update subscription from checkout rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("family not found: %d", familyID)
	}
	return nil
}

// UpdateSubscriptionStatus syncs subscription status from a Stripe webhook event.
func (s *FamilyStore) UpdateSubscriptionStatus(stripeSubscriptionID, status string, periodEnd time.Time, cancelAtPeriodEnd bool) error {
	result, err := s.db.Exec(
		`UPDATE families SET subscription_status = $1, subscription_current_period_end = $2, subscription_cancel_at_period_end = $3 WHERE stripe_subscription_id = $4`,
		status, periodEnd, cancelAtPeriodEnd, stripeSubscriptionID,
	)
	if err != nil {
		return fmt.Errorf("update subscription status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update subscription status rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("subscription not found: %s", stripeSubscriptionID)
	}
	return nil
}

// ClearSubscription resets a family back to free and NULLs all subscription fields.
func (s *FamilyStore) ClearSubscription(stripeSubscriptionID string) error {
	result, err := s.db.Exec(
		`UPDATE families SET account_type = 'free', stripe_customer_id = NULL, stripe_subscription_id = NULL, subscription_status = NULL, subscription_current_period_end = NULL, subscription_cancel_at_period_end = FALSE WHERE stripe_subscription_id = $1`,
		stripeSubscriptionID,
	)
	if err != nil {
		return fmt.Errorf("clear subscription: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("clear subscription rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("subscription not found: %s", stripeSubscriptionID)
	}
	return nil
}

func ValidateSlug(slug string) error {
	if len(slug) < 3 || len(slug) > 30 {
		return fmt.Errorf("slug must be between 3 and 30 characters")
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen")
	}
	return nil
}
