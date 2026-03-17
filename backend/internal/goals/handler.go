package goals

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/models"
	"bank-of-dad/repositories"
)

const (
	MaxNameLength  = 50
	MaxTargetCents = 99999999
	MaxActiveGoals = 5
)

// Handler handles savings goal HTTP requests.
type Handler struct {
	goalRepo           *repositories.SavingsGoalRepo
	childRepo          *repositories.ChildRepo
	goalAllocationRepo *repositories.GoalAllocationRepo
}

// NewHandler creates a new goals handler.
func NewHandler(goalRepo *repositories.SavingsGoalRepo, childRepo *repositories.ChildRepo, goalAllocationRepo *repositories.GoalAllocationRepo) *Handler {
	return &Handler{
		goalRepo:           goalRepo,
		childRepo:          childRepo,
		goalAllocationRepo: goalAllocationRepo,
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
	child, err := h.childRepo.GetByID(childID)
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
}

// SavingsGoalsResponse represents the response for listing goals.
type SavingsGoalsResponse struct {
	Goals                 []*models.SavingsGoal `json:"goals"`
	AvailableBalanceCents int64                 `json:"available_balance_cents"`
	TotalSavedCents       int64                 `json:"total_saved_cents"`
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
	count, err := h.goalRepo.CountActiveByChild(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to check active goals."})
		return
	}
	if count >= MaxActiveGoals {
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: "max_goals_reached", Message: "Maximum of 5 active goals reached."})
		return
	}

	goal, err := h.goalRepo.Create(childID, name, req.TargetCents, req.Emoji)
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

	goals, err := h.goalRepo.ListByChild(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to list goals."})
		return
	}

	if goals == nil {
		goals = []*models.SavingsGoal{}
	}

	// Compute totals from active goals
	var totalSavedCents int64
	for _, g := range goals {
		if g.Status == "active" {
			totalSavedCents += g.SavedCents
		}
	}

	// Get child's total balance to compute available balance
	child, err := h.childRepo.GetByID(childID)
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

// UpdateRequest represents a partial update request for a savings goal.
type UpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	TargetCents *int64  `json:"target_cents,omitempty"`
	Emoji       *string `json:"emoji"`
}

// HandleUpdate handles PUT /api/children/{id}/savings-goals/{goalId}
func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	childID, err := parseChildID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_child_id", Message: "Invalid child ID."})
		return
	}

	goalID, err := parseGoalID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_goal_id", Message: "Invalid goal ID."})
		return
	}

	if status, errResp := h.verifyChildOwner(r, childID); errResp != nil {
		writeJSON(w, status, errResp)
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "Invalid request body."})
		return
	}

	// Validate fields if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_name", Message: "Name cannot be empty."})
			return
		}
		if len(name) > MaxNameLength {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_name", Message: "Name must be 50 characters or less."})
			return
		}
		req.Name = &name
	}

	if req.TargetCents != nil {
		if *req.TargetCents <= 0 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_target", Message: "Target amount must be greater than zero."})
			return
		}
		if *req.TargetCents > MaxTargetCents {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_target", Message: "Target amount must be $999,999.99 or less."})
			return
		}
	}

	// Build update params
	params := &repositories.UpdateGoalParams{
		Name:        req.Name,
		TargetCents: req.TargetCents,
	}

	// Handle emoji: the JSON decoder will set Emoji to non-nil even for explicit null,
	// but we use EmojiSet to indicate the field was present in the JSON.
	// Since UpdateRequest uses *string for emoji, if it was present in JSON it will be set.
	// We need to detect if the key was present. For simplicity, always set if non-nil.
	if req.Emoji != nil {
		params.Emoji = req.Emoji
		params.EmojiSet = true
	}

	goal, err := h.goalRepo.Update(goalID, childID, params)
	if err != nil {
		if err == repositories.ErrGoalNotFound {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Goal not found or not active."})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to update goal."})
		return
	}

	writeJSON(w, http.StatusOK, goal)
}

// DeleteGoalResponse represents the response after deleting a goal.
type DeleteGoalResponse struct {
	ReleasedCents         int64 `json:"released_cents"`
	AvailableBalanceCents int64 `json:"available_balance_cents"`
}

// HandleDelete handles DELETE /api/children/{id}/savings-goals/{goalId}
func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	childID, err := parseChildID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_child_id", Message: "Invalid child ID."})
		return
	}

	goalID, err := parseGoalID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_goal_id", Message: "Invalid goal ID."})
		return
	}

	if status, errResp := h.verifyChildOwner(r, childID); errResp != nil {
		writeJSON(w, status, errResp)
		return
	}

	releasedCents, err := h.goalRepo.Delete(goalID, childID)
	if err != nil {
		if err == repositories.ErrGoalNotFound {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Goal not found or not active."})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to delete goal."})
		return
	}

	// Get updated available balance
	availableBalance, err := h.goalRepo.GetAvailableBalance(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get available balance."})
		return
	}

	writeJSON(w, http.StatusOK, DeleteGoalResponse{
		ReleasedCents:         releasedCents,
		AvailableBalanceCents: availableBalance,
	})
}

// AllocateRequest represents an allocate/de-allocate request body.
type AllocateRequest struct {
	AmountCents int64 `json:"amount_cents"`
}

// AllocateResponse represents the response after allocating funds to a goal.
type AllocateResponse struct {
	Goal                  *models.SavingsGoal `json:"goal"`
	AvailableBalanceCents int64               `json:"available_balance_cents"`
	Completed             bool                `json:"completed"`
}

// AllocationsListResponse represents a list of goal allocations.
type AllocationsListResponse struct {
	Allocations []*models.GoalAllocation `json:"allocations"`
}

// HandleAllocate handles POST /api/children/{id}/savings-goals/{goalId}/allocate
func (h *Handler) HandleAllocate(w http.ResponseWriter, r *http.Request) {
	childID, err := parseChildID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_child_id", Message: "Invalid child ID."})
		return
	}

	goalID, err := parseGoalID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_goal_id", Message: "Invalid goal ID."})
		return
	}

	if status, errResp := h.verifyChildOwner(r, childID); errResp != nil {
		writeJSON(w, status, errResp)
		return
	}

	var req AllocateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "Invalid request body."})
		return
	}

	if req.AmountCents == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_amount", Message: "Amount must be non-zero."})
		return
	}

	goal, err := h.goalRepo.Allocate(goalID, childID, req.AmountCents)
	if err != nil {
		switch err {
		case repositories.ErrGoalNotFound:
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Goal not found or not active."})
		case repositories.ErrInsufficientAvailable:
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "insufficient_balance", Message: "Amount exceeds available balance."})
		case repositories.ErrDeallocationExceedsSaved:
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "exceeds_saved", Message: "De-allocation amount exceeds saved amount."})
		case repositories.ErrZeroAllocation:
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_amount", Message: "Amount must be non-zero."})
		default:
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to allocate funds."})
		}
		return
	}

	// Get updated available balance
	availableBalance, err := h.goalRepo.GetAvailableBalance(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get available balance."})
		return
	}

	writeJSON(w, http.StatusOK, AllocateResponse{
		Goal:                  goal,
		AvailableBalanceCents: availableBalance,
		Completed:             goal.Status == "completed",
	})
}

// HandleListAllocations handles GET /api/children/{id}/savings-goals/{goalId}/allocations
func (h *Handler) HandleListAllocations(w http.ResponseWriter, r *http.Request) {
	childID, err := parseChildID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_child_id", Message: "Invalid child ID."})
		return
	}

	goalID, err := parseGoalID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_goal_id", Message: "Invalid goal ID."})
		return
	}

	if status, errResp := h.verifyChildAccess(r, childID); errResp != nil {
		writeJSON(w, status, errResp)
		return
	}

	// Verify the goal belongs to this child
	goal, err := h.goalRepo.GetByID(goalID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to lookup goal."})
		return
	}
	if goal == nil || goal.ChildID != childID {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Goal not found."})
		return
	}

	allocations, err := h.goalAllocationRepo.ListByGoal(goalID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to list allocations."})
		return
	}

	if allocations == nil {
		allocations = []*models.GoalAllocation{}
	}

	writeJSON(w, http.StatusOK, AllocationsListResponse{
		Allocations: allocations,
	})
}
