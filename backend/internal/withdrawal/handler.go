package withdrawal

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/models"
	"bank-of-dad/repositories"
)

const (
	MaxAmountCents = 99999999 // $999,999.99
	MaxReasonLength = 500
)

// Handler handles withdrawal request HTTP endpoints.
type Handler struct {
	wrRepo    *repositories.WithdrawalRequestRepo
	txRepo    *repositories.TransactionRepo
	childRepo *repositories.ChildRepo
	goalRepo  *repositories.SavingsGoalRepo
}

// NewHandler creates a new withdrawal request handler.
func NewHandler(wrRepo *repositories.WithdrawalRequestRepo, txRepo *repositories.TransactionRepo, childRepo *repositories.ChildRepo, goalRepo *repositories.SavingsGoalRepo) *Handler {
	return &Handler{
		wrRepo:    wrRepo,
		txRepo:    txRepo,
		childRepo: childRepo,
		goalRepo:  goalRepo,
	}
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// SubmitRequest represents the request body for submitting a withdrawal request.
type SubmitRequest struct {
	AmountCents int    `json:"amount_cents"`
	Reason      string `json:"reason"`
}

// HandleSubmitRequest handles POST /api/child/withdrawal-requests
func (h *Handler) HandleSubmitRequest(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "child" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only children can submit withdrawal requests.",
		})
		return
	}

	childID := middleware.GetUserID(r)
	familyID := middleware.GetFamilyID(r)

	// Check if child account is disabled
	child, err := h.childRepo.GetByID(childID)
	if err != nil || child == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to lookup account.",
		})
		return
	}
	if child.IsDisabled {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "account_disabled",
			Message: "Your account is disabled.",
		})
		return
	}

	// Parse request body
	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body.",
		})
		return
	}

	// Validate amount
	if req.AmountCents <= 0 || req.AmountCents > MaxAmountCents {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_amount",
			Message: "Amount must be between 1 cent and $999,999.99.",
		})
		return
	}

	// Validate reason
	reason := strings.TrimSpace(req.Reason)
	if reason == "" || len(reason) > MaxReasonLength {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_reason",
			Message: "Reason is required and must be 500 characters or less.",
		})
		return
	}

	// Check available balance
	availableBalance := child.BalanceCents
	if h.goalRepo != nil {
		totalSaved, err := h.goalRepo.GetTotalSavedByChild(childID)
		if err == nil && totalSaved > 0 {
			availableBalance = child.BalanceCents - totalSaved
		}
	}
	if int64(req.AmountCents) > availableBalance {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Error:   "insufficient_funds",
			Message: "Requested amount exceeds your available balance.",
		})
		return
	}

	// Create the withdrawal request
	wr := &models.WithdrawalRequest{
		ChildID:     childID,
		FamilyID:    familyID,
		AmountCents: req.AmountCents,
		Reason:      reason,
	}

	created, err := h.wrRepo.Create(wr)
	if err != nil {
		if err == repositories.ErrPendingRequestExists {
			writeJSON(w, http.StatusConflict, ErrorResponse{
				Error:   "pending_request_exists",
				Message: "You already have a pending withdrawal request. Cancel it or wait for a response before submitting a new one.",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create withdrawal request.",
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"withdrawal_request": created,
	})
}

// HandleCancelRequest handles POST /api/child/withdrawal-requests/{id}/cancel
func (h *Handler) HandleCancelRequest(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "child" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only children can cancel their own requests.",
		})
		return
	}

	childID := middleware.GetUserID(r)

	reqID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid request ID.",
		})
		return
	}

	if err := h.wrRepo.Cancel(reqID, childID); err != nil {
		if err == repositories.ErrInvalidStatusTransition {
			writeJSON(w, http.StatusConflict, ErrorResponse{
				Error:   "invalid_status",
				Message: "Request is not pending or does not belong to you.",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to cancel request.",
		})
		return
	}

	// Fetch updated request
	updated, _ := h.wrRepo.GetByID(reqID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"withdrawal_request": updated,
	})
}

// HandleChildListRequests handles GET /api/child/withdrawal-requests
func (h *Handler) HandleChildListRequests(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "child" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only children can view their own requests.",
		})
		return
	}

	childID := middleware.GetUserID(r)
	status := r.URL.Query().Get("status")

	requests, err := h.wrRepo.ListByChild(childID, status)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list requests.",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"withdrawal_requests": requests,
	})
}

// ApproveRequest represents the request body for approving a withdrawal request.
type ApproveRequest struct {
	ConfirmGoalImpact bool `json:"confirm_goal_impact"`
}

// HandleApprove handles POST /api/withdrawal-requests/{id}/approve
func (h *Handler) HandleApprove(w http.ResponseWriter, r *http.Request) {
	parentID := middleware.GetUserID(r)
	familyID := middleware.GetFamilyID(r)

	reqID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid request ID.",
		})
		return
	}

	// Get the withdrawal request
	wr, err := h.wrRepo.GetByID(reqID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to lookup request.",
		})
		return
	}
	if wr == nil || wr.FamilyID != familyID {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Withdrawal request not found.",
		})
		return
	}

	if wr.Status != models.WithdrawalRequestStatusPending {
		writeJSON(w, http.StatusConflict, ErrorResponse{
			Error:   "invalid_status",
			Message: "Request is not pending.",
		})
		return
	}

	// Check child account
	child, err := h.childRepo.GetByID(wr.ChildID)
	if err != nil || child == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to lookup child.",
		})
		return
	}
	if child.IsDisabled {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Error:   "account_disabled",
			Message: "Child's account is disabled. Deny this request instead.",
		})
		return
	}

	// Check available balance
	if child.BalanceCents < int64(wr.AmountCents) {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Error:   "insufficient_funds",
			Message: "Child no longer has sufficient funds for this request.",
		})
		return
	}

	// Parse request body for confirm_goal_impact
	var approveReq ApproveRequest
	_ = json.NewDecoder(r.Body).Decode(&approveReq)

	// Check for goal impact
	if h.goalRepo != nil {
		totalSaved, err := h.goalRepo.GetTotalSavedByChild(wr.ChildID)
		if err == nil && totalSaved > 0 {
			newBalanceAfter := child.BalanceCents - int64(wr.AmountCents)
			if newBalanceAfter < totalSaved && !approveReq.ConfirmGoalImpact {
				totalToRelease := totalSaved - newBalanceAfter
				affectedGoals, err := h.goalRepo.GetAffectedGoals(wr.ChildID, totalToRelease)
				if err == nil && len(affectedGoals) > 0 {
					writeJSON(w, http.StatusConflict, map[string]interface{}{
						"error":                "goal_impact_warning",
						"message":              "This approval will reduce savings goals allocations.",
						"affected_goals":       affectedGoals,
						"total_released_cents": totalToRelease,
					})
					return
				}
			}
		}
	}

	// Create withdrawal transaction
	note := "Withdrawal request: " + wr.Reason
	tx, newBalance, err := h.txRepo.WithdrawAsType(wr.ChildID, parentID, int64(wr.AmountCents), note, models.TransactionTypeWithdrawalRequest)
	if err != nil {
		if err == models.ErrInsufficientFunds {
			writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
				Error:   "insufficient_funds",
				Message: "Child no longer has sufficient funds for this request.",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to process withdrawal.",
		})
		return
	}

	// Reduce goals proportionally if confirmed
	if h.goalRepo != nil && approveReq.ConfirmGoalImpact {
		totalSaved, err := h.goalRepo.GetTotalSavedByChild(wr.ChildID)
		if err == nil && totalSaved > newBalance {
			totalToRelease := totalSaved - newBalance
			if err := h.goalRepo.ReduceGoalsProportionally(wr.ChildID, totalToRelease); err != nil {
				log.Printf("WARN: failed to reduce goals proportionally for child %d: %v", wr.ChildID, err)
			}
		}
	}

	// Update request status
	if err := h.wrRepo.Approve(reqID, parentID, tx.ID); err != nil {
		log.Printf("ERROR: withdrawal succeeded but request status update failed for request %d: %v", reqID, err)
	}

	// Fetch updated request
	updated, _ := h.wrRepo.GetByID(reqID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"withdrawal_request": updated,
		"new_balance_cents":  newBalance,
	})
}

// DenyRequest represents the request body for denying a withdrawal request.
type DenyRequest struct {
	Reason string `json:"reason"`
}

// HandleDeny handles POST /api/withdrawal-requests/{id}/deny
func (h *Handler) HandleDeny(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	parentID := middleware.GetUserID(r)

	reqID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid request ID.",
		})
		return
	}

	// Get the withdrawal request
	wr, err := h.wrRepo.GetByID(reqID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to lookup request.",
		})
		return
	}
	if wr == nil || wr.FamilyID != familyID {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Withdrawal request not found.",
		})
		return
	}

	// Parse optional denial reason
	var denyReq DenyRequest
	_ = json.NewDecoder(r.Body).Decode(&denyReq)

	reason := strings.TrimSpace(denyReq.Reason)
	if len(reason) > MaxReasonLength {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_reason",
			Message: "Denial reason must be 500 characters or less.",
		})
		return
	}

	if err := h.wrRepo.Deny(reqID, parentID, reason); err != nil {
		if err == repositories.ErrInvalidStatusTransition {
			writeJSON(w, http.StatusConflict, ErrorResponse{
				Error:   "invalid_status",
				Message: "Request is not pending.",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to deny request.",
		})
		return
	}

	updated, _ := h.wrRepo.GetByID(reqID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"withdrawal_request": updated,
	})
}

// HandlePendingCount handles GET /api/withdrawal-requests/pending/count
func (h *Handler) HandlePendingCount(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)

	count, err := h.wrRepo.PendingCountByFamily(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to count pending requests.",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count": count,
	})
}

// HandleParentListRequests handles GET /api/withdrawal-requests
func (h *Handler) HandleParentListRequests(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	status := r.URL.Query().Get("status")

	var childID int64
	if cidStr := r.URL.Query().Get("child_id"); cidStr != "" {
		var err error
		childID, err = strconv.ParseInt(cidStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_child_id",
				Message: "Invalid child_id parameter.",
			})
			return
		}
	}

	requests, err := h.wrRepo.ListByFamily(familyID, status, childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list requests.",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"withdrawal_requests": requests,
	})
}
