package settings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/repositories"
)

var validBankName = regexp.MustCompile(`^[\p{L}\p{N} '\-]+$`)

type Handlers struct {
	familyRepo *repositories.FamilyRepo
	parentRepo *repositories.ParentRepo
}

func NewHandlers(familyRepo *repositories.FamilyRepo, parentRepo *repositories.ParentRepo) *Handlers {
	return &Handlers{familyRepo: familyRepo, parentRepo: parentRepo}
}

func (h *Handlers) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	if familyID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No family associated"})
		return
	}

	tz, err := h.familyRepo.GetTimezone(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	bankName, err := h.familyRepo.GetBankName(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	parentID := middleware.GetUserID(r)
	parent, err := h.parentRepo.GetByID(parentID)
	var notifications map[string]bool
	if err == nil && parent != nil {
		notifications = map[string]bool{
			"notify_withdrawal_requests": parent.NotifyWithdrawalRequests,
			"notify_chore_completions":   parent.NotifyChoreCompletions,
			"notify_decisions":           parent.NotifyDecisions,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"timezone":      tz,
		"bank_name":     bankName,
		"notifications": notifications,
	})
}

func (h *Handlers) HandleUpdateTimezone(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	if familyID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No family associated"})
		return
	}

	var req struct {
		Timezone string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Timezone == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid timezone",
			"message": "Timezone is required",
		})
		return
	}

	// Validate against IANA timezone database
	if _, err := time.LoadLocation(req.Timezone); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid timezone",
			"message": fmt.Sprintf("%q is not a valid IANA timezone identifier", req.Timezone),
		})
		return
	}

	if err := h.familyRepo.UpdateTimezone(familyID, req.Timezone); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":  "Timezone updated",
		"timezone": req.Timezone,
	})
}

func (h *Handlers) HandleUpdateBankName(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	if familyID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No family associated"})
		return
	}

	var req struct {
		BankName string `json:"bank_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	req.BankName = strings.TrimSpace(req.BankName)

	if req.BankName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "bad_request",
			"message": "Bank name is required",
		})
		return
	}

	if len([]rune(req.BankName)) > 12 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "bad_request",
			"message": "Bank name must be 12 characters or fewer",
		})
		return
	}

	if !validBankName.MatchString(req.BankName) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "bad_request",
			"message": "Bank name contains invalid characters",
		})
		return
	}

	if err := h.familyRepo.UpdateBankName(familyID, req.BankName); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Bank name updated",
		"bank_name": req.BankName,
	})
}

func (h *Handlers) HandleGetNotificationPrefs(w http.ResponseWriter, r *http.Request) {
	parentID := middleware.GetUserID(r)

	parent, err := h.parentRepo.GetByID(parentID)
	if err != nil || parent == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"notify_withdrawal_requests": parent.NotifyWithdrawalRequests,
		"notify_chore_completions":   parent.NotifyChoreCompletions,
		"notify_decisions":           parent.NotifyDecisions,
	})
}

func (h *Handlers) HandleUpdateNotificationPrefs(w http.ResponseWriter, r *http.Request) {
	parentID := middleware.GetUserID(r)

	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	validFields := map[string]bool{
		"notify_withdrawal_requests": true,
		"notify_chore_completions":   true,
		"notify_decisions":           true,
	}

	prefs := make(map[string]bool)
	for key, val := range raw {
		if !validFields[key] {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "invalid_request",
				"message": "notification preferences must be boolean values",
			})
			return
		}
		boolVal, ok := val.(bool)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "invalid_request",
				"message": "notification preferences must be boolean values",
			})
			return
		}
		prefs[key] = boolVal
	}

	if len(prefs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "invalid_request",
			"message": "notification preferences must be boolean values",
		})
		return
	}

	if err := h.parentRepo.UpdateNotificationPrefs(parentID, prefs); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	// Fetch updated parent
	parent, err := h.parentRepo.GetByID(parentID)
	if err != nil || parent == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":                    "Notification preferences updated",
		"notify_withdrawal_requests": parent.NotifyWithdrawalRequests,
		"notify_chore_completions":   parent.NotifyChoreCompletions,
		"notify_decisions":           parent.NotifyDecisions,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
