package interest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bank-of-dad/internal/allowance"
	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/store"
)

// Handler handles interest-related HTTP requests.
type Handler struct {
	interestStore         *store.InterestStore
	interestScheduleStore *store.InterestScheduleStore
	childStore            *store.ChildStore
}

// NewHandler creates a new interest handler.
func NewHandler(interestStore *store.InterestStore, childStore *store.ChildStore, interestScheduleStore *store.InterestScheduleStore) *Handler {
	return &Handler{
		interestStore:         interestStore,
		interestScheduleStore: interestScheduleStore,
		childStore:            childStore,
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

// SetInterestRequest represents the combined request body for setting interest rate and schedule.
type SetInterestRequest struct {
	InterestRateBps int              `json:"interest_rate_bps"`
	Frequency       store.Frequency  `json:"frequency,omitempty"`
	DayOfWeek       *int             `json:"day_of_week,omitempty"`
	DayOfMonth      *int             `json:"day_of_month,omitempty"`
}

// SetInterestResponse represents the combined response after setting interest rate and schedule.
type SetInterestResponse struct {
	InterestRateBps     int                     `json:"interest_rate_bps"`
	InterestRateDisplay string                  `json:"interest_rate_display"`
	Schedule            *store.InterestSchedule `json:"schedule"`
}

// HandleSetInterest handles PUT /api/children/{id}/interest
// It sets the interest rate and manages the payout schedule atomically:
// - rate > 0: schedule fields are required, creates/updates the schedule
// - rate == 0: disables interest and deletes any existing schedule
func (h *Handler) HandleSetInterest(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
		return
	}

	childID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_child_id", Message: "Invalid child ID."})
		return
	}

	child, err := h.childStore.GetByID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to lookup child."})
		return
	}
	if child == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Child not found."})
		return
	}

	familyID := middleware.GetFamilyID(r)
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
		return
	}

	var req SetInterestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "Invalid request body."})
		return
	}

	if req.InterestRateBps < 0 || req.InterestRateBps > 10000 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid interest rate", Message: "Interest rate must be between 0% and 100%."})
		return
	}

	// When rate > 0, schedule fields are required
	if req.InterestRateBps > 0 {
		if errMsg := allowance.ValidateFrequencyAndDay(req.Frequency, req.DayOfWeek, req.DayOfMonth); errMsg != "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_schedule", Message: errMsg})
			return
		}
	}

	// Set the interest rate
	if err := h.interestStore.SetInterestRate(childID, req.InterestRateBps); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to set interest rate."})
		return
	}

	var schedule *store.InterestSchedule

	if req.InterestRateBps == 0 {
		// Rate is 0: delete any existing schedule
		existing, err := h.interestScheduleStore.GetByChildID(childID)
		if err == nil && existing != nil {
			h.interestScheduleStore.Delete(existing.ID) //nolint:errcheck // best-effort cleanup
		}
	} else {
		// Rate > 0: create or update schedule
		parentID := middleware.GetUserID(r)
		existing, err := h.interestScheduleStore.GetByChildID(childID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to check existing schedule."})
			return
		}

		if existing != nil {
			existing.Frequency = req.Frequency
			existing.DayOfWeek = req.DayOfWeek
			existing.DayOfMonth = req.DayOfMonth
			nextRun := calculateInterestNextRun(existing, time.Now().UTC())
			existing.NextRunAt = &nextRun

			schedule, err = h.interestScheduleStore.Update(existing)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to update schedule."})
				return
			}
		} else {
			sched := &store.InterestSchedule{
				ChildID:    childID,
				ParentID:   parentID,
				Frequency:  req.Frequency,
				DayOfWeek:  req.DayOfWeek,
				DayOfMonth: req.DayOfMonth,
				Status:     store.ScheduleStatusActive,
			}
			nextRun := calculateInterestNextRun(sched, time.Now().UTC())
			sched.NextRunAt = &nextRun

			schedule, err = h.interestScheduleStore.Create(sched)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to create schedule."})
				return
			}
		}
	}

	writeJSON(w, http.StatusOK, SetInterestResponse{
		InterestRateBps:     req.InterestRateBps,
		InterestRateDisplay: FormatRateDisplay(req.InterestRateBps),
		Schedule:            schedule,
	})
}

// SetInterestScheduleRequest represents the request body for setting an interest schedule.
type SetInterestScheduleRequest struct {
	Frequency  store.Frequency `json:"frequency"`
	DayOfWeek  *int            `json:"day_of_week,omitempty"`
	DayOfMonth *int            `json:"day_of_month,omitempty"`
}

// HandleSetInterestSchedule handles PUT /api/children/{childId}/interest-schedule
func (h *Handler) HandleSetInterestSchedule(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Only parents can manage interest schedules."})
		return
	}

	childID, err := strconv.ParseInt(r.PathValue("childId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_child_id", Message: "Invalid child ID."})
		return
	}

	child, err := h.childStore.GetByID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to lookup child."})
		return
	}
	if child == nil || child.FamilyID != middleware.GetFamilyID(r) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Child not found."})
		return
	}

	var req SetInterestScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "Invalid request body."})
		return
	}

	if errMsg := allowance.ValidateFrequencyAndDay(req.Frequency, req.DayOfWeek, req.DayOfMonth); errMsg != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_frequency", Message: errMsg})
		return
	}

	parentID := middleware.GetUserID(r)

	existing, err := h.interestScheduleStore.GetByChildID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to check existing schedule."})
		return
	}

	if existing != nil {
		existing.Frequency = req.Frequency
		existing.DayOfWeek = req.DayOfWeek
		existing.DayOfMonth = req.DayOfMonth
		nextRun := calculateInterestNextRun(existing, time.Now().UTC())
		existing.NextRunAt = &nextRun

		updated, err := h.interestScheduleStore.Update(existing)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to update schedule."})
			return
		}
		writeJSON(w, http.StatusOK, updated)
	} else {
		sched := &store.InterestSchedule{
			ChildID:   childID,
			ParentID:  parentID,
			Frequency: req.Frequency,
			DayOfWeek: req.DayOfWeek,
			DayOfMonth: req.DayOfMonth,
			Status:    store.ScheduleStatusActive,
		}
		nextRun := calculateInterestNextRun(sched, time.Now().UTC())
		sched.NextRunAt = &nextRun

		created, err := h.interestScheduleStore.Create(sched)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to create schedule."})
			return
		}
		writeJSON(w, http.StatusOK, created)
	}
}

// HandleGetInterestSchedule handles GET /api/children/{childId}/interest-schedule
func (h *Handler) HandleGetInterestSchedule(w http.ResponseWriter, r *http.Request) {
	childID, err := strconv.ParseInt(r.PathValue("childId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_child_id", Message: "Invalid child ID."})
		return
	}

	child, err := h.childStore.GetByID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to lookup child."})
		return
	}
	if child == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Child not found."})
		return
	}

	familyID := middleware.GetFamilyID(r)
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission."})
		return
	}

	userType := middleware.GetUserType(r)
	userID := middleware.GetUserID(r)
	if userType == "child" && userID != childID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You can only view your own interest schedule."})
		return
	}

	sched, err := h.interestScheduleStore.GetByChildID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get schedule."})
		return
	}

	writeJSON(w, http.StatusOK, sched)
}

// HandleDeleteInterestSchedule handles DELETE /api/children/{childId}/interest-schedule
func (h *Handler) HandleDeleteInterestSchedule(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Only parents can delete interest schedules."})
		return
	}

	childID, err := strconv.ParseInt(r.PathValue("childId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_child_id", Message: "Invalid child ID."})
		return
	}

	child, err := h.childStore.GetByID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to lookup child."})
		return
	}
	if child == nil || child.FamilyID != middleware.GetFamilyID(r) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Child not found."})
		return
	}

	sched, err := h.interestScheduleStore.GetByChildID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get schedule."})
		return
	}
	if sched == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "No interest schedule found for this child."})
		return
	}

	if err := h.interestScheduleStore.Delete(sched.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to delete schedule."})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// calculateInterestNextRun reuses the allowance schedule calculation logic for interest schedules.
func calculateInterestNextRun(sched *store.InterestSchedule, now time.Time) time.Time {
	// Create a temporary AllowanceSchedule to reuse CalculateNextRun
	tmpSched := &store.AllowanceSchedule{
		Frequency:  sched.Frequency,
		DayOfWeek:  sched.DayOfWeek,
		DayOfMonth: sched.DayOfMonth,
	}
	return allowance.CalculateNextRun(tmpSched, now)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck // best-effort response encoding
}
