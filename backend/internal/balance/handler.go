package balance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/store"
)

const (
	MaxAmountCents = 99999999 // $999,999.99
	MaxNoteLength  = 500
)

// Handler handles balance-related HTTP requests.
type Handler struct {
	txStore    *store.TransactionStore
	childStore *store.ChildStore
}

// NewHandler creates a new balance handler.
func NewHandler(txStore *store.TransactionStore, childStore *store.ChildStore) *Handler {
	return &Handler{
		txStore:    txStore,
		childStore: childStore,
	}
}

// DepositRequest represents a deposit request body.
type DepositRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Note        string `json:"note,omitempty"`
}

// WithdrawRequest represents a withdrawal request body.
type WithdrawRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Note        string `json:"note,omitempty"`
}

// TransactionResponse represents the response after a successful transaction.
type TransactionResponse struct {
	Transaction     *store.Transaction `json:"transaction"`
	NewBalanceCents int64              `json:"new_balance_cents"`
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
	child, err := h.childStore.GetByID(childID)
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
	tx, newBalance, err := h.txStore.Deposit(childID, parentID, req.AmountCents, note)
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
	child, err := h.childStore.GetByID(childID)
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

	// Perform withdrawal
	parentID := middleware.GetUserID(r)
	tx, newBalance, err := h.txStore.Withdraw(childID, parentID, req.AmountCents, note)
	if err != nil {
		if err == store.ErrInsufficientFunds {
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
	ChildID      int64 `json:"child_id"`
	BalanceCents int64 `json:"balance_cents"`
}

// TransactionListResponse represents a list of transactions.
type TransactionListResponse struct {
	Transactions []store.Transaction `json:"transactions"`
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
	child, err := h.childStore.GetByID(childID)
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

	writeJSON(w, http.StatusOK, BalanceResponse{
		ChildID:      child.ID,
		BalanceCents: child.BalanceCents,
	})
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
	child, err := h.childStore.GetByID(childID)
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
	transactions, err := h.txStore.ListByChild(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve transactions.",
		})
		return
	}

	// Return empty array instead of null for no transactions
	if transactions == nil {
		transactions = []store.Transaction{}
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
