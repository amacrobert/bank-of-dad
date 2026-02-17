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
	familyStore           *store.FamilyStore
}

// NewHandler creates a new interest handler.
func NewHandler(interestStore *store.InterestStore, childStore *store.ChildStore, interestScheduleStore *store.InterestScheduleStore, familyStore *store.FamilyStore) *Handler {
	return &Handler{
		interestStore:         interestStore,
		interestScheduleStore: interestScheduleStore,
		childStore:            childStore,
		familyStore:           familyStore,
	}
}

// getFamilyTimezone loads the *time.Location for a family, falling back to UTC.
func (h *Handler) getFamilyTimezone(familyID int64) *time.Location {
	tz, err := h.familyStore.GetTimezone(familyID)
	if err != nil || tz == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
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
		loc := h.getFamilyTimezone(familyID)
		existing, err := h.interestScheduleStore.GetByChildID(childID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to check existing schedule."})
			return
		}

		if existing != nil {
			existing.Frequency = req.Frequency
			existing.DayOfWeek = req.DayOfWeek
			existing.DayOfMonth = req.DayOfMonth
			nextRun := calculateInterestNextRun(existing, time.Now().UTC(), loc)
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
			nextRun := calculateInterestNextRun(sched, time.Now().UTC(), loc)
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

// calculateInterestNextRun reuses the allowance schedule calculation logic for interest schedules.
func calculateInterestNextRun(sched *store.InterestSchedule, now time.Time, loc *time.Location) time.Time {
	// Create a temporary AllowanceSchedule to reuse CalculateNextRun
	tmpSched := &store.AllowanceSchedule{
		Frequency:  sched.Frequency,
		DayOfWeek:  sched.DayOfWeek,
		DayOfMonth: sched.DayOfMonth,
	}
	return allowance.CalculateNextRun(tmpSched, now, loc)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck // best-effort response encoding
}
