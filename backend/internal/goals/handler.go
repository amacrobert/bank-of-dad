package goals

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/store"
)

const (
	MaxNameLength  = 50
	MaxTargetCents = 99999999
	MaxActiveGoals = 5
)

// Handler handles savings goal HTTP requests.
type Handler struct {
	goalStore  *store.SavingsGoalStore
	childStore *store.ChildStore
}

// NewHandler creates a new goals handler.
func NewHandler(goalStore *store.SavingsGoalStore, childStore *store.ChildStore) *Handler {
	return &Handler{
		goalStore:  goalStore,
		childStore: childStore,
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

// parseChildID extracts and validates the child ID from the URL path.
func parseChildID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// parseGoalID extracts and validates the goal ID from the URL path.
func parseGoalID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("goalId"), 10, 64)
}

// verifyChildAccess checks that the authenticated user can access the given child's data.
// Children can only access their own data. Parents can access any child in their family.
func (h *Handler) verifyChildAccess(r *http.Request, childID int64) (int, *ErrorResponse) {
	child, err := h.childStore.GetByID(childID)
	if err != nil {
		return http.StatusInternalServerError, &ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to lookup child.",
		}
	}
	if child == nil {
		return http.StatusNotFound, &ErrorResponse{
			Error:   "not_found",
			Message: "Child not found.",
		}
	}

	familyID := middleware.GetFamilyID(r)
	if child.FamilyID != familyID {
		return http.StatusForbidden, &ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have permission to access this child's goals.",
		}
	}

	userType := middleware.GetUserType(r)
	userID := middleware.GetUserID(r)
	if userType == "child" && userID != childID {
		return http.StatusForbidden, &ErrorResponse{
			Error:   "forbidden",
			Message: "You can only access your own goals.",
		}
	}

	return 0, nil
}

// verifyChildOwner checks that the authenticated user is a child accessing their own data.
// Parents are not allowed to create/update/delete/allocate goals.
func (h *Handler) verifyChildOwner(r *http.Request, childID int64) (int, *ErrorResponse) {
	userType := middleware.GetUserType(r)
	if userType != "child" {
		return http.StatusForbidden, &ErrorResponse{
			Error:   "forbidden",
			Message: "Only children can manage their own goals.",
		}
	}

	return h.verifyChildAccess(r, childID)
}

// CreateRequest represents a create savings goal request body.
type CreateRequest struct {
	Name        string  `json:"name"`
	TargetCents int64   `json:"target_cents"`
	Emoji       *string `json:"emoji,omitempty"`
	TargetDate  *string `json:"target_date,omitempty"`
}

// SavingsGoalsResponse represents the response for listing goals.
type SavingsGoalsResponse struct {
	Goals                 []*store.SavingsGoal `json:"goals"`
	AvailableBalanceCents int64                `json:"available_balance_cents"`
	TotalSavedCents       int64                `json:"total_saved_cents"`
}

// HandleCreate handles POST /api/children/{id}/savings-goals
func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	childID, err := parseChildID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_child_id", Message: "Invalid child ID."})
		return
	}

	if status, errResp := h.verifyChildOwner(r, childID); errResp != nil {
		writeJSON(w, status, errResp)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "Invalid request body."})
		return
	}

	// Validate name
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_name", Message: "Name is required."})
		return
	}
	if len(name) > MaxNameLength {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_name", Message: "Name must be 50 characters or less."})
		return
	}

	// Validate target_cents
	if req.TargetCents <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_target", Message: "Target amount must be greater than zero."})
		return
	}
	if req.TargetCents > MaxTargetCents {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_target", Message: "Target amount must be $999,999.99 or less."})
		return
	}

	// Check max active goals
	count, err := h.goalStore.CountActiveByChild(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to check active goals."})
		return
	}
	if count >= MaxActiveGoals {
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: "max_goals_reached", Message: "Maximum of 5 active goals reached."})
		return
	}

	// Parse optional target date
	var targetDate *time.Time
	if req.TargetDate != nil && *req.TargetDate != "" {
		t, err := time.Parse("2006-01-02", *req.TargetDate)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_target_date", Message: "Target date must be in YYYY-MM-DD format."})
			return
		}
		targetDate = &t
	}

	goal, err := h.goalStore.Create(childID, name, req.TargetCents, req.Emoji, targetDate)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to create goal."})
		return
	}

	writeJSON(w, http.StatusCreated, goal)
}

// HandleList handles GET /api/children/{id}/savings-goals
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	childID, err := parseChildID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_child_id", Message: "Invalid child ID."})
		return
	}

	if status, errResp := h.verifyChildAccess(r, childID); errResp != nil {
		writeJSON(w, status, errResp)
		return
	}

	goals, err := h.goalStore.ListByChild(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to list goals."})
		return
	}

	if goals == nil {
		goals = []*store.SavingsGoal{}
	}

	// Compute totals from active goals
	var totalSavedCents int64
	for _, g := range goals {
		if g.Status == "active" {
			totalSavedCents += g.SavedCents
		}
	}

	// Get child's total balance to compute available balance
	child, err := h.childStore.GetByID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get balance."})
		return
	}

	availableBalanceCents := child.BalanceCents - totalSavedCents

	writeJSON(w, http.StatusOK, SavingsGoalsResponse{
		Goals:                 goals,
		AvailableBalanceCents: availableBalanceCents,
		TotalSavedCents:       totalSavedCents,
	})
}

// HandleUpdate handles PUT /api/children/{id}/savings-goals/{goalId}
func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, ErrorResponse{Error: "not_implemented"})
}

// HandleDelete handles DELETE /api/children/{id}/savings-goals/{goalId}
func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, ErrorResponse{Error: "not_implemented"})
}

// HandleAllocate handles POST /api/children/{id}/savings-goals/{goalId}/allocate
func (h *Handler) HandleAllocate(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, ErrorResponse{Error: "not_implemented"})
}

// HandleListAllocations handles GET /api/children/{id}/savings-goals/{goalId}/allocations
func (h *Handler) HandleListAllocations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, ErrorResponse{Error: "not_implemented"})
}
