package chore

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/notification"
	"bank-of-dad/models"
	"bank-of-dad/repositories"
)

const (
	MaxNameLength        = 100
	MaxDescriptionLength = 500
	MaxRewardCents       = 99999999
)

// Handler handles chore-related HTTP requests.
type Handler struct {
	choreRepo         *repositories.ChoreRepo
	choreInstanceRepo *repositories.ChoreInstanceRepo
	txRepo            *repositories.TransactionRepo
	childRepo         *repositories.ChildRepo
	notifier          *notification.Service
	parentRepo        *repositories.ParentRepo
	familyRepo        *repositories.FamilyRepo
}

// NewHandler creates a new chore handler.
func NewHandler(choreRepo *repositories.ChoreRepo, choreInstanceRepo *repositories.ChoreInstanceRepo, txRepo *repositories.TransactionRepo, childRepo *repositories.ChildRepo, notifier *notification.Service, parentRepo *repositories.ParentRepo, familyRepo *repositories.FamilyRepo) *Handler {
	return &Handler{
		choreRepo:         choreRepo,
		choreInstanceRepo: choreInstanceRepo,
		txRepo:            txRepo,
		childRepo:         childRepo,
		notifier:          notifier,
		parentRepo:        parentRepo,
		familyRepo:        familyRepo,
	}
}

// CreateChoreRequest represents the request body for creating a chore.
type CreateChoreRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	RewardCents int     `json:"reward_cents"`
	Recurrence  string  `json:"recurrence"`
	DayOfWeek   *int    `json:"day_of_week,omitempty"`
	DayOfMonth  *int    `json:"day_of_month,omitempty"`
	ChildIDs    []int64 `json:"child_ids"`
}

// ChoreResponse represents a chore in API responses.
type ChoreResponse struct {
	ID          int64                `json:"id"`
	FamilyID    int64                `json:"family_id"`
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
	RewardCents int                  `json:"reward_cents"`
	Recurrence  string               `json:"recurrence"`
	DayOfWeek   *int                 `json:"day_of_week,omitempty"`
	DayOfMonth  *int                 `json:"day_of_month,omitempty"`
	IsActive    bool                 `json:"is_active"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Assignments []AssignmentResponse `json:"assignments"`
	PendingCount int                 `json:"pending_count,omitempty"`
}

// AssignmentResponse represents a chore assignment in API responses.
type AssignmentResponse struct {
	ChildID   int64  `json:"child_id"`
	ChildName string `json:"child_name"`
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

// HandleCreateChore handles POST /api/chores
func (h *Handler) HandleCreateChore(w http.ResponseWriter, r *http.Request) {
	// Auth: parent only
	if middleware.GetUserType(r) != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can create chores.",
		})
		return
	}

	// Parse request body
	var req CreateChoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body.",
		})
		return
	}

	// Validate name
	name := strings.TrimSpace(req.Name)
	if len(name) == 0 || len(name) > MaxNameLength {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_name",
			Message: "Name must be between 1 and 100 characters.",
		})
		return
	}

	// Validate description
	desc := strings.TrimSpace(req.Description)
	if len(desc) > MaxDescriptionLength {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Description must be 500 characters or less.",
		})
		return
	}

	// Validate reward_cents
	if req.RewardCents < 0 || req.RewardCents > MaxRewardCents {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_amount",
			Message: "Reward must be between 0 and $999,999.99.",
		})
		return
	}

	// Validate recurrence
	validRecurrences := map[string]bool{
		"one_time": true,
		"daily":    true,
		"weekly":   true,
		"monthly":  true,
	}
	if !validRecurrences[req.Recurrence] {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_recurrence",
			Message: "Recurrence must be one of: one_time, daily, weekly, monthly.",
		})
		return
	}

	// Validate day_of_week for weekly
	if req.Recurrence == "weekly" {
		if req.DayOfWeek == nil || *req.DayOfWeek < 0 || *req.DayOfWeek > 6 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_recurrence",
				Message: "Weekly chores require day_of_week (0-6).",
			})
			return
		}
	}

	// Validate day_of_month for monthly
	if req.Recurrence == "monthly" {
		if req.DayOfMonth == nil || *req.DayOfMonth < 1 || *req.DayOfMonth > 31 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_recurrence",
				Message: "Monthly chores require day_of_month (1-31).",
			})
			return
		}
	}

	// Validate child_ids
	if len(req.ChildIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_children",
			Message: "At least one child must be assigned.",
		})
		return
	}

	// Verify each child exists and belongs to the same family
	familyID := middleware.GetFamilyID(r)
	parentID := middleware.GetUserID(r)

	childNames := make(map[int64]string)
	for _, childID := range req.ChildIDs {
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
				Error:   "child_not_found",
				Message: "Child not found.",
			})
			return
		}
		if child.FamilyID != familyID {
			writeJSON(w, http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "Child does not belong to your family.",
			})
			return
		}
		childNames[childID] = child.FirstName
	}

	// Build description pointer
	var descPtr *string
	if desc != "" {
		descPtr = &desc
	}

	// Create chore
	chore := &models.Chore{
		FamilyID:          familyID,
		CreatedByParentID: parentID,
		Name:              name,
		Description:       descPtr,
		RewardCents:       req.RewardCents,
		Recurrence:        models.ChoreRecurrence(req.Recurrence),
		DayOfWeek:         req.DayOfWeek,
		DayOfMonth:        req.DayOfMonth,
		IsActive:          true,
	}

	createdChore, err := h.choreRepo.Create(chore)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create chore.",
		})
		return
	}

	// Create assignments
	assignments := make([]AssignmentResponse, 0, len(req.ChildIDs))
	for _, childID := range req.ChildIDs {
		assignment := &models.ChoreAssignment{
			ChoreID: createdChore.ID,
			ChildID: childID,
		}
		if _, err := h.choreRepo.CreateAssignment(assignment); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to create chore assignment.",
			})
			return
		}
		assignments = append(assignments, AssignmentResponse{
			ChildID:   childID,
			ChildName: childNames[childID],
		})
	}

	// For one-time chores, create instances immediately
	if req.Recurrence == "one_time" {
		for _, childID := range req.ChildIDs {
			instance := &models.ChoreInstance{
				ChoreID:     createdChore.ID,
				ChildID:     childID,
				RewardCents: createdChore.RewardCents,
				Status:      models.ChoreInstanceStatusAvailable,
			}
			if _, err := h.choreInstanceRepo.CreateInstance(instance); err != nil {
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{
					Error:   "internal_error",
					Message: "Failed to create chore instance.",
				})
				return
			}
		}
	}

	// Build response
	resp := ChoreResponse{
		ID:           createdChore.ID,
		FamilyID:     createdChore.FamilyID,
		Name:         createdChore.Name,
		Description:  createdChore.Description,
		RewardCents:  createdChore.RewardCents,
		Recurrence:   string(createdChore.Recurrence),
		DayOfWeek:    createdChore.DayOfWeek,
		DayOfMonth:   createdChore.DayOfMonth,
		IsActive:     createdChore.IsActive,
		CreatedAt:    createdChore.CreatedAt,
		UpdatedAt:    createdChore.UpdatedAt,
		Assignments:  assignments,
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"chore": resp})
}

// HandleListChores handles GET /api/chores
func (h *Handler) HandleListChores(w http.ResponseWriter, r *http.Request) {
	// Auth: parent only
	if middleware.GetUserType(r) != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can list chores.",
		})
		return
	}

	familyID := middleware.GetFamilyID(r)

	choresWithAssignments, err := h.choreRepo.ListByFamily(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list chores.",
		})
		return
	}

	// Convert to response format
	chores := make([]ChoreResponse, len(choresWithAssignments))
	for i, cwa := range choresWithAssignments {
		assignments := make([]AssignmentResponse, len(cwa.Assignments))
		for j, a := range cwa.Assignments {
			assignments[j] = AssignmentResponse{
				ChildID:   a.ChildID,
				ChildName: a.ChildName,
			}
		}

		chores[i] = ChoreResponse{
			ID:           cwa.ID,
			FamilyID:     cwa.FamilyID,
			Name:         cwa.Name,
			Description:  cwa.Description,
			RewardCents:  cwa.RewardCents,
			Recurrence:   string(cwa.Recurrence),
			DayOfWeek:    cwa.DayOfWeek,
			DayOfMonth:   cwa.DayOfMonth,
			IsActive:     cwa.IsActive,
			CreatedAt:    cwa.CreatedAt,
			UpdatedAt:    cwa.UpdatedAt,
			Assignments:  assignments,
			PendingCount: cwa.PendingCount,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"chores": chores})
}

// InstanceResponse represents a chore instance in API responses.
type InstanceResponse struct {
	ID               int64                      `json:"id"`
	ChoreID          int64                      `json:"chore_id"`
	ChoreName        string                     `json:"chore_name,omitempty"`
	ChoreDescription *string                    `json:"chore_description,omitempty"`
	ChildID          int64                      `json:"child_id"`
	ChildName        string                     `json:"child_name,omitempty"`
	RewardCents      int                        `json:"reward_cents"`
	Status           models.ChoreInstanceStatus `json:"status"`
	PeriodStart      *time.Time                 `json:"period_start,omitempty"`
	PeriodEnd        *time.Time                 `json:"period_end,omitempty"`
	CompletedAt      *time.Time                 `json:"completed_at,omitempty"`
	ReviewedAt       *time.Time                 `json:"reviewed_at,omitempty"`
	RejectionReason  *string                    `json:"rejection_reason,omitempty"`
	TransactionID    *int64                     `json:"transaction_id,omitempty"`
	CreatedAt        time.Time                  `json:"created_at"`
}

// RejectRequest represents the request body for rejecting a chore instance.
type RejectRequest struct {
	Reason string `json:"reason,omitempty"`
}

// HandleChildListChores handles GET /api/child/chores
func (h *Handler) HandleChildListChores(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserType(r) != "child" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only children can access this endpoint.",
		})
		return
	}

	childID := middleware.GetUserID(r)

	available, pending, completed, err := h.choreInstanceRepo.ListByChild(childID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list chores.",
		})
		return
	}

	toInstanceResponses := func(items []repositories.ChoreInstanceWithDetails) []InstanceResponse {
		result := make([]InstanceResponse, len(items))
		for i, item := range items {
			result[i] = InstanceResponse{
				ID:               item.ID,
				ChoreID:          item.ChoreID,
				ChoreName:        item.ChoreName,
				ChoreDescription: item.ChoreDescription,
				ChildID:          item.ChildID,
				RewardCents:      item.RewardCents,
				Status:           item.Status,
				PeriodStart:      item.PeriodStart,
				PeriodEnd:        item.PeriodEnd,
				CompletedAt:      item.CompletedAt,
				ReviewedAt:       item.ReviewedAt,
				RejectionReason:  item.RejectionReason,
				TransactionID:    item.TransactionID,
				CreatedAt:        item.CreatedAt,
			}
		}
		return result
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"available": toInstanceResponses(available),
		"pending":   toInstanceResponses(pending),
		"completed": toInstanceResponses(completed),
	})
}

// HandleCompleteChore handles POST /api/child/chores/{id}/complete
func (h *Handler) HandleCompleteChore(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserType(r) != "child" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only children can complete chores.",
		})
		return
	}

	instanceID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid instance ID.",
		})
		return
	}

	childID := middleware.GetUserID(r)

	if err := h.choreInstanceRepo.MarkComplete(instanceID, childID); err != nil {
		if errors.Is(err, repositories.ErrInvalidStatusTransition) {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_status",
				Message: "Cannot complete this chore.",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to complete chore.",
		})
		return
	}

	instance, err := h.choreInstanceRepo.GetByID(instanceID)
	if err != nil || instance == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve updated instance.",
		})
		return
	}

	// Queue chore completion notification (fire-and-forget via batcher)
	if h.notifier != nil {
		familyID := middleware.GetFamilyID(r)
		child, _ := h.childRepo.GetByID(childID)
		chore, _ := h.choreRepo.GetByID(instance.ChoreID)
		if child != nil && chore != nil {
			h.notifier.QueueChoreCompletion(familyID, childID, child.FirstName, chore.Name, instance.RewardCents)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"instance": InstanceResponse{
			ID:          instance.ID,
			ChoreID:     instance.ChoreID,
			ChildID:     instance.ChildID,
			RewardCents: instance.RewardCents,
			Status:      instance.Status,
			PeriodStart: instance.PeriodStart,
			PeriodEnd:   instance.PeriodEnd,
			CompletedAt: instance.CompletedAt,
			ReviewedAt:  instance.ReviewedAt,
			CreatedAt:   instance.CreatedAt,
		},
	})
}

// HandleListPending handles GET /api/chores/pending
func (h *Handler) HandleListPending(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserType(r) != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can view pending chores.",
		})
		return
	}

	familyID := middleware.GetFamilyID(r)

	pendingInstances, err := h.choreInstanceRepo.ListPendingByFamily(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list pending chores.",
		})
		return
	}

	instances := make([]InstanceResponse, len(pendingInstances))
	for i, p := range pendingInstances {
		instances[i] = InstanceResponse{
			ID:              p.ID,
			ChoreID:         p.ChoreID,
			ChoreName:       p.ChoreName,
			ChildID:         p.ChildID,
			ChildName:       p.ChildName,
			RewardCents:     p.RewardCents,
			Status:          p.Status,
			PeriodStart:     p.PeriodStart,
			PeriodEnd:       p.PeriodEnd,
			CompletedAt:     p.CompletedAt,
			ReviewedAt:      p.ReviewedAt,
			RejectionReason: p.RejectionReason,
			TransactionID:   p.TransactionID,
			CreatedAt:       p.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"instances": instances})
}

// HandleListCompleted handles GET /api/chores/completed?limit=10&offset=0
func (h *Handler) HandleListCompleted(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserType(r) != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can view completed chores.",
		})
		return
	}

	familyID := middleware.GetFamilyID(r)

	limit := 10
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	completedInstances, total, err := h.choreInstanceRepo.ListCompletedByFamily(familyID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list completed chores.",
		})
		return
	}

	instances := make([]InstanceResponse, len(completedInstances))
	for i, p := range completedInstances {
		instances[i] = InstanceResponse{
			ID:          p.ID,
			ChoreID:     p.ChoreID,
			ChoreName:   p.ChoreName,
			ChildID:     p.ChildID,
			ChildName:   p.ChildName,
			RewardCents: p.RewardCents,
			Status:      p.Status,
			PeriodStart: p.PeriodStart,
			PeriodEnd:   p.PeriodEnd,
			CompletedAt: p.CompletedAt,
			ReviewedAt:  p.ReviewedAt,
			CreatedAt:   p.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"instances": instances, "total": total})
}

// HandleApprove handles POST /api/chore-instances/{id}/approve
func (h *Handler) HandleApprove(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserType(r) != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can approve chores.",
		})
		return
	}

	instanceID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid instance ID.",
		})
		return
	}

	parentID := middleware.GetUserID(r)
	familyID := middleware.GetFamilyID(r)

	instance, err := h.choreInstanceRepo.GetByID(instanceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve instance.",
		})
		return
	}
	if instance == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Chore instance not found.",
		})
		return
	}

	// Verify instance belongs to parent's family
	child, err := h.childRepo.GetByID(instance.ChildID)
	if err != nil || child == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to verify family ownership.",
		})
		return
	}
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Instance does not belong to your family.",
		})
		return
	}

	// Verify status is pending_approval
	if instance.Status != models.ChoreInstanceStatusPendingApproval {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_status",
			Message: "Instance is not pending approval.",
		})
		return
	}

	// Check child is not disabled
	if child.IsDisabled {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "child_disabled",
			Message: "Child account is disabled.",
		})
		return
	}

	// Get chore name for transaction note
	chore, err := h.choreRepo.GetByID(instance.ChoreID)
	if err != nil || chore == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve chore.",
		})
		return
	}

	var transactionID *int64
	var newBalance int64

	if instance.RewardCents > 0 {
		tx, balance, err := h.txRepo.DepositChore(instance.ChildID, parentID, int64(instance.RewardCents), "Chore: "+chore.Name)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to create deposit.",
			})
			return
		}
		transactionID = &tx.ID
		newBalance = balance
	}

	if err := h.choreInstanceRepo.Approve(instanceID, parentID, transactionID); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to approve instance.",
		})
		return
	}

	// Get updated instance
	updated, err := h.choreInstanceRepo.GetByID(instanceID)
	if err != nil || updated == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve updated instance.",
		})
		return
	}

	// Notify other parents about the approval (fire-and-forget)
	if h.notifier != nil {
		bankName := "Bank of Dad"
		if h.familyRepo != nil {
			if bn, err := h.familyRepo.GetBankName(familyID); err == nil {
				bankName = bn
			}
		}
		actingParentName := ""
		if h.parentRepo != nil {
			if p, err := h.parentRepo.GetByID(parentID); err == nil && p != nil {
				actingParentName = p.DisplayName
			}
		}
		h.notifier.NotifyDecision(r.Context(), familyID, parentID, actingParentName, child.FirstName, "chore", "approved", 0, chore.Name, "", bankName)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"instance": InstanceResponse{
			ID:            updated.ID,
			ChoreID:       updated.ChoreID,
			ChildID:       updated.ChildID,
			RewardCents:   updated.RewardCents,
			Status:        updated.Status,
			PeriodStart:   updated.PeriodStart,
			PeriodEnd:     updated.PeriodEnd,
			CompletedAt:   updated.CompletedAt,
			ReviewedAt:    updated.ReviewedAt,
			TransactionID: updated.TransactionID,
			CreatedAt:     updated.CreatedAt,
		},
		"new_balance": newBalance,
	})
}

// HandleReject handles POST /api/chore-instances/{id}/reject
func (h *Handler) HandleReject(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserType(r) != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only parents can reject chores.",
		})
		return
	}

	instanceID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid instance ID.",
		})
		return
	}

	// Parse optional reason
	var reqBody RejectRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
	}
	reason := strings.TrimSpace(reqBody.Reason)
	if len(reason) > MaxDescriptionLength {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Reason must be 500 characters or less.",
		})
		return
	}

	parentID := middleware.GetUserID(r)
	familyID := middleware.GetFamilyID(r)

	instance, err := h.choreInstanceRepo.GetByID(instanceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve instance.",
		})
		return
	}
	if instance == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Chore instance not found.",
		})
		return
	}

	// Verify family ownership
	child, err := h.childRepo.GetByID(instance.ChildID)
	if err != nil || child == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to verify family ownership.",
		})
		return
	}
	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Instance does not belong to your family.",
		})
		return
	}

	// Verify status is pending_approval
	if instance.Status != models.ChoreInstanceStatusPendingApproval {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_status",
			Message: "Instance is not pending approval.",
		})
		return
	}

	if err := h.choreInstanceRepo.Reject(instanceID, parentID, reason); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to reject instance.",
		})
		return
	}

	// Get updated instance
	updated, err := h.choreInstanceRepo.GetByID(instanceID)
	if err != nil || updated == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve updated instance.",
		})
		return
	}

	// Notify other parents about the rejection (fire-and-forget)
	if h.notifier != nil {
		bankName := "Bank of Dad"
		if h.familyRepo != nil {
			if bn, err := h.familyRepo.GetBankName(familyID); err == nil {
				bankName = bn
			}
		}
		actingParentName := ""
		if h.parentRepo != nil {
			if p, err := h.parentRepo.GetByID(parentID); err == nil && p != nil {
				actingParentName = p.DisplayName
			}
		}
		choreName := ""
		if ch, err := h.choreRepo.GetByID(instance.ChoreID); err == nil && ch != nil {
			choreName = ch.Name
		}
		h.notifier.NotifyDecision(r.Context(), familyID, parentID, actingParentName, child.FirstName, "chore", "rejected", 0, choreName, reason, bankName)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"instance": InstanceResponse{
			ID:              updated.ID,
			ChoreID:         updated.ChoreID,
			ChildID:         updated.ChildID,
			RewardCents:     updated.RewardCents,
			Status:          updated.Status,
			PeriodStart:     updated.PeriodStart,
			PeriodEnd:       updated.PeriodEnd,
			CompletedAt:     updated.CompletedAt,
			ReviewedAt:      updated.ReviewedAt,
			RejectionReason: updated.RejectionReason,
			CreatedAt:       updated.CreatedAt,
		},
	})
}

// HandleActivate handles PATCH /api/chores/{id}/activate
func (h *Handler) HandleActivate(w http.ResponseWriter, r *http.Request) {
	h.setChoreActive(w, r, true)
}

// HandleDeactivate handles PATCH /api/chores/{id}/deactivate
func (h *Handler) HandleDeactivate(w http.ResponseWriter, r *http.Request) {
	h.setChoreActive(w, r, false)
}

func (h *Handler) setChoreActive(w http.ResponseWriter, r *http.Request, active bool) {
	if middleware.GetUserType(r) != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Only parents can manage chores."})
		return
	}

	choreID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "Invalid chore ID."})
		return
	}

	existingChore, err := h.choreRepo.GetByID(choreID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error"})
		return
	}
	if existingChore == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Chore not found."})
		return
	}

	familyID := middleware.GetFamilyID(r)
	if existingChore.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Chore does not belong to your family."})
		return
	}

	if err := h.choreRepo.SetActive(choreID, active); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to update chore."})
		return
	}

	updated, _ := h.choreRepo.GetByID(choreID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"chore": ChoreResponse{
			ID:          updated.ID,
			FamilyID:    updated.FamilyID,
			Name:        updated.Name,
			Description: updated.Description,
			RewardCents: updated.RewardCents,
			Recurrence:  string(updated.Recurrence),
			DayOfWeek:   updated.DayOfWeek,
			DayOfMonth:  updated.DayOfMonth,
			IsActive:    updated.IsActive,
			CreatedAt:   updated.CreatedAt,
			UpdatedAt:   updated.UpdatedAt,
		},
	})
}

// UpdateChoreRequest represents the request body for updating a chore.
type UpdateChoreRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	RewardCents *int    `json:"reward_cents,omitempty"`
	Recurrence  *string `json:"recurrence,omitempty"`
	DayOfWeek   *int    `json:"day_of_week,omitempty"`
	DayOfMonth  *int    `json:"day_of_month,omitempty"`
}

// HandleUpdateChore handles PUT /api/chores/{id}
func (h *Handler) HandleUpdateChore(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserType(r) != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Only parents can update chores."})
		return
	}

	choreID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "Invalid chore ID."})
		return
	}

	existingChore, err := h.choreRepo.GetByID(choreID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error"})
		return
	}
	if existingChore == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Chore not found."})
		return
	}

	familyID := middleware.GetFamilyID(r)
	if existingChore.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Chore does not belong to your family."})
		return
	}

	var req UpdateChoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "Invalid request body."})
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) == 0 || len(name) > MaxNameLength {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "validation_error", Message: "Name is required and must be 100 characters or less."})
			return
		}
		existingChore.Name = name
	}

	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		if len(desc) > MaxDescriptionLength {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "validation_error", Message: "Description must be 500 characters or less."})
			return
		}
		if desc == "" {
			existingChore.Description = nil
		} else {
			existingChore.Description = &desc
		}
	}

	if req.RewardCents != nil {
		if *req.RewardCents < 0 || *req.RewardCents > MaxRewardCents {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "validation_error", Message: "Reward must be between 0 and $999,999.99."})
			return
		}
		existingChore.RewardCents = *req.RewardCents
	}

	if req.Recurrence != nil {
		recurrence := models.ChoreRecurrence(*req.Recurrence)
		switch recurrence {
		case models.ChoreRecurrenceOneTime, models.ChoreRecurrenceDaily, models.ChoreRecurrenceWeekly, models.ChoreRecurrenceMonthly:
			existingChore.Recurrence = recurrence
		default:
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "validation_error", Message: "Invalid recurrence type."})
			return
		}
	}

	if req.DayOfWeek != nil {
		existingChore.DayOfWeek = req.DayOfWeek
	}
	if req.DayOfMonth != nil {
		existingChore.DayOfMonth = req.DayOfMonth
	}

	updated, err := h.choreRepo.Update(existingChore)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to update chore."})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"chore": ChoreResponse{
			ID:          updated.ID,
			FamilyID:    updated.FamilyID,
			Name:        updated.Name,
			Description: updated.Description,
			RewardCents: updated.RewardCents,
			Recurrence:  string(updated.Recurrence),
			DayOfWeek:   updated.DayOfWeek,
			DayOfMonth:  updated.DayOfMonth,
			IsActive:    updated.IsActive,
			CreatedAt:   updated.CreatedAt,
			UpdatedAt:   updated.UpdatedAt,
		},
	})
}

// HandleDeleteChore handles DELETE /api/chores/{id}
func (h *Handler) HandleDeleteChore(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserType(r) != "parent" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Only parents can delete chores."})
		return
	}

	choreID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "Invalid chore ID."})
		return
	}

	existingChore, err := h.choreRepo.GetByID(choreID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error"})
		return
	}
	if existingChore == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Chore not found."})
		return
	}

	familyID := middleware.GetFamilyID(r)
	if existingChore.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Chore does not belong to your family."})
		return
	}

	// Delete all instances for this chore before deleting the chore itself
	if err := h.choreInstanceRepo.DeleteByChoreID(choreID); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to delete chore instances."})
		return
	}

	if err := h.choreRepo.Delete(choreID); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to delete chore."})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleChildEarnings handles GET /api/child/chores/earnings
func (h *Handler) HandleChildEarnings(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserType(r) != "child" {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "Only children can view earnings."})
		return
	}

	childID := middleware.GetUserID(r)

	totalCents, completedCount, recent, err := h.choreInstanceRepo.GetEarnings(childID, 10)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get earnings."})
		return
	}

	recentItems := make([]map[string]interface{}, len(recent))
	for i, r := range recent {
		recentItems[i] = map[string]interface{}{
			"chore_name":  r.ChoreName,
			"reward_cents": r.RewardCents,
			"approved_at": r.ApprovedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_earned_cents": totalCents,
		"chores_completed":  completedCount,
		"recent":            recentItems,
	})
}
