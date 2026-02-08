package interest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/store"
)

// Handler handles interest-related HTTP requests.
type Handler struct {
	interestStore *store.InterestStore
	childStore    *store.ChildStore
}

// NewHandler creates a new interest handler.
func NewHandler(interestStore *store.InterestStore, childStore *store.ChildStore) *Handler {
	return &Handler{
		interestStore: interestStore,
		childStore:    childStore,
	}
}

// InterestRateRequest represents the request body for setting an interest rate.
type InterestRateRequest struct {
	InterestRateBps int `json:"interest_rate_bps"`
}

// InterestRateResponse represents the response after setting an interest rate.
type InterestRateResponse struct {
	ChildID             int64  `json:"child_id"`
	InterestRateBps     int    `json:"interest_rate_bps"`
	InterestRateDisplay string `json:"interest_rate_display"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// FormatRateDisplay converts basis points to a display string like "5.00%".
func FormatRateDisplay(bps int) string {
	return fmt.Sprintf("%.2f%%", float64(bps)/100.0)
}

// HandleSetInterestRate handles PUT /api/children/{id}/interest-rate
func (h *Handler) HandleSetInterestRate(w http.ResponseWriter, r *http.Request) {
	// Check user type - only parents can set interest rates
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
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

	// Get child and verify existence
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

	// Verify parent owns this child's family
	familyID := middleware.GetFamilyID(r)
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
		return
	}

	// Parse request body
	var req InterestRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body.",
		})
		return
	}

	// Validate rate
	if req.InterestRateBps < 0 || req.InterestRateBps > 10000 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid interest rate",
			Message: "Interest rate must be between 0% and 100%",
		})
		return
	}

	// Set rate
	if err := h.interestStore.SetInterestRate(childID, req.InterestRateBps); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to set interest rate.",
		})
		return
	}

	writeJSON(w, http.StatusOK, InterestRateResponse{
		ChildID:             childID,
		InterestRateBps:     req.InterestRateBps,
		InterestRateDisplay: FormatRateDisplay(req.InterestRateBps),
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck // best-effort response encoding
}
