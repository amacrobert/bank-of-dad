package repositories

import (
	"fmt"
	"time"

	"bank-of-dad/models"

	"errors"

	"gorm.io/gorm"
)

// SubscriptionInfo holds the subscription-related fields from the families table.
type SubscriptionInfo struct {
	AccountType                   string
	StripeCustomerID              *string
	StripeSubscriptionID          *string
	SubscriptionStatus            *string
	SubscriptionCurrentPeriodEnd  *time.Time
	SubscriptionCancelAtPeriodEnd bool
}

// FamilyRepo provides GORM-based access to the families table.
type FamilyRepo struct {
	db *gorm.DB
}

// NewFamilyRepo creates a new FamilyRepo.
func NewFamilyRepo(db *gorm.DB) *FamilyRepo {
	return &FamilyRepo{db: db}
}

// Create inserts a new family with the given slug.
func (r *FamilyRepo) Create(slug string) (*models.Family, error) {
	f := models.Family{Slug: slug}
	if err := r.db.Create(&f).Error; err != nil {
		if isDuplicateKey(err) {
			return nil, fmt.Errorf("slug already taken: %s", slug)
		}
		return nil, fmt.Errorf("insert family: %w", err)
	}
	return &f, nil
}

// GetByID retrieves a family by its ID. Returns (nil, nil) if not found.
func (r *FamilyRepo) GetByID(id int64) (*models.Family, error) {
	var f models.Family
	err := r.db.First(&f, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by id: %w", err)
	}
	return &f, nil
}

// GetBySlug retrieves a family by its slug. Returns (nil, nil) if not found.
func (r *FamilyRepo) GetBySlug(slug string) (*models.Family, error) {
	var f models.Family
	err := r.db.Where("slug = ?", slug).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by slug: %w", err)
	}
	return &f, nil
}

// GetTimezone returns the timezone for a family.
func (r *FamilyRepo) GetTimezone(familyID int64) (string, error) {
	var f models.Family
	err := r.db.Select("timezone").First(&f, familyID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("get timezone: %w", gorm.ErrRecordNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("get timezone: %w", err)
	}
	return f.Timezone, nil
}

// UpdateTimezone updates the timezone for a family.
func (r *FamilyRepo) UpdateTimezone(familyID int64, timezone string) error {
	result := r.db.Model(&models.Family{}).Where("id = ?", familyID).Update("timezone", timezone)
	if result.Error != nil {
		return fmt.Errorf("update timezone: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("family not found: %d", familyID)
	}
	return nil
}

// GetBankName returns the bank name for a family.
func (r *FamilyRepo) GetBankName(familyID int64) (string, error) {
	var f models.Family
	err := r.db.Select("bank_name").First(&f, familyID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("get bank name: %w", gorm.ErrRecordNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("get bank name: %w", err)
	}
	return f.BankName, nil
}

// UpdateBankName updates the bank name for a family.
func (r *FamilyRepo) UpdateBankName(familyID int64, bankName string) error {
	result := r.db.Model(&models.Family{}).Where("id = ?", familyID).Update("bank_name", bankName)
	if result.Error != nil {
		return fmt.Errorf("update bank name: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("family not found: %d", familyID)
	}
	return nil
}

// SlugExists checks whether a slug is already in use.
func (r *FamilyRepo) SlugExists(slug string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Family{}).Where("slug = ?", slug).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check slug exists: %w", err)
	}
	return count > 0, nil
}

// SuggestSlugs returns available slug alternatives based on a base string.
func (r *FamilyRepo) SuggestSlugs(base string) []string {
	candidates := []string{
		base + "-1",
		base + "-2",
		"the-" + base,
	}

	// Filter to only valid slugs
	var valid []string
	for _, sug := range candidates {
		if models.ValidateSlug(sug) == nil {
			valid = append(valid, sug)
		}
	}
	if len(valid) == 0 {
		return nil
	}

	// Batch query for existing slugs
	var existing []string
	if err := r.db.Model(&models.Family{}).Where("slug IN ?", valid).Pluck("slug", &existing).Error; err != nil {
		return nil
	}

	taken := make(map[string]bool, len(existing))
	for _, s := range existing {
		taken[s] = true
	}

	var available []string
	for _, sug := range valid {
		if !taken[sug] {
			available = append(available, sug)
		}
	}
	return available
}

// DeleteAll removes all data for a family and its parent within a transaction.
func (r *FamilyRepo) DeleteAll(familyID, parentID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete refresh tokens for all children in the family
		if err := tx.Exec(`DELETE FROM refresh_tokens WHERE user_type = 'child' AND user_id IN (SELECT id FROM children WHERE family_id = ?)`, familyID).Error; err != nil {
			return fmt.Errorf("delete child refresh tokens: %w", err)
		}

		// Delete auth events for all children in the family
		if err := tx.Exec(`DELETE FROM auth_events WHERE user_type = 'child' AND user_id IN (SELECT id FROM children WHERE family_id = ?)`, familyID).Error; err != nil {
			return fmt.Errorf("delete child auth events: %w", err)
		}

		// Delete children (cascades transactions, allowance_schedules, interest_schedules)
		if err := tx.Exec(`DELETE FROM children WHERE family_id = ?`, familyID).Error; err != nil {
			return fmt.Errorf("delete children: %w", err)
		}

		// Delete refresh tokens for the parent
		if err := tx.Exec(`DELETE FROM refresh_tokens WHERE user_type = 'parent' AND user_id = ?`, parentID).Error; err != nil {
			return fmt.Errorf("delete parent refresh tokens: %w", err)
		}

		// Delete auth events for the parent and family-scoped events
		if err := tx.Exec(`DELETE FROM auth_events WHERE (user_type = 'parent' AND user_id = ?) OR family_id = ?`, parentID, familyID).Error; err != nil {
			return fmt.Errorf("delete parent auth events: %w", err)
		}

		// Delete the parent
		if err := tx.Exec(`DELETE FROM parents WHERE id = ?`, parentID).Error; err != nil {
			return fmt.Errorf("delete parent: %w", err)
		}

		// Delete the family
		if err := tx.Exec(`DELETE FROM families WHERE id = ?`, familyID).Error; err != nil {
			return fmt.Errorf("delete family: %w", err)
		}

		return nil
	})
}

// GetSubscriptionByFamilyID returns the subscription info for a family.
func (r *FamilyRepo) GetSubscriptionByFamilyID(familyID int64) (*SubscriptionInfo, error) {
	var f models.Family
	err := r.db.Select("account_type, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, subscription_cancel_at_period_end").
		First(&f, familyID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription by family id: %w", err)
	}
	return &SubscriptionInfo{
		AccountType:                   f.AccountType,
		StripeCustomerID:              f.StripeCustomerID,
		StripeSubscriptionID:          f.StripeSubscriptionID,
		SubscriptionStatus:            f.SubscriptionStatus,
		SubscriptionCurrentPeriodEnd:  f.SubscriptionCurrentPeriodEnd,
		SubscriptionCancelAtPeriodEnd: f.SubscriptionCancelAtPeriodEnd,
	}, nil
}

// GetFamilyByStripeCustomerID looks up a family by its Stripe Customer ID.
func (r *FamilyRepo) GetFamilyByStripeCustomerID(customerID string) (*models.Family, error) {
	var f models.Family
	err := r.db.Where("stripe_customer_id = ?", customerID).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by stripe customer id: %w", err)
	}
	return &f, nil
}

// GetFamilyByStripeSubscriptionID looks up a family by its Stripe Subscription ID.
func (r *FamilyRepo) GetFamilyByStripeSubscriptionID(subscriptionID string) (*models.Family, error) {
	var f models.Family
	err := r.db.Where("stripe_subscription_id = ?", subscriptionID).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get family by stripe subscription id: %w", err)
	}
	return &f, nil
}

// UpdateSubscriptionFromCheckout updates the family's subscription fields after a successful Stripe Checkout.
func (r *FamilyRepo) UpdateSubscriptionFromCheckout(familyID int64, stripeCustomerID, stripeSubscriptionID, status string, periodEnd time.Time) error {
	result := r.db.Model(&models.Family{}).Where("id = ?", familyID).Updates(map[string]interface{}{
		"account_type":                    "plus",
		"stripe_customer_id":              stripeCustomerID,
		"stripe_subscription_id":          stripeSubscriptionID,
		"subscription_status":             status,
		"subscription_current_period_end": periodEnd,
	})
	if result.Error != nil {
		return fmt.Errorf("update subscription from checkout: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("family not found: %d", familyID)
	}
	return nil
}

// UpdateSubscriptionStatus syncs subscription status from a Stripe webhook event.
func (r *FamilyRepo) UpdateSubscriptionStatus(stripeSubscriptionID, status string, periodEnd time.Time, cancelAtPeriodEnd bool) error {
	result := r.db.Model(&models.Family{}).Where("stripe_subscription_id = ?", stripeSubscriptionID).Updates(map[string]interface{}{
		"subscription_status":               status,
		"subscription_current_period_end":   periodEnd,
		"subscription_cancel_at_period_end": cancelAtPeriodEnd,
	})
	if result.Error != nil {
		return fmt.Errorf("update subscription status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription not found: %s", stripeSubscriptionID)
	}
	return nil
}

// ClearSubscription resets a family back to free and NULLs all subscription fields.
func (r *FamilyRepo) ClearSubscription(stripeSubscriptionID string) error {
	result := r.db.Model(&models.Family{}).Where("stripe_subscription_id = ?", stripeSubscriptionID).Updates(map[string]interface{}{
		"account_type":                      "free",
		"stripe_customer_id":                nil,
		"stripe_subscription_id":            nil,
		"subscription_status":               nil,
		"subscription_current_period_end":   nil,
		"subscription_cancel_at_period_end": false,
	})
	if result.Error != nil {
		return fmt.Errorf("clear subscription: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription not found: %s", stripeSubscriptionID)
	}
	return nil
}

