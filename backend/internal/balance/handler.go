package balance

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/models"
	"bank-of-dad/repositories"
)

const (
	MaxAmountCents = 99999999 // $999,999.99
	MaxNoteLength  = 500
)

// Handler handles balance-related HTTP requests.
type Handler struct {
	txRepo               *repositories.TransactionRepo
	childRepo            *repositories.ChildRepo
	interestRepo         *repositories.InterestRepo
	interestScheduleRepo *repositories.InterestScheduleRepo
	goalRepo             *repositories.SavingsGoalRepo
}

// NewHandler creates a new balance handler.
func NewHandler(txRepo *repositories.TransactionRepo, childRepo *repositories.ChildRepo, interestRepo *repositories.InterestRepo, interestScheduleRepo *repositories.InterestScheduleRepo, goalRepo *repositories.SavingsGoalRepo) *Handler {
	return &Handler{
		txRepo:               txRepo,
		childRepo:            childRepo,
		interestRepo:         interestRepo,
		interestScheduleRepo: interestScheduleRepo,
		goalRepo:             goalRepo,
	}
}

// DepositRequest represents a deposit request body.
type DepositRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Note        string `json:"note,omitempty"`
}

// WithdrawRequest represents a withdrawal request body.
type WithdrawRequest struct {
	AmountCents      int64  `json:"amount_cents"`
	Note             string `json:"note,omitempty"`
	ConfirmGoalImpact bool  `json:"confirm_goal_impact,omitempty"`
}

// TransactionResponse represents the response after a successful transaction.
type TransactionResponse struct {
	Transaction     *models.Transaction `json:"transaction"`
	NewBalanceCents int64               `json:"new_balance_cents"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HandleDeposit handles POST /api/children/{id}/deposit
func (h *Handler) HandleDeposit(w http.ResponseWriter, r *http.Request) {
	// Check user type - only parents can deposit
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can deposit money.",
		})
		return
	}

	// Parse child ID
	childIDStr := r.PathValue("id")
	childID, err := strconv.ParseInt(childIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_child_id",
			Message: "Invalid child ID.",
		})
		return
	}

	// Get child and verify authorization
	child, err := h.childRepo.GetByID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to lookup child.",
		})
		return
	}
	if child == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Child not found.",
		})
		return
	}

	// Verify parent has access to this child's family
	familyID := middleware.GetFamilyID(r)
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have permission to access this child's account.",
		})
		return
	}

	// Check if child account is disabled (free tier limit)
	if child.IsDisabled {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "Account disabled",
			Message: "This account is disabled. Upgrade to Plus to enable all children.",
		})
		return
	}

	// Parse request body
	var req DepositRequest
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

	// Validate note
	note := strings.TrimSpace(req.Note)
	if len(note) > MaxNoteLength {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_note",
			Message: "Note must be 500 characters or less.",
		})
		return
	}

	// Perform deposit
	parentID := middleware.GetUserID(r)
	tx, newBalance, err := h.txRepo.Deposit(childID, parentID, req.AmountCents, note)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to process deposit.",
		})
		return
	}

	writeJSON(w, http.StatusOK, TransactionResponse{
		Transaction:     tx,
		NewBalanceCents: newBalance,
	})
}

// HandleWithdraw handles POST /api/children/{id}/withdraw
func (h *Handler) HandleWithdraw(w http.ResponseWriter, r *http.Request) {
	// Check user type - only parents can withdraw
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can withdraw money.",
		})
		return
	}

	// Parse child ID
	childIDStr := r.PathValue("id")
	childID, err := strconv.ParseInt(childIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_child_id",
			Message: "Invalid child ID.",
		})
		return
	}

	// Get child and verify authorization
	child, err := h.childRepo.GetByID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to lookup child.",
		})
		return
	}
	if child == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Child not found.",
		})
		return
	}

	// Verify parent has access to this child's family
	familyID := middleware.GetFamilyID(r)
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have permission to access this child's account.",
		})
		return
	}

	// Check if child account is disabled (free tier limit)
	if child.IsDisabled {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "Account disabled",
			Message: "This account is disabled. Upgrade to Plus to enable all children.",
		})
		return
	}

	// Parse request body
	var req WithdrawRequest
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

	// Validate note
	note := strings.TrimSpace(req.Note)
	if len(note) > MaxNoteLength {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_note",
			Message: "Note must be 500 characters or less.",
		})
		return
	}

	// Check for goal impact before withdrawal
	parentID := middleware.GetUserID(r)

	if h.goalRepo != nil {
		totalSaved, err := h.goalRepo.GetTotalSavedByChild(childID)
		if err == nil && totalSaved > 0 {
			// Balance after withdrawal
			newBalanceAfter := child.BalanceCents - req.AmountCents
			if newBalanceAfter < totalSaved && !req.ConfirmGoalImpact {
				// Goals would be impacted — return warning
				totalToRelease := totalSaved - newBalanceAfter
				affectedGoals, err := h.goalRepo.GetAffectedGoals(childID, totalToRelease)
				if err == nil && len(affectedGoals) > 0 {
					writeJSON(w, http.StatusConflict, map[string]interface{}{
						"error":              "goal_impact_warning",
						"message":            "This withdrawal will reduce savings goals allocations.",
						"affected_goals":     affectedGoals,
						"total_released_cents": totalToRelease,
					})
					return
				}
			}
		}
	}

	// Perform withdrawal
	tx, newBalance, err := h.txRepo.Withdraw(childID, parentID, req.AmountCents, note)
	if err != nil {
		if err == models.ErrInsufficientFunds {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "insufficient_funds",
				Message: formatInsufficientFundsMessage(req.AmountCents, newBalance),
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to process withdrawal.",
		})
		return
	}

	// If goals were impacted and confirmed, reduce them proportionally
	if h.goalRepo != nil && req.ConfirmGoalImpact {
		totalSaved, err := h.goalRepo.GetTotalSavedByChild(childID)
		if err == nil && totalSaved > newBalance {
			totalToRelease := totalSaved - newBalance
			if err := h.goalRepo.ReduceGoalsProportionally(childID, totalToRelease); err != nil {
				log.Printf("WARN: failed to reduce goals proportionally for child %d: %v", childID, err)
			}
		}
	}

	writeJSON(w, http.StatusOK, TransactionResponse{
		Transaction:     tx,
		NewBalanceCents: newBalance,
	})
}

func formatInsufficientFundsMessage(requested, available int64) string {
	requestedDollars := float64(requested) / 100
	availableDollars := float64(available) / 100
	return "Cannot withdraw $" + formatMoney(requestedDollars) + ". Current balance is $" + formatMoney(availableDollars) + "."
}

func formatMoney(amount float64) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}

// BalanceResponse represents a balance query response.
type BalanceResponse struct {
	ChildID               int64   `json:"child_id"`
	FirstName             string  `json:"first_name"`
	BalanceCents          int64   `json:"balance_cents"`
	InterestRateBps       int     `json:"interest_rate_bps"`
	InterestRateDisplay   string  `json:"interest_rate_display"`
	NextInterestAt        *string `json:"next_interest_at,omitempty"`
	AvailableBalanceCents *int64  `json:"available_balance_cents,omitempty"`
	TotalSavedCents       *int64  `json:"total_saved_cents,omitempty"`
	ActiveGoalsCount      *int    `json:"active_goals_count,omitempty"`
}

// TransactionListResponse represents a list of transactions.
type TransactionListResponse struct {
	Transactions []models.Transaction `json:"transactions"`
}

// HandleGetBalance handles GET /api/children/{id}/balance
func (h *Handler) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	// Parse child ID
	childIDStr := r.PathValue("id")
	childID, err := strconv.ParseInt(childIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_child_id",
			Message: "Invalid child ID.",
		})
		return
	}

	// Get child
	child, err := h.childRepo.GetByID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to lookup child.",
		})
		return
	}
	if child == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Child not found.",
		})
		return
	}

	// Check authorization
	userType := middleware.GetUserType(r)
	userID := middleware.GetUserID(r)
	familyID := middleware.GetFamilyID(r)

	// Child must be in the same family
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have permission to view this balance.",
		})
		return
	}

	// If user is a child, they can only view their own balance
	if userType == "child" && userID != childID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You can only view your own balance.",
		})
		return
	}

	// Get interest rate
	rateBps := 0
	if h.interestRepo != nil {
		rateBps, _ = h.interestRepo.GetInterestRate(childID)
	}

	// Get next interest payment date
	var nextInterestAt *string
	if h.interestScheduleRepo != nil {
		sched, err := h.interestScheduleRepo.GetByChildID(childID)
		if err == nil && sched != nil && sched.NextRunAt != nil && sched.Status == models.ScheduleStatusActive {
			s := sched.NextRunAt.Format(time.RFC3339)
			nextInterestAt = &s
		}
	}

	resp := BalanceResponse{
		ChildID:             child.ID,
		FirstName:           child.FirstName,
		BalanceCents:        child.BalanceCents,
		InterestRateBps:     rateBps,
		InterestRateDisplay: fmt.Sprintf("%.2f%%", float64(rateBps)/100.0),
		NextInterestAt:      nextInterestAt,
	}

	// Include savings goal information if goalRepo is available
	if h.goalRepo != nil {
		availableBalance, err := h.goalRepo.GetAvailableBalance(childID)
		if err == nil {
			totalSaved := child.BalanceCents - availableBalance
			resp.AvailableBalanceCents = &availableBalance
			resp.TotalSavedCents = &totalSaved
		}
		activeCount, err := h.goalRepo.CountActiveByChild(childID)
		if err == nil {
			resp.ActiveGoalsCount = &activeCount
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleGetTransactions handles GET /api/children/{id}/transactions
func (h *Handler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	// Parse child ID
	childIDStr := r.PathValue("id")
	childID, err := strconv.ParseInt(childIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_child_id",
			Message: "Invalid child ID.",
		})
		return
	}

	// Get child to verify existence and family
	child, err := h.childRepo.GetByID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to lookup child.",
		})
		return
	}
	if child == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Child not found.",
		})
		return
	}

	// Check authorization
	userType := middleware.GetUserType(r)
	userID := middleware.GetUserID(r)
	familyID := middleware.GetFamilyID(r)

	// Child must be in the same family
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have permission to view these transactions.",
		})
		return
	}

	// If user is a child, they can only view their own transactions
	if userType == "child" && userID != childID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You can only view your own transactions.",
		})
		return
	}

	// Get transactions
	transactions, err := h.txRepo.ListByChild(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve transactions.",
		})
		return
	}

	// Return empty array instead of null for no transactions
	if transactions == nil {
		transactions = []models.Transaction{}
	}

	writeJSON(w, http.StatusOK, TransactionListResponse{
		Transactions: transactions,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
