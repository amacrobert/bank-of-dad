package repositories

import (
	"errors"
	"fmt"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

var (
	ErrWithdrawalRequestNotFound = errors.New("withdrawal request not found")
	ErrPendingRequestExists      = errors.New("child already has a pending withdrawal request")
)

// WithdrawalRequestWithChild extends WithdrawalRequest with the child's first name.
type WithdrawalRequestWithChild struct {
	models.WithdrawalRequest
	ChildName string `json:"child_name" gorm:"column:child_name"`
}

// WithdrawalRequestRepo handles database operations for withdrawal requests using GORM.
type WithdrawalRequestRepo struct {
	db *gorm.DB
}

// NewWithdrawalRequestRepo creates a new WithdrawalRequestRepo.
func NewWithdrawalRequestRepo(db *gorm.DB) *WithdrawalRequestRepo {
	return &WithdrawalRequestRepo{db: db}
}

// GetByID retrieves a withdrawal request by its ID. Returns (nil, nil) if not found.
func (r *WithdrawalRequestRepo) GetByID(id int64) (*models.WithdrawalRequest, error) {
	var req models.WithdrawalRequest
	err := r.db.First(&req, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get withdrawal request by id: %w", err)
	}
	return &req, nil
}

// PendingCountByFamily returns the number of pending withdrawal requests for a family.
func (r *WithdrawalRequestRepo) PendingCountByFamily(familyID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.WithdrawalRequest{}).
		Where("family_id = ? AND status = ?", familyID, models.WithdrawalRequestStatusPending).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count pending withdrawal requests: %w", err)
	}
	return count, nil
}

// Create inserts a new withdrawal request with pending status.
// Returns ErrPendingRequestExists if the child already has a pending request.
func (r *WithdrawalRequestRepo) Create(req *models.WithdrawalRequest) (*models.WithdrawalRequest, error) {
	req.Status = models.WithdrawalRequestStatusPending
	if err := r.db.Create(req).Error; err != nil {
		if isDuplicateKey(err) {
			return nil, ErrPendingRequestExists
		}
		return nil, fmt.Errorf("create withdrawal request: %w", err)
	}
	return req, nil
}

// Approve transitions a withdrawal request from pending to approved.
// Sets reviewed_by_parent_id, reviewed_at, and transaction_id.
func (r *WithdrawalRequestRepo) Approve(id int64, parentID int64, transactionID int64) error {
	result := r.db.Model(&models.WithdrawalRequest{}).
		Where("id = ? AND status = ?", id, models.WithdrawalRequestStatusPending).
		Updates(map[string]interface{}{
			"status":               models.WithdrawalRequestStatusApproved,
			"reviewed_at":          gorm.Expr("NOW()"),
			"reviewed_by_parent_id": parentID,
			"transaction_id":       transactionID,
			"updated_at":           gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return fmt.Errorf("approve withdrawal request: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrInvalidStatusTransition
	}
	return nil
}

// Deny transitions a withdrawal request from pending to denied.
// Sets reviewed_by_parent_id, reviewed_at, and optional denial_reason.
func (r *WithdrawalRequestRepo) Deny(id int64, parentID int64, reason string) error {
	updates := map[string]interface{}{
		"status":               models.WithdrawalRequestStatusDenied,
		"reviewed_at":          gorm.Expr("NOW()"),
		"reviewed_by_parent_id": parentID,
		"updated_at":           gorm.Expr("NOW()"),
	}
	if reason != "" {
		updates["denial_reason"] = reason
	}

	result := r.db.Model(&models.WithdrawalRequest{}).
		Where("id = ? AND status = ?", id, models.WithdrawalRequestStatusPending).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("deny withdrawal request: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrInvalidStatusTransition
	}
	return nil
}

// Cancel transitions a withdrawal request from pending to cancelled.
// Verifies the request belongs to the specified child.
func (r *WithdrawalRequestRepo) Cancel(id int64, childID int64) error {
	result := r.db.Model(&models.WithdrawalRequest{}).
		Where("id = ? AND child_id = ? AND status = ?", id, childID, models.WithdrawalRequestStatusPending).
		Updates(map[string]interface{}{
			"status":     models.WithdrawalRequestStatusCancelled,
			"updated_at": gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return fmt.Errorf("cancel withdrawal request: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrInvalidStatusTransition
	}
	return nil
}

// ListByChild returns withdrawal requests for a child, with optional status filter.
// Ordered by created_at desc.
func (r *WithdrawalRequestRepo) ListByChild(childID int64, status string) ([]models.WithdrawalRequest, error) {
	query := r.db.Where("child_id = ?", childID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var results []models.WithdrawalRequest
	err := query.Order("created_at DESC").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("list withdrawal requests by child: %w", err)
	}
	return results, nil
}

// ListByFamily returns withdrawal requests for a family with child name, optional status and child_id filters.
// Ordered by created_at desc.
func (r *WithdrawalRequestRepo) ListByFamily(familyID int64, status string, childID int64) ([]WithdrawalRequestWithChild, error) {
	query := r.db.Table("withdrawal_requests").
		Select("withdrawal_requests.*, children.first_name as child_name").
		Joins("JOIN children ON children.id = withdrawal_requests.child_id").
		Where("withdrawal_requests.family_id = ?", familyID)

	if status != "" {
		query = query.Where("withdrawal_requests.status = ?", status)
	}
	if childID > 0 {
		query = query.Where("withdrawal_requests.child_id = ?", childID)
	}

	var results []WithdrawalRequestWithChild
	err := query.Order("withdrawal_requests.created_at DESC").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("list withdrawal requests by family: %w", err)
	}
	return results, nil
}
