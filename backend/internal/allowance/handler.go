package allowance

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
	MaxAmountCents = 99999999 // $999,999.99
	MaxNoteLength  = 500
)

// Handler handles allowance schedule HTTP requests.
type Handler struct {
	scheduleStore *store.ScheduleStore
	childStore    *store.ChildStore
}

// NewHandler creates a new allowance handler.
func NewHandler(scheduleStore *store.ScheduleStore, childStore *store.ChildStore) *Handler {
	return &Handler{
		scheduleStore: scheduleStore,
		childStore:    childStore,
	}
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// CreateScheduleRequest represents a request to create a schedule.
type CreateScheduleRequest struct {
	ChildID     int64           `json:"child_id"`
	AmountCents int64           `json:"amount_cents"`
	Frequency   store.Frequency `json:"frequency"`
	DayOfWeek   *int            `json:"day_of_week,omitempty"`
	DayOfMonth  *int            `json:"day_of_month,omitempty"`
	Note        string          `json:"note,omitempty"`
}

// UpdateScheduleRequest represents a request to update a schedule.
type UpdateScheduleRequest struct {
	AmountCents *int64           `json:"amount_cents,omitempty"`
	Frequency   *store.Frequency `json:"frequency,omitempty"`
	DayOfWeek   *int             `json:"day_of_week,omitempty"`
	DayOfMonth  *int             `json:"day_of_month,omitempty"`
	Note        *string          `json:"note,omitempty"`
}

// ScheduleListResponse wraps a list of schedules with child names.
type ScheduleListResponse struct {
	Schedules []store.ScheduleWithChild `json:"schedules"`
}

// UpcomingAllowancesResponse wraps a list of upcoming allowances.
type UpcomingAllowancesResponse struct {
	Allowances []store.UpcomingAllowance `json:"allowances"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// HandleCreateSchedule handles POST /api/schedules
func (h *Handler) HandleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can create allowance schedules.",
		})
		return
	}

	var req CreateScheduleRequest
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

	// Validate frequency and day
	if err := ValidateFrequencyAndDay(req.Frequency, req.DayOfWeek, req.DayOfMonth); err != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_frequency",
			Message: err,
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

	// Verify child exists and belongs to parent's family
	child, dbErr := h.childStore.GetByID(req.ChildID)
	if dbErr != nil {
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

	familyID := middleware.GetFamilyID(r)
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Child not found.",
		})
		return
	}

	parentID := middleware.GetUserID(r)

	// Build schedule
	sched := &store.AllowanceSchedule{
		ChildID:     req.ChildID,
		ParentID:    parentID,
		AmountCents: req.AmountCents,
		Frequency:   req.Frequency,
		DayOfWeek:   req.DayOfWeek,
		DayOfMonth:  req.DayOfMonth,
		Status:      store.ScheduleStatusActive,
	}
	if note != "" {
		sched.Note = &note
	}

	// Calculate next run
	nextRun := CalculateNextRun(sched, time.Now().UTC())
	sched.NextRunAt = &nextRun

	created, dbErr := h.scheduleStore.Create(sched)
	if dbErr != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create schedule.",
		})
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// HandleListSchedules handles GET /api/schedules
func (h *Handler) HandleListSchedules(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can view schedules.",
		})
		return
	}

	familyID := middleware.GetFamilyID(r)
	schedules, err := h.scheduleStore.ListByParentFamily(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list schedules.",
		})
		return
	}

	if schedules == nil {
		schedules = []store.ScheduleWithChild{}
	}

	writeJSON(w, http.StatusOK, ScheduleListResponse{Schedules: schedules})
}

// HandleGetSchedule handles GET /api/schedules/{id}
func (h *Handler) HandleGetSchedule(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can view schedules.",
		})
		return
	}

	scheduleID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid schedule ID.",
		})
		return
	}

	sched, err := h.scheduleStore.GetByID(scheduleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get schedule.",
		})
		return
	}
	if sched == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Schedule not found.",
		})
		return
	}

	// Verify the schedule belongs to the parent's family
	child, err := h.childStore.GetByID(sched.ChildID)
	if err != nil || child == nil || child.FamilyID != middleware.GetFamilyID(r) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Schedule not found.",
		})
		return
	}

	writeJSON(w, http.StatusOK, sched)
}

// HandleUpdateSchedule handles PUT /api/schedules/{id}
func (h *Handler) HandleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can update schedules.",
		})
		return
	}

	scheduleID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid schedule ID.",
		})
		return
	}

	sched, err := h.scheduleStore.GetByID(scheduleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get schedule.",
		})
		return
	}
	if sched == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Schedule not found.",
		})
		return
	}

	// Verify family ownership
	child, err := h.childStore.GetByID(sched.ChildID)
	if err != nil || child == nil || child.FamilyID != middleware.GetFamilyID(r) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Schedule not found.",
		})
		return
	}

	var req UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body.",
		})
		return
	}

	// Apply updates
	if req.AmountCents != nil {
		if *req.AmountCents <= 0 || *req.AmountCents > MaxAmountCents {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_amount",
				Message: "Amount must be between 1 cent and $999,999.99.",
			})
			return
		}
		sched.AmountCents = *req.AmountCents
	}

	if req.Frequency != nil {
		sched.Frequency = *req.Frequency
	}
	if req.DayOfWeek != nil {
		sched.DayOfWeek = req.DayOfWeek
	}
	if req.DayOfMonth != nil {
		sched.DayOfMonth = req.DayOfMonth
	}

	// Validate frequency/day combination
	if errMsg := ValidateFrequencyAndDay(sched.Frequency, sched.DayOfWeek, sched.DayOfMonth); errMsg != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_frequency",
			Message: errMsg,
		})
		return
	}

	if req.Note != nil {
		trimmed := strings.TrimSpace(*req.Note)
		if len(trimmed) > MaxNoteLength {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_note",
				Message: "Note must be 500 characters or less.",
			})
			return
		}
		if trimmed == "" {
			sched.Note = nil
		} else {
			sched.Note = &trimmed
		}
	}

	// Recalculate next_run_at if frequency or day changed
	if req.Frequency != nil || req.DayOfWeek != nil || req.DayOfMonth != nil {
		nextRun := CalculateNextRun(sched, time.Now().UTC())
		sched.NextRunAt = &nextRun
	}

	updated, err := h.scheduleStore.Update(sched)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update schedule.",
		})
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// HandleDeleteSchedule handles DELETE /api/schedules/{id}
func (h *Handler) HandleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can delete schedules.",
		})
		return
	}

	scheduleID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid schedule ID.",
		})
		return
	}

	sched, err := h.scheduleStore.GetByID(scheduleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get schedule.",
		})
		return
	}
	if sched == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Schedule not found.",
		})
		return
	}

	child, err := h.childStore.GetByID(sched.ChildID)
	if err != nil || child == nil || child.FamilyID != middleware.GetFamilyID(r) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Schedule not found.",
		})
		return
	}

	if err := h.scheduleStore.Delete(scheduleID); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete schedule.",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandlePauseSchedule handles POST /api/schedules/{id}/pause
func (h *Handler) HandlePauseSchedule(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can pause schedules.",
		})
		return
	}

	scheduleID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid schedule ID.",
		})
		return
	}

	sched, err := h.scheduleStore.GetByID(scheduleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get schedule.",
		})
		return
	}
	if sched == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Schedule not found.",
		})
		return
	}

	child, err := h.childStore.GetByID(sched.ChildID)
	if err != nil || child == nil || child.FamilyID != middleware.GetFamilyID(r) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Schedule not found.",
		})
		return
	}

	if sched.Status == store.ScheduleStatusPaused {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "already_paused",
			Message: "Schedule is already paused.",
		})
		return
	}

	if err := h.scheduleStore.UpdateStatus(scheduleID, store.ScheduleStatusPaused); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to pause schedule.",
		})
		return
	}

	updated, _ := h.scheduleStore.GetByID(scheduleID)
	writeJSON(w, http.StatusOK, updated)
}

// HandleResumeSchedule handles POST /api/schedules/{id}/resume
func (h *Handler) HandleResumeSchedule(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can resume schedules.",
		})
		return
	}

	scheduleID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid schedule ID.",
		})
		return
	}

	sched, err := h.scheduleStore.GetByID(scheduleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get schedule.",
		})
		return
	}
	if sched == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Schedule not found.",
		})
		return
	}

	child, err := h.childStore.GetByID(sched.ChildID)
	if err != nil || child == nil || child.FamilyID != middleware.GetFamilyID(r) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Schedule not found.",
		})
		return
	}

	if sched.Status == store.ScheduleStatusActive {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "already_active",
			Message: "Schedule is already active.",
		})
		return
	}

	if err := h.scheduleStore.UpdateStatus(scheduleID, store.ScheduleStatusActive); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to resume schedule.",
		})
		return
	}

	// Recalculate next_run_at from now
	nextRun := CalculateNextRun(sched, time.Now().UTC())
	if err := h.scheduleStore.UpdateNextRunAt(scheduleID, nextRun); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update schedule.",
		})
		return
	}

	updated, _ := h.scheduleStore.GetByID(scheduleID)
	writeJSON(w, http.StatusOK, updated)
}

// HandleGetUpcomingAllowances handles GET /api/children/{childId}/upcoming-allowances
func (h *Handler) HandleGetUpcomingAllowances(w http.ResponseWriter, r *http.Request) {
	childIDStr := r.PathValue("childId")
	childID, err := strconv.ParseInt(childIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_child_id",
			Message: "Invalid child ID.",
		})
		return
	}

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

	// Authorization: child can only see their own; parent can see any child in family
	userType := middleware.GetUserType(r)
	userID := middleware.GetUserID(r)
	familyID := middleware.GetFamilyID(r)

	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have permission to view this child's allowances.",
		})
		return
	}

	if userType == "child" && userID != childID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You can only view your own upcoming allowances.",
		})
		return
	}

	schedules, err := h.scheduleStore.ListActiveByChild(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get upcoming allowances.",
		})
		return
	}

	allowances := make([]store.UpcomingAllowance, 0, len(schedules))
	for _, s := range schedules {
		if s.NextRunAt != nil {
			allowances = append(allowances, store.UpcomingAllowance{
				AmountCents: s.AmountCents,
				NextDate:    *s.NextRunAt,
				Note:        s.Note,
			})
		}
	}

	writeJSON(w, http.StatusOK, UpcomingAllowancesResponse{Allowances: allowances})
}

// SetChildAllowanceRequest represents a request to create or update a child's allowance.
type SetChildAllowanceRequest struct {
	AmountCents int64           `json:"amount_cents"`
	Frequency   store.Frequency `json:"frequency"`
	DayOfWeek   *int            `json:"day_of_week,omitempty"`
	DayOfMonth  *int            `json:"day_of_month,omitempty"`
	Note        string          `json:"note,omitempty"`
}

// HandleGetChildAllowance handles GET /api/children/{childId}/allowance
func (h *Handler) HandleGetChildAllowance(w http.ResponseWriter, r *http.Request) {
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
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission to view this child's allowance."})
		return
	}

	userType := middleware.GetUserType(r)
	userID := middleware.GetUserID(r)
	if userType == "child" && userID != childID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You can only view your own allowance."})
		return
	}

	sched, err := h.scheduleStore.GetByChildID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get allowance."})
		return
	}

	writeJSON(w, http.StatusOK, sched)
}

// HandleSetChildAllowance handles PUT /api/children/{childId}/allowance
func (h *Handler) HandleSetChildAllowance(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Only parents can manage allowances."})
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
	if child == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Child not found."})
		return
	}

	familyID := middleware.GetFamilyID(r)
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Child not found."})
		return
	}

	var req SetChildAllowanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "Invalid request body."})
		return
	}

	if req.AmountCents <= 0 || req.AmountCents > MaxAmountCents {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_amount", Message: "Amount must be between 1 cent and $999,999.99."})
		return
	}

	if errMsg := ValidateFrequencyAndDay(req.Frequency, req.DayOfWeek, req.DayOfMonth); errMsg != "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_frequency", Message: errMsg})
		return
	}

	note := strings.TrimSpace(req.Note)
	if len(note) > MaxNoteLength {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_note", Message: "Note must be 500 characters or less."})
		return
	}

	parentID := middleware.GetUserID(r)

	// Check if schedule already exists for this child
	existing, err := h.scheduleStore.GetByChildID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to check existing allowance."})
		return
	}

	if existing != nil {
		// Update existing schedule
		existing.AmountCents = req.AmountCents
		existing.Frequency = req.Frequency
		existing.DayOfWeek = req.DayOfWeek
		existing.DayOfMonth = req.DayOfMonth
		if note != "" {
			existing.Note = &note
		} else {
			existing.Note = nil
		}
		nextRun := CalculateNextRun(existing, time.Now().UTC())
		existing.NextRunAt = &nextRun

		updated, err := h.scheduleStore.Update(existing)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to update allowance."})
			return
		}
		writeJSON(w, http.StatusOK, updated)
	} else {
		// Create new schedule
		sched := &store.AllowanceSchedule{
			ChildID:     childID,
			ParentID:    parentID,
			AmountCents: req.AmountCents,
			Frequency:   req.Frequency,
			DayOfWeek:   req.DayOfWeek,
			DayOfMonth:  req.DayOfMonth,
			Status:      store.ScheduleStatusActive,
		}
		if note != "" {
			sched.Note = &note
		}
		nextRun := CalculateNextRun(sched, time.Now().UTC())
		sched.NextRunAt = &nextRun

		created, err := h.scheduleStore.Create(sched)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to create allowance."})
			return
		}
		writeJSON(w, http.StatusOK, created)
	}
}

// HandleDeleteChildAllowance handles DELETE /api/children/{childId}/allowance
func (h *Handler) HandleDeleteChildAllowance(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Only parents can delete allowances."})
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

	sched, err := h.scheduleStore.GetByChildID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get allowance."})
		return
	}
	if sched == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "No allowance schedule found for this child."})
		return
	}

	if err := h.scheduleStore.Delete(sched.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to delete allowance."})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandlePauseChildAllowance handles POST /api/children/{childId}/allowance/pause
func (h *Handler) HandlePauseChildAllowance(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Only parents can pause allowances."})
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

	sched, err := h.scheduleStore.GetByChildID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get allowance."})
		return
	}
	if sched == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "No allowance schedule found for this child."})
		return
	}

	if sched.Status == store.ScheduleStatusPaused {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "already_paused", Message: "Allowance is already paused."})
		return
	}

	if err := h.scheduleStore.UpdateStatus(sched.ID, store.ScheduleStatusPaused); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to pause allowance."})
		return
	}

	updated, _ := h.scheduleStore.GetByID(sched.ID)
	writeJSON(w, http.StatusOK, updated)
}

// HandleResumeChildAllowance handles POST /api/children/{childId}/allowance/resume
func (h *Handler) HandleResumeChildAllowance(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	if userType != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Only parents can resume allowances."})
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

	sched, err := h.scheduleStore.GetByChildID(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get allowance."})
		return
	}
	if sched == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "No allowance schedule found for this child."})
		return
	}

	if sched.Status == store.ScheduleStatusActive {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "already_active", Message: "Allowance is already active."})
		return
	}

	if err := h.scheduleStore.UpdateStatus(sched.ID, store.ScheduleStatusActive); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to resume allowance."})
		return
	}

	nextRun := CalculateNextRun(sched, time.Now().UTC())
	if err := h.scheduleStore.UpdateNextRunAt(sched.ID, nextRun); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to update allowance."})
		return
	}

	updated, _ := h.scheduleStore.GetByID(sched.ID)
	writeJSON(w, http.StatusOK, updated)
}

// ValidateFrequencyAndDay returns an error message if the frequency/day combination is invalid.
func ValidateFrequencyAndDay(freq store.Frequency, dayOfWeek *int, dayOfMonth *int) string {
	switch freq {
	case store.FrequencyWeekly, store.FrequencyBiweekly:
		if dayOfWeek == nil {
			return "day_of_week is required for weekly and biweekly schedules."
		}
		if *dayOfWeek < 0 || *dayOfWeek > 6 {
			return "day_of_week must be between 0 (Sunday) and 6 (Saturday)."
		}
	case store.FrequencyMonthly:
		if dayOfMonth == nil {
			return "day_of_month is required for monthly schedules."
		}
		if *dayOfMonth < 1 || *dayOfMonth > 31 {
			return "day_of_month must be between 1 and 31."
		}
	default:
		return "Frequency must be 'weekly', 'biweekly', or 'monthly'."
	}
	return ""
}
